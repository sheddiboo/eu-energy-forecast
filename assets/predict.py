"""@bruin
name: predict_day_ahead_price
type: python
image: python:3.11
depends:
  - transform_silver_gold
@bruin"""

import subprocess
import os

project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))

print("Starting Live AI Prediction...")
subprocess.run("uv run python assets/intelligence/predict_price.py", shell=True, cwd=project_root, check=True)
