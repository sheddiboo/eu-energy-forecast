#!/bin/bash
# @bruin
# name: fetch_raw_energy_data
# type: bash
# @bruin

echo "Starting Golang Ingestion..."

# 1. Fetch ENTSO-E Grid Data
cd assets/ingestion/entsoe_extractor && go run main.go parser.go

# 2. Fetch Open-Meteo Weather Data
cd ../weather_extractor && go run main.go

# 3. Push to AWS S3
cd .. && go run aws_uploader.go