#!/bin/bash
# @bruin
# name: fetch_raw_energy_data
# type: bash
# @bruin

echo "Starting Golang Ingestion..."
# We use absolute paths from the project root to ensure it always finds the file
cd assets/ingestion && go run main.go
