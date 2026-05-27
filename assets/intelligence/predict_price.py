import os
import sys
import joblib
import pandas as pd
from datetime import datetime, timedelta

# Appends the current directory to the system path to allow local module imports
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import llm_trader

def generate_daily_prediction():
    print("--- EU Energy Day-Ahead Forecaster (INFERENCE) ---")

    # 1. Load the Pre-Trained Model Artifact
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    model_path = os.path.join(project_root, "assets", "intelligence", "baseline_model.pkl")
    
    try:
        artifact = joblib.load(model_path)
        model = artifact['model']
        model_mae = artifact['mae']
        feature_cols = artifact['features']
        print(f"Loaded existing baseline model (Historical MAE: +/- EUR {model_mae:.2f})")
    except FileNotFoundError:
        print("Error: Model artifact not found. Please run train_model.py first.")
        return

    # 2. Fetch the Latest Gold Data from the Data Lake
    print("Fetching today's live Gold features from S3...")
    try:
        df = pd.read_parquet("s3://eu-energy-raw-ireland-sj/gold/ml_features/")
    except Exception as e:
        print(f"Error reading from S3: {e}")
        return

    # 3. Dynamically Target the Day-Ahead Market (T+24 Hours)
    # Calculates tomorrow's exact date based on the server's current execution time
    tomorrow = datetime.now() + timedelta(days=1)
    target_date = tomorrow.date()

    # Casts the timestamp column to proper datetime objects for accurate filtering
    df['utc_timestamp'] = pd.to_datetime(df['utc_timestamp'])

    # Isolates the target market and filters strictly for tomorrow's forecast features
    target_country = 'Germany'
    df_market = df[df['country'] == target_country].copy()
    day_ahead_data = df_market[df_market['utc_timestamp'].dt.date == target_date].copy()

    # Grabs the first available hourly block for the target day
    day_ahead_data = day_ahead_data.dropna().head(1)

    if day_ahead_data.empty:
        print(f"Error: No live data available for the target date ({target_date}).")
        return

    # Extracts the exact features the model was trained on
    X_live = day_ahead_data[feature_cols]
    live_timestamp = day_ahead_data['utc_timestamp'].iloc[0]

    # 4. Execute the Quantitative ML Prediction
    print(f"\nPredicting day-ahead price for: {live_timestamp}")
    baseline_price = model.predict(X_live)[0]
    print(f"Quantitative ML Prediction: EUR {baseline_price:.2f}/MWh")

    # 5. Fetch the Qualitative LLM Overlay
    print("\nTriggering LLM Trader for Market Sentiment Analysis...")
    llm_response = llm_trader.get_market_sentiment_multiplier(
        baseline_prediction=baseline_price, 
        model_mae=model_mae
    )
    
    # Parses the structured JSON output from the LLM
    multiplier = llm_response.get("multiplier", 1.0)
    sentiment = llm_response.get("sentiment", "Neutral")
    reasoning = llm_response.get("reasoning", "No reasoning provided.")
    
    # Calculates the final adjusted price output
    final_price = baseline_price * multiplier
    
    # 6. Display the Final System Output
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

if __name__ == "__main__":
    generate_daily_prediction()