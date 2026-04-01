package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"os"
)

// Define the XML Blueprint (We only grab the tags we care about!)
type MarketDocument struct {
	XMLName    xml.Name     `xml:"Publication_MarketDocument"`
	TimeSeries []TimeSeries `xml:"TimeSeries"`
}

type TimeSeries struct {
	Period Period `xml:"Period"`
}

type Period struct {
	Points []Point `xml:"Point"`
}

type Point struct {
	Position string `xml:"position"`
	Price    string `xml:"price.amount"`
}

// ParseAndSaveCSV takes the raw XML bytes, flattens them, and writes a clean CSV
func ParseAndSaveCSV(xmlData []byte, filename string) error {
	var doc MarketDocument

	// Unmarshal decodes the ugly XML byte array into our clean Go Structs
	err := xml.Unmarshal(xmlData, &doc)
	if err != nil {
		return fmt.Errorf("failed to parse XML: %v", err)
	}

	// Create the output CSV file in your landing zone
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the header row
	writer.Write([]string{"Position", "Price_EUR_MWh"})

	// Loop through the nested data and flatten it into standard rows
	for _, ts := range doc.TimeSeries {
		for _, p := range ts.Period.Points {
			record := []string{p.Position, p.Price}
			writer.Write(record)
		}
	}

	return nil
}
