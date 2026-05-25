import os
import sys
import joblib

# Ensure Python can find our llm_trader file in the same directory
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import llm_trader

def generate_daily_prediction():
    print("--- EU Energy Day-Ahead Forecaster ---")

    # 1. Load the Quantitative ML Model
    # We step back one directory to the root to ensure absolute paths work
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    model_path = os.path.join(project_root, "assets", "intelligence", "baseline_model.pkl")
    
    print(f"Loading ML model from: assets/intelligence/baseline_model.pkl")
    try:
        model = joblib.load(model_path)
    except FileNotFoundError:
        print("Error: Could not find the baseline model. Make sure the path is correct.")
        return

    # Note: In a fully live production environment, you would fetch today's weather
    # and load data here, convert it to a DataFrame, and run model.predict(live_data).
    # For this architectural test, we will simulate the ML output that your model would generate.
    baseline_price = 85.50
    model_mae = 21.10
    print(f"Quantitative Baseline Prediction: EUR {baseline_price:.2f}/MWh")

    # 2. Get the Qualitative LLM Overlay
    print("\nFetching Market Sentiment Overlay...")
    llm_response = llm_trader.get_market_sentiment_multiplier(
        baseline_prediction=baseline_price, 
        model_mae=model_mae
    )
    
    multiplier = llm_response.get("multiplier", 1.0)
    sentiment = llm_response.get("sentiment", "Neutral")
    reasoning = llm_response.get("reasoning", "No reasoning provided.")
    
    # 3. Calculate Final Adjusted Price
    final_price = baseline_price * multiplier
    
    # 4. Display Final Output
    print("\n--- FINAL DAY-AHEAD FORECAST ---")
    print(f"Baseline ML Price: EUR {baseline_price:.2f}/MWh")
    print(f"Market Sentiment:  {sentiment} (Multiplier: {multiplier})")
    print(f"AI Reasoning:      {reasoning}")
    print(f"Final AI Price:    EUR {final_price:.2f}/MWh")
    print("--------------------------------------")

if __name__ == "__main__":
    generate_daily_prediction()