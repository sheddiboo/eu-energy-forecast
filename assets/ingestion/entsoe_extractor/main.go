package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("⚡ Starting EU Energy API Test in Golang...")

	// 1. Load the .env file from the root directory (3 levels up)
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Fatal("❌ Error loading .env file. Are you sure it's 3 levels up?")
	}

	// 2. Fetch the token securely from the environment variable
	token := os.Getenv("ENTSOE_API_KEY")
	if token == "" {
		log.Fatal("❌ ENTSOE_API_KEY is empty or not found in .env")
	}

	// The API endpoint for Germany Day-Ahead Prices
	url := "https://web-api.tp.entsoe.eu/api?securityToken=" + token + "&documentType=A44&in_Domain=10Y1001A1001A82H&out_Domain=10Y1001A1001A82H&periodStart=202401010000&periodEnd=202401012300"

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("❌ Request failed:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("❌ Failed to read body:", err)
		return
	}

	// Print the results
	fmt.Printf("✅ Status Code: %d\n\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		fmt.Println("🎉 API Key is Valid! Here is a snippet of the raw XML data:")
		// Print the first 400 characters of the XML
		fmt.Println(string(body)[:400])
	} else {
		fmt.Println("Something went wrong. Full response:")
		fmt.Println(string(body))
	}
}