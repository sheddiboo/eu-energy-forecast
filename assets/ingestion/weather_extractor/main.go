package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// WeatherResponse maps the exact metrics needed for the Athena Silver Layer
type WeatherResponse struct {
	Hourly struct {
		Time         []string  `json:"time"`
		Temperature  []float64 `json:"temperature_2m"`
		WindSpeed10m []float64 `json:"wind_speed_10m"`
	} `json:"hourly"`
}

func fetchWeather(city string, lat, lon float64, wg *sync.WaitGroup, limiter chan struct{}) {
	defer wg.Done()
	limiter <- struct{}{}
	defer func() { <-limiter }()

	// Fetches the entire 11-year history in a single API call for the specific coordinates
	url := fmt.Sprintf("https://archive-api.open-meteo.com/v1/archive?latitude=%f&longitude=%f&start_date=2015-01-01&end_date=2026-05-19&hourly=temperature_2m,wind_speed_10m&timezone=UTC", lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", city, err)
		return
	}
	defer resp.Body.Close()

	var data WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error decoding JSON for %s: %v\n", city, err)
		return
	}

	filePath := fmt.Sprintf("../../../data/raw/%s_weather.csv", city)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file for %s: %v\n", city, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Writes exactly what Athena expects: Timestamp, Temperature, WindSpeed
	writer.Write([]string{"Timestamp", "Temperature", "WindSpeed"})

	for i := range data.Hourly.Time {
		writer.Write([]string{
			// Open-Meteo returns '2015-01-01T00:00'. We replace 'T' with a space to match ENTSO-E formatting
			string(data.Hourly.Time[i][:10]) + " " + string(data.Hourly.Time[i][11:]),
			fmt.Sprintf("%.2f", data.Hourly.Temperature[i]),
			fmt.Sprintf("%.2f", data.Hourly.WindSpeed10m[i]),
		})
	}
	fmt.Printf("Saved 11-year weather data for %s\n", city)
}

func main() {
	var wg sync.WaitGroup

	cities := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"Austria", 48.21, 16.37}, {"Belgium", 50.85, 4.35},
		{"Bulgaria", 42.70, 23.32}, {"Croatia", 45.81, 15.98},
		{"CzechRepublic", 50.08, 14.44}, {"Denmark", 55.68, 12.57},
		{"Estonia", 59.44, 24.75}, {"Finland", 60.17, 24.94},
		{"France", 48.85, 2.35}, {"Germany", 50.11, 8.68},
		{"Greece", 37.98, 23.73}, {"Hungary", 47.50, 19.04},
		{"Ireland", 53.35, -6.26}, {"Italy", 41.90, 12.49},
		{"Latvia", 56.95, 24.11}, {"Lithuania", 54.69, 25.28},
		{"Netherlands", 52.37, 4.89}, {"Norway", 59.91, 10.75},
		{"Poland", 52.23, 21.01}, {"Portugal", 38.72, -9.14},
		{"Romania", 44.43, 26.10}, {"Slovakia", 48.15, 17.11},
		{"Slovenia", 46.05, 14.51}, {"Spain", 40.41, -3.70},
		{"Sweden", 59.33, 18.06}, {"Switzerland", 47.37, 8.54},
	}

	limiter := make(chan struct{}, 5)

	fmt.Println("Starting 11-Year Open-Meteo Weather Extraction...")
	for _, c := range cities {
		wg.Add(1)
		go fetchWeather(c.name, c.lat, c.lon, &wg, limiter)
	}

	wg.Wait()
	fmt.Println("Weather extraction complete.")
}
