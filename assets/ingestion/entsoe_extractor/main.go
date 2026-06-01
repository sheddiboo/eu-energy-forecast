// entsoe/main.go
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// ApiRequest stores parameters for specific geographic regions and data types.
type ApiRequest struct {
	Country      string
	Domain       string
	DocumentType string
	ProcessType  string
	FeatureName  string
	Start        string
	End          string
}

// A global HTTP client configured to handle slow ENTSO-E responses.
var client = &http.Client{
	Timeout: 390 * time.Second,
}

// fetchENTSOE orchestrates the network request for a specific time chunk.
func fetchENTSOE(req ApiRequest, token string, wg *sync.WaitGroup, limiter chan struct{}) {
	defer wg.Done()
	defer func() { <-limiter }()

	outputPath := fmt.Sprintf("../../../data/raw/%s_%s.csv", req.Country, req.FeatureName)

	url := fmt.Sprintf("https://web-api.tp.entsoe.eu/api?securityToken=%s&documentType=%s&periodStart=%s&periodEnd=%s",
		token, req.DocumentType, req.Start, req.End)

	if req.DocumentType == "A44" {
		url += fmt.Sprintf("&in_Domain=%s&out_Domain=%s", req.Domain, req.Domain)
	} else if req.DocumentType == "A65" {
		url += fmt.Sprintf("&outBiddingZone_Domain=%s", req.Domain)
	} else {
		url += fmt.Sprintf("&in_Domain=%s", req.Domain)
	}

	if req.ProcessType != "" {
		url += "&processType=" + req.ProcessType
	}

	fmt.Printf("Requesting %s for %s (Period: %s - %s)...\n", req.FeatureName, req.Country, req.Start, req.End)

	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Get(url)

		if err != nil {
			if attempt < maxRetries {
				fmt.Printf("Network drop for %s %s (Attempt %d/%d). Retrying in 10s...\n", req.Country, req.FeatureName, attempt, maxRetries)
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Printf("Fatal Network Error for %s %s after %d attempts: %v\n", req.Country, req.FeatureName, maxRetries, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, _ := io.ReadAll(resp.Body)
			err = ParseAndSaveCSV(body, outputPath)
			if err != nil {
				fmt.Printf("Parsing failed for %s %s: %v\n", req.Country, req.FeatureName, err)
			} else {
				fmt.Printf("Success. Appended to %s_%s.csv\n", req.Country, req.FeatureName)
			}
		} else {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("HTTP %d Error for %s %s: %s\n", resp.StatusCode, req.Country, req.FeatureName, string(body))
		}
		break
	}

	// CRITICAL RATE LIMIT PROTECTOR: Force a 5-second pause between EVERY API hit.
	time.Sleep(5 * time.Second)
}

func main() {
	start := time.Now()
	fmt.Println("Starting Daily Pan-European Ingestion...")

	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Println("No .env file found. Proceeding with system environment variables.") // <--- THE ONLY CHANGE
	}

	token := os.Getenv("ENTSOE_API_KEY")

	countries := []struct{ Name, Code string }{
		{"Austria", "10YAT-APG------L"}, {"Belgium", "10YBE----------2"},
		{"Bulgaria", "10YCA-BULGARIA-R"}, {"Croatia", "10YHR-HEP------M"},
		{"CzechRepublic", "10YCZ-CEPS-----N"}, {"Denmark", "10YDK-1--------W"},
		{"Estonia", "10Y1001A1001A39I"}, {"Finland", "10YFI-1--------U"},
		{"France", "10YFR-RTE------C"}, {"Germany", "10Y1001A1001A82H"},
		{"Greece", "10YGR-HTSO-----Y"}, {"Hungary", "10YHU-MAVIR----U"},
		{"Ireland", "10Y1001A1001A597"}, {"Italy", "10Y1001A1001A73I"},
		{"Latvia", "10YLV-1001A00074"}, {"Lithuania", "10YLT-1001A0008Q"},
		{"Netherlands", "10YNL----------L"}, {"Norway", "10YNO-1--------2"},
		{"Poland", "10YPL-AREA-----S"}, {"Portugal", "10YPT-REN------W"},
		{"Romania", "10YRO-TEL------P"}, {"Slovakia", "10YSK-SEPS-----K"},
		{"Slovenia", "10YSI-ELES-----O"}, {"Spain", "10YES-REE------0"},
		{"Sweden", "10Y1001A1001A46L"}, {"Switzerland", "10YCH-SWISSGRIDZ"},
	}

	features := []struct{ Doc, Proc, Name string }{
		{"A44", "", "Prices"},
		{"A75", "A16", "Actual_Generation"},
		{"A65", "A16", "Actual_Load"},
		{"A65", "A01", "Forecast_Load"},
		{"A69", "A01", "Forecast_WindSolar"},
	}

	// Dynamic time window: Yesterday to Tomorrow
	now := time.Now().UTC()
	startStr := now.AddDate(0, 0, -1).Format("200601020000")
	endStr := now.AddDate(0, 0, 1).Format("200601022300")

	timeChunks := []struct{ Start, End string }{
		{startStr, endStr},
	}

	var wg sync.WaitGroup
	limiter := make(chan struct{}, 1) // Enforces single-file processing to save the API

	for _, chunk := range timeChunks {
		fmt.Printf("\n--- Processing Time Chunk: %s to %s ---\n", chunk.Start, chunk.End)
		for _, c := range countries {
			for _, f := range features {
				wg.Add(1)
				limiter <- struct{}{}
				go fetchENTSOE(ApiRequest{c.Name, c.Code, f.Doc, f.Proc, f.Name, chunk.Start, chunk.End}, token, &wg, limiter)
			}
		}
	}

	wg.Wait()
	fmt.Printf("\nFull safe extraction completed in %s\n", time.Since(start))
}
