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

// ParseAndSaveCSV transforms raw XML bytes into a structured CSV, appending if the file exists.
func ParseAndSaveCSV(xmlData []byte, filename string) error {
	var doc MarketDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return fmt.Errorf("failed to unmarshal XML: %v", err)
	}

	if len(doc.TimeSeries) == 0 {
		return fmt.Errorf("no data available from ENTSO-E for this feature")
	}

	// SMART APPEND LOGIC: Check if file exists to determine if we need to write headers
	fileExists := false
	if info, err := os.Stat(filename); err == nil && info.Size() > 0 {
		fileExists = true
	}

	// Open file in append mode. Create it if it doesn't exist.
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	f := strings.ToLower(filename)
	isPrice := strings.Contains(f, "prices")
	isGeneration := strings.Contains(f, "generation") || strings.Contains(f, "windsolar")

	// Only write headers if this is a brand new file (i.e., the 2015 chunk)
	if !fileExists {
		header := []string{"Timestamp"}
		if isPrice {
			header = append(header, "Price_EUR_MWh")
		} else if isGeneration {
			header = append(header, "Gen_MW", "FuelType")
		} else {
			header = append(header, "Value_MW")
		}
		writer.Write(header)
	}

	for _, ts := range doc.TimeSeries {
		layout := "2006-01-02T15:04Z"
		startTime, err := time.Parse(layout, ts.Period.TimeInterval.Start)
		if err != nil {
			startTime, err = time.Parse(time.RFC3339, ts.Period.TimeInterval.Start)
			if err != nil {
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
