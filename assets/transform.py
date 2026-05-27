"""@bruin
name: transform_silver_gold
type: python
image: python:3.11
depends:
  - fetch_raw_energy_data
@bruin"""

import subprocess
import os

project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))

print("Executing Athena Medallion Transformations...")
subprocess.run("uv run python assets/transformation/run_athena.py", shell=True, cwd=project_root, check=True)
