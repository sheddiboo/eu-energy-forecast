import os
import joblib
import pandas as pd
from sklearn.metrics import mean_absolute_error
from sklearn.ensemble import RandomForestRegressor

def retrain_baseline_model():
    print("--- EU Energy MLOps: Weekly Model Retraining ---")

    # 1. Fetch the full historical Gold layer data from S3
    print("Fetching full Gold layer dataset from S3...")
    try:
        df = pd.read_parquet("s3://eu-energy-raw-ireland-sj/gold/ml_features/")
    except Exception as e:
        print(f"Error reading from S3: {e}. Ensure AWS credentials and s3fs are configured.")
        return

    # Sorts the time-series data chronologically
    df = df.sort_values(by=['country', 'utc_timestamp'])
    
    # Isolates the target market for model training
    target_country = 'Germany'
    df_market = df[df['country'] == target_country].copy()
    
    # Removes rows with NaN values caused by lagging features
    df_clean = df_market.dropna()

    if df_clean.empty:
        print(f"Error: Not enough clean data for {target_country}.")
        return

    # 2. Prepare Features and Target
    target_col = 'price_eur_per_mwh'
    feature_cols = [col for col in df_clean.columns if col not in ['utc_timestamp', 'country', target_col]]

    # Excludes the absolute latest day to avoid testing on incomplete future data
    train_data = df_clean.iloc[:-1]
    
    X_train = train_data[feature_cols]
    y_train = train_data[target_col]

    # 3. Initialize and Train the Model
    print(f"Training Random Forest Regressor on {len(X_train)} historical records...")
    model = RandomForestRegressor(n_estimators=100, random_state=42, n_jobs=-1)
    model.fit(X_train, y_train)

    # 4. Evaluate Model Performance (MAE)
    predictions = model.predict(X_train)
    model_mae = mean_absolute_error(y_train, predictions)
    print(f"Training Complete. Historical MAE: +/- EUR {model_mae:.2f}/MWh")

    # 5. Package and Save the Model Artifact
    # Bundles the trained model and its performance metrics into a single dictionary
    model_artifact = {
        'model': model,
        'mae': model_mae,
        'features': feature_cols
    }

    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    model_path = os.path.join(project_root, "assets", "intelligence", "baseline_model.pkl")
    
    joblib.dump(model_artifact, model_path)
    print(f"Model artifact successfully saved to: {model_path}")

if __name__ == "__main__":
    retrain_baseline_model()