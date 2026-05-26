#!/bin/bash
# @bruin
# name: transform_silver_gold
# type: bash
# depends:
#   - fetch_raw_energy_data
# @bruin

echo "Step 1: Emptying old Silver and Gold S3 directories..."
aws s3 rm s3://eu-energy-raw-ireland-sj/silver/master_data/ --recursive
aws s3 rm s3://eu-energy-raw-ireland-sj/gold/ml_features/ --recursive

echo "Step 2: Executing Athena SQL Transformations..."
uv pip install boto3
uv run python assets/transformation/run_athena.py