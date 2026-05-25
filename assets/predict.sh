#!/bin/bash
# @bruin
# name: predict_day_ahead_price
# type: bash
# depends:
#   - fetch_raw_energy_data
# @bruin

echo "Starting Python Inference..."
uv run python assets/intelligence/predict_price.py
