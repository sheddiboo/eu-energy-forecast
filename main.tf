terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "eu-west-1" 
}

resource "aws_glue_catalog_database" "energy_db" {
  name        = "eu_energy_db"
  description = "Database for EU Energy Forecaster Athena tables"
}

# --- Results Bucket (Already Managed) ---
resource "aws_s3_bucket" "athena_results" {
  bucket = "eu-energy-athena-results-sj-ireland"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "athena_results_access" {
  bucket                  = aws_s3_bucket.athena_results.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# --- Raw Data Bucket (Bringing this under Terraform control) ---
resource "aws_s3_bucket" "bronze_raw" {
  bucket = "eu-energy-raw-ireland-sj"
  force_destroy = true
}

resource "aws_athena_workgroup" "energy_workgroup" {
  name = "energy_analytics_wg"
  
  configuration {
    enforce_workgroup_configuration    = false
    publish_cloudwatch_metrics_enabled = true

    result_configuration {
      output_location = "s3://${aws_s3_bucket.athena_results.bucket}/query-results/"
    }
  }
}

output "athena_database_name" {
  value = aws_glue_catalog_database.energy_db.name
}

output "athena_workgroup_name" {
  value = aws_athena_workgroup.energy_workgroup.name
}

output "athena_results_s3_bucket" {
  value = aws_s3_bucket.athena_results.bucket
}

output "raw_data_s3_bucket" {
  value = aws_s3_bucket.bronze_raw.bucket
}