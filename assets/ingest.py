"""@bruin
name: fetch_raw_energy_data
type: python
image: python:3.11
@bruin"""

import subprocess
import os

project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))

print("Running Golang Ingestion...")
subprocess.run("go run main.go parser.go", shell=True, cwd=os.path.join(project_root, "assets", "ingestion", "entsoe_extractor"), check=True)
subprocess.run("go run main.go", shell=True, cwd=os.path.join(project_root, "assets", "ingestion", "weather_extractor"), check=True)
subprocess.run("go run aws_uploader.go", shell=True, cwd=os.path.join(project_root, "assets", "ingestion"), check=True)