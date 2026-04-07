package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// WeatherResponse maps all 10 critical energy-weather metrics from the Open-Meteo JSON output.
type WeatherResponse struct {
	Hourly struct {
		Time             []string  `json:"time"`
		Temperature      []float64 `json:"temperature_2m"`
		ApparentTemp     []float64 `json:"apparent_temperature"`
		RelativeHumidity []float64 `json:"relative_humidity_2m"`
		Precipitation    []float64 `json:"precipitation"`
		Snowfall         []float64 `json:"snowfall"`
		CloudCover       []float64 `json:"cloud_cover"`
		ShortwaveRad     []float64 `json:"shortwave_radiation"`
		WindSpeed10m     []float64 `json:"wind_speed_10m"`
		WindSpeed100m    []float64 `json:"wind_speed_100m"`
		WindGusts        []float64 `json:"wind_gusts_10m"`
	} `json:"hourly"`
}

// fetchWeather executes the API request for a specific coordinate and serializes the response to a CSV file.
func fetchWeather(city string, lat, lon float64, wg *sync.WaitGroup, limiter chan struct{}) {
	// The function signals the WaitGroup upon completion.
	defer wg.Done()

	// The routine acquires a slot in the semaphore channel to restrict concurrent network traffic.
	limiter <- struct{}{}
	// The routine releases the semaphore slot when the function exits.
	defer func() { <-limiter }()

	// The script constructs the URL, requesting the complete 10-feature meteorological suite.
	url := fmt.Sprintf("https://archive-api.open-meteo.com/v1/archive?latitude=%f&longitude=%f&start_date=2026-01-01&end_date=2026-04-01&hourly=temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,snowfall,cloud_cover,shortwave_radiation,wind_speed_10m,wind_speed_100m,wind_gusts_10m&timezone=UTC", lat, lon)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", city, err)
		return
	}
	defer resp.Body.Close()

	// The application decodes the JSON payload directly into the defined Go structure.
	var data WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Printf("Error decoding JSON for %s: %v\n", city, err)
		return
	}

	// The system creates a local CSV file within the raw data landing zone.
	filePath := fmt.Sprintf("../../../data/raw/%s_weather.csv", city)
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file for %s: %v\n", city, err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// The application writes the expanded CSV header containing all 10 features.
	writer.Write([]string{
		"Timestamp", "Temperature_2m", "Apparent_Temp", "Rel_Humidity_2m", "Precipitation",
		"Snowfall", "Cloud_Cover", "Shortwave_Rad", "WindSpeed_10m",
		"WindSpeed_100m", "WindGusts",
	})

	// The script iterates through the hourly data arrays and writes each consolidated row.
	for i := range data.Hourly.Time {
		writer.Write([]string{
			data.Hourly.Time[i],
			fmt.Sprintf("%.2f", data.Hourly.Temperature[i]),
			fmt.Sprintf("%.2f", data.Hourly.ApparentTemp[i]),
			fmt.Sprintf("%.2f", data.Hourly.RelativeHumidity[i]),
			fmt.Sprintf("%.2f", data.Hourly.Precipitation[i]),
			fmt.Sprintf("%.2f", data.Hourly.Snowfall[i]),
			fmt.Sprintf("%.2f", data.Hourly.CloudCover[i]),
			fmt.Sprintf("%.2f", data.Hourly.ShortwaveRad[i]),
			fmt.Sprintf("%.2f", data.Hourly.WindSpeed10m[i]),
			fmt.Sprintf("%.2f", data.Hourly.WindSpeed100m[i]),
			fmt.Sprintf("%.2f", data.Hourly.WindGusts[i]),
		})
	}
	fmt.Printf("Saved complete meteorological data for %s to %s\n", city, filePath)
}

func main() {
	var wg sync.WaitGroup

	// The program defines a mapping of the primary energy hubs for all 26 EU countries.
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

	// The main thread initializes a semaphore channel to restrict the pipeline to 5 concurrent requests.
	limiter := make(chan struct{}, 5)

	// The system dispatches a distinct Goroutine for each geographical coordinate.
	for _, c := range cities {
		wg.Add(1)
		go fetchWeather(c.name, c.lat, c.lon, &wg, limiter)
	}

	// The main thread blocks until all Goroutines have successfully completed their I/O operations.
	wg.Wait()
	fmt.Println("Enterprise weather extraction complete.")
}
