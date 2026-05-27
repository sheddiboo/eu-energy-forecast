import boto3
import time
import os

REGION = 'eu-west-1'
DATABASE = 'eu_energy_db'
# Athena requires a location to save query logs. We will store them in a temporary folder in your raw bucket.
OUTPUT_LOCATION = 's3://eu-energy-raw-ireland-sj/athena_query_results/'

athena = boto3.client('athena', region_name=REGION)

def execute_athena_query(query, description):
    print(f"Starting: {description}...")
    response = athena.start_query_execution(
        QueryString=query,
        QueryExecutionContext={'Database': DATABASE},
        ResultConfiguration={'OutputLocation': OUTPUT_LOCATION}
    )
    query_id = response['QueryExecutionId']
    
    while True:
        status = athena.get_query_execution(QueryExecutionId=query_id)
        state = status['QueryExecution']['Status']['State']
        
        if state == 'SUCCEEDED':
            print(f"✓ Success: {description}")
            break
        elif state in ['FAILED', 'CANCELLED']:
            reason = status['QueryExecution']['Status'].get('StateChangeReason', 'Unknown reason')
            print(f"✗ Failed: {description}\nReason: {reason}")
            raise Exception(f"Athena query failed: {reason}")
            
        time.sleep(2)

if __name__ == "__main__":
    # 1. Manually empty the S3 directories first
    s3 = boto3.resource('s3')
    bucket = s3.Bucket('eu-energy-raw-ireland-sj')
    
    print("Emptying S3 directories for clean build...")
    bucket.objects.filter(Prefix='silver/master_data/').delete()
    bucket.objects.filter(Prefix='gold/ml_features/').delete()

    # 2. Drop existing tables
    execute_athena_query("DROP TABLE IF EXISTS silver_master_energy_data;", "Drop old Silver table")
    execute_athena_query("DROP TABLE IF EXISTS gold_ml_features;", "Drop old Gold table")
    
    # 3. Build layers
    base_dir = os.path.dirname(__file__)
    
    with open(os.path.join(base_dir, 'silver_layer.sql'), 'r') as f:
        silver_sql = f.read()
        execute_athena_query(silver_sql, "Build Silver Layer (CTAS)")
        
    with open(os.path.join(base_dir, 'gold_layer.sql'), 'r') as f:
        gold_sql = f.read()
        execute_athena_query(gold_sql, "Build Gold Layer (CTAS)")
        
    print("All Athena transformations completed successfully!")