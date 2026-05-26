#!/bin/bash
# @bruin
# name: predict_day_ahead_price
# type: bash
# depends:
#   - transform_silver_gold
# @bruin

echo "Starting Python Inference..."
uv run python assets/intelligence/predict_price.py