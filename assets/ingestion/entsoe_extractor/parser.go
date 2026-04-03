package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// MarketDocument represents the root structure of the ENTSO-E XML response.
type MarketDocument struct {
	TimeSeries []TimeSeries `xml:"TimeSeries"`
}

type TimeSeries struct {
	BusinessType string `xml:"businessType"`
	MktPSRType   struct {
		PsrType string `xml:"psrType"`
	} `xml:"MktPSRType"`
	Period Period `xml:"Period"`
}

type Period struct {
	TimeInterval struct {
		Start string `xml:"start"`
	} `xml:"timeInterval"`
	Resolution string  `xml:"resolution"`
	Points     []Point `xml:"Point"`
}

type Point struct {
	Position string `xml:"position"`
	Price    string `xml:"price.amount"`
	Quantity string `xml:"quantity"`
}

// ParseAndSaveCSV transforms raw XML bytes into a structured CSV with calculated timestamps.
func ParseAndSaveCSV(xmlData []byte, filename string) error {
	var doc MarketDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %v", err)
	}

	if len(doc.TimeSeries) == 0 {
		return fmt.Errorf("no data available from ENTSO-E for this feature")
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	f := strings.ToLower(filename)
	isPrice := strings.Contains(f, "prices")
	isGeneration := strings.Contains(f, "generation") || strings.Contains(f, "windsolar")

	header := []string{"Timestamp"}
	if isPrice {
		header = append(header, "Price_EUR_MWh")
	} else if isGeneration {
		header = append(header, "Gen_MW", "FuelType")
	} else {
		header = append(header, "Value_MW")
	}
	writer.Write(header)

	for _, ts := range doc.TimeSeries {
		// A custom layout is defined to match the ENTSO-E standard, which omits seconds.
		layout := "2006-01-02T15:04Z"
		startTime, err := time.Parse(layout, ts.Period.TimeInterval.Start)
		if err != nil {
			// A fallback to RFC3339 is provided to handle potential API format variations.
			startTime, err = time.Parse(time.RFC3339, ts.Period.TimeInterval.Start)
			if err != nil {
				// Errors are surfaced to standard output rather than failing silently.
				fmt.Printf("Timestamp parsing error for %s: %v\n", filename, err)
				continue
			}
		}

		resolution := time.Hour
		if ts.Period.Resolution == "PT15M" {
			resolution = 15 * time.Minute
		} else if ts.Period.Resolution == "PT30M" {
			resolution = 30 * time.Minute
		}

		for _, p := range ts.Period.Points {
			pos, _ := strconv.Atoi(p.Position)

			timestamp := startTime.Add(time.Duration(pos-1) * resolution).Format("2006-01-02 15:04:05")

			val := p.Price
			if val == "" {
				val = p.Quantity
			}

			if val != "" {
				row := []string{timestamp, val}
				if isGeneration {
					// Extract the fuel type. Fallback to businessType if it is missing.
					fuelType := ts.MktPSRType.PsrType
					if fuelType == "" {
						fuelType = ts.BusinessType
					}
					row = append(row, fuelType)
				}
				writer.Write(row)
			}
		}
	}
	return nil
}
