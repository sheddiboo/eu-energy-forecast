package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

const (
	bucketName = "eu-energy-raw-ireland-sj"
	region     = "eu-west-1"
	dataDir    = "../../data/raw" // Navigates up 2 levels to the root data folder
	s3Prefix   = "bronze/"        // The destination folder structure within S3
)

// uploadFile manages the transfer of a single file to AWS S3, routing it to the appropriate sub-folder.
func uploadFile(ctx context.Context, client *s3.Client, filePath string, wg *sync.WaitGroup, limiter chan struct{}) {
	defer wg.Done()
	defer func() { <-limiter }()

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file %s: %v\n", filePath, err)
		return
	}
	defer file.Close()

	fileName := filepath.Base(filePath)
	var subFolder string

	// Determine the correct S3 sub-folder based on the feature name in the file
	if strings.Contains(fileName, "Prices") {
		subFolder = "prices/"
	} else if strings.Contains(fileName, "Actual_Load") {
		subFolder = "actual_load/"
	} else if strings.Contains(fileName, "Forecast_Load") {
		subFolder = "forecast_load/"
	} else if strings.Contains(fileName, "Actual_Generation") {
		subFolder = "actual_generation/"
	} else if strings.Contains(fileName, "Forecast_WindSolar") {
		subFolder = "forecast_wind_solar/"
	} else if strings.Contains(fileName, "weather") {
		subFolder = "weather/"
	} else {
		subFolder = "misc/"
	}

	s3Key := s3Prefix + subFolder + fileName

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	if err != nil {
		log.Printf("Failed to upload %s: %v\n", fileName, err)
	} else {
		fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", fileName, bucketName, s3Key)
	}
}

func main() {
	start := time.Now()
	fmt.Println("Starting Full S3 Pipeline Upload (All Grid and Weather Files)...")

	// Load environment variables for AWS authentication
	err := godotenv.Load("../../.env") // Navigates up 2 levels to the root .env file
	if err != nil {
		log.Println("No .env file found. Proceeding with system environment variables.")
	}

	// Initialize the AWS Client configuration for the target region
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg)

	fmt.Println("Scanning for local CSV files...")
	var wg sync.WaitGroup

	// Implement a semaphore to limit concurrent uploads
	limiter := make(chan struct{}, 10)

	// Read all files located in the local raw data directory
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("Failed to read directory %s: %v", dataDir, err)
	}

	uploadCount := 0
	for _, file := range files {
		// Upload all CSV files
		if !file.IsDir() && filepath.Ext(file.Name()) == ".csv" {
			uploadCount++
			wg.Add(1)
			limiter <- struct{}{}

			filePath := filepath.Join(dataDir, file.Name())
			go uploadFile(context.TODO(), client, filePath, &wg, limiter)
		}
	}

	wg.Wait()
	fmt.Printf("Transfer complete. %d files safely synced to S3 in %s\n", uploadCount, time.Since(start))
}
