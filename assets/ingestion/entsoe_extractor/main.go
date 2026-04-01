package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting EU Energy API Test in Golang...")

	// Load the .env file from the root directory
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Fatal("Error loading .env file. Are you sure it's 3 levels up?")
	}

	// Fetch the token securely from the environment variable
	token := os.Getenv("ENTSOE_API_KEY")
	if token == "" {
		log.Fatal("ENTSOE_API_KEY is empty or not found in .env")
	}

	// The API endpoint for Germany Day-Ahead Prices
	url := "https://web-api.tp.entsoe.eu/api?securityToken=" + token + "&documentType=A44&in_Domain=10Y1001A1001A82H&out_Domain=10Y1001A1001A82H&periodStart=202401010000&periodEnd=202401012300"

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Failed to read body:", err)
		return
	}

	// Print the results
	fmt.Printf("Status Code: %d\n\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("API Key Valid! Parsing XML and flattening to CSV...")

		// Define the path to your raw data landing zone
		outputPath := "../../../data/raw/germany_prices.csv"

		// Call the function built in parser.go
		err = ParseAndSaveCSV(body, outputPath)
		if err != nil {
			fmt.Println("Parsing Failed:", err)
			return
		}

		fmt.Println("Success! Check the data/raw folder for your clean CSV.")
	} else {
		fmt.Printf("Something went wrong. Status: %d\n", resp.StatusCode)
		fmt.Println(string(body))
	}
}
