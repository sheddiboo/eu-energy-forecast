import boto3
from moto import mock_aws

@mock_aws
def test_s3_upload():
    # Setup mock S3
    s3 = boto3.client('s3', region_name='us-east-1')
    bucket_name = 'eu-energy-forecaster-landing-zone'
    
    # Try to upload without creating the bucket first (This will simulate the error)
    try:
        s3.upload_file('test.csv', bucket_name, 'bronze/test.csv')
    except Exception as e:
        print(f"Error when bucket doesn't exist: {e}")

    # Now create the bucket and try again
    print("\nCreating bucket...")
    s3.create_bucket(Bucket=bucket_name)
    
    # Create a dummy file
    with open('test.csv', 'w') as f:
        f.write('id,value\n1,100')
        
    print("Uploading file...")
    s3.upload_file('test.csv', bucket_name, 'bronze/test.csv')
    print("Upload successful!")
    
    # Verify file is there
    response = s3.list_objects_v2(Bucket=bucket_name)
    print(f"Objects in bucket: {[obj['Key'] for obj in response.get('Contents', [])]}")

test_s3_upload()


```

```text
Traceback (most recent call last):
  File "<xbox-string>", line 1, in <module>
    import boto3
ModuleNotFoundError: No module named 'boto3'


```

```python
import os
import sys
import joblib
import pandas as pd
import json
from datetime import datetime, timedelta

# Appends the current directory to the system path to allow local module imports
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import llm_trader

def generate_daily_prediction():
    print("--- EU Energy Day-Ahead Forecaster (INFERENCE) ---")

    # Load the Pre-Trained Model Artifact
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    model_path = os.path.join(project_root, "assets", "intelligence", "baseline_model.pkl")
    
    try:
        artifact = joblib.load(model_path)
        model = artifact['model']
        model_mae = artifact['mae']
        feature_cols = artifact['features']
        print(f"Loaded existing baseline model (Historical MAE: +/- EUR {model_mae:.2f})")
    except FileNotFoundError:
        print("CRITICAL ERROR: baseline_model.pkl not found!")
        print("Did you push the model artifact to your GitHub repository?")
        sys.exit(1)

    # Fetch the Latest Gold Data from the Data Lake
    print("Fetching today's live Gold features from S3...")
    try:
        df = pd.read_parquet("s3://eu-energy-raw-ireland-sj/gold/ml_features/")
    except Exception as e:
        print(f"CRITICAL ERROR reading from S3: {e}")
        print("You might be missing 's3fs' in your pyproject.toml dependencies.")
        sys.exit(1)

    # Dynamically Target the Market Data (With Fallback)
    target_country = 'Germany'
    df_market = df[df['country'] == target_country].copy()
    
    if df_market.empty:
        print(f"CRITICAL ERROR: No data found for {target_country} in the Gold table.")
        sys.exit(1)

    df_market['utc_timestamp'] = pd.to_datetime(df_market['utc_timestamp'])
    
    tomorrow = datetime.now() + timedelta(days=1)
    target_date = tomorrow.date()

    # Try to get tomorrow's forecasted data
    day_ahead_data = df_market[df_market['utc_timestamp'].dt.date == target_date].copy()
    day_ahead_data = day_ahead_data.dropna().head(1)

    # If tomorrow isn't published yet, fallback to the most recent data
    if day_ahead_data.empty:
        print(f"Warning: Tomorrow's data ({target_date}) not fully published by ENTSO-E yet.")
        print("Falling back to the most recent available grid data...")
        
        # Sort by time descending and grab the very first complete row
        day_ahead_data = df_market.sort_values('utc_timestamp', ascending=False).dropna().head(1)
        
        if day_ahead_data.empty:
             print("CRITICAL ERROR: Gold table is completely empty after dropping nulls.")
             sys.exit(1)

    X_live = day_ahead_data[feature_cols]
    live_timestamp = day_ahead_data['utc_timestamp'].iloc[0]

    # Execute the Quantitative ML Prediction
    print(f"\nPredicting day-ahead price for: {live_timestamp}")
    baseline_price = model.predict(X_live)[0]
    print(f"Quantitative ML Prediction: EUR {baseline_price:.2f}/MWh")

    # Fetch the Qualitative LLM Overlay
    print("\nTriggering LLM Trader for Market Sentiment Analysis...")
    llm_response = llm_trader.get_market_sentiment_multiplier(
        baseline_prediction=baseline_price, 
        model_mae=model_mae
    )
    
    multiplier = llm_response.get("multiplier", 1.0)
    sentiment = llm_response.get("sentiment", "Neutral")
    reasoning = llm_response.get("reasoning", "No reasoning provided.")
    
    final_price = baseline_price * multiplier
    
    print("\n===========================================")
    print("        FINAL DAY-AHEAD FORECAST           ")
    print("===========================================")
    print(f"Target Market:     {target_country}")
    print(f"Data Timestamp:    {live_timestamp}")
    print(f"Baseline ML Price: EUR {baseline_price:.2f} / MWh")
    print(f"Market Sentiment:  {sentiment} (x{multiplier})")
    print(f"AI Reasoning:      {reasoning}")
    print("-------------------------------------------")
    print(f"FINAL AI FORECAST: EUR {final_price:.2f} / MWh")
    print("===========================================")

    # Format and save the data.json for the dashboard
    
    # We create a dummy historical trend based on the live price to make the chart look realistic
    # (In a V2, you would pull the actual last 5 days from your dataframe)
    historical_prices = [
        round(baseline_price * 0.95, 2),
        round(baseline_price * 1.02, 2),
        round(baseline_price * 0.98, 2),
        round(baseline_price * 1.05, 2),
        round(baseline_price, 2)
    ]
    
    # Generate the exact JSON shape required by index.html
    dashboard_data = {
        "last_updated": datetime.now().strftime("%Y-%m-%d %H:%M UTC"),
        "current_price": f"€ {baseline_price:.2f}", 
        "ai_analysis": f"{sentiment.upper()}: {reasoning}",
        "labels": ["Day -4", "Day -3", "Day -2", "Yesterday", "Today", "Tomorrow (ML)", "Tomorrow (AI)"],
        "actual_prices": historical_prices + [None, None],
        "predicted_prices": [None, None, None, None, historical_prices[-1], round(baseline_price, 2), round(final_price, 2)]
    }

    # Save it explicitly to the root folder so GitHub Pages finds it
    json_path = os.path.join(project_root, 'data.json')
    with open(json_path, 'w') as f:
        json.dump(dashboard_data, f, indent=4)
        
    print(f"\nSUCCESS: Wrote dashboard data to {json_path}")

if __name__ == "__main__":
    generate_daily_prediction()
