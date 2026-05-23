import os
import json
from dotenv import load_dotenv
from groq import Groq

# Load environment variables from the .env file in the root directory
load_dotenv()

# Initialize the Groq client
client = Groq(api_key=os.environ.get("GROQ_API_KEY"))

def test_senior_trader():
    print("Connecting to Groq API and initializing Senior Trader...")
    
    # The System Prompt is the brain of your agent. It defines the rules.
    system_prompt = """
    You are a Senior Energy Trader in the European power market. 
    Your objective is to read recent energy news and adjust a machine learning baseline price prediction.
    
    RULES:
    1. You will receive a summary of recent news.
    2. You must determine if the news is Bullish (drives prices up), Bearish (drives prices down), or Neutral.
    3. You must output a mathematical multiplier between 0.95 and 1.05.
       - 1.05 = Maximum panic / severe supply shortage
       - 1.00 = Neutral / business as usual
       - 0.95 = Massive oversupply / demand crash
    4. You MUST output your response in strict JSON format. Do not include any other text.
    
    JSON FORMAT:
    {
        "reasoning": "A one sentence explanation of your market sentiment.",
        "sentiment": "Bullish",
        "multiplier": 1.02
    }
    """
    
    # A dummy news headline to test the logic
    dummy_news = "Germany announces immediate closure of three major coal power plants due to environmental regulations. Wind output is forecasted to be low this week."
    
    try:
        response = client.chat.completions.create(
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"Here is the recent news: {dummy_news}"}
            ],
            model="llama-3.3-70b-versatile",
            temperature=0.0, # Set to 0 for maximum logical consistency, no creative hallucination
            response_format={"type": "json_object"} # Force Groq to return clean JSON
        )
        
        # Parse and print the result
        result = json.loads(response.choices[0].message.content)
        print("\nAPI Connection Successful. Trader Response:")
        print(json.dumps(result, indent=4))
        
    except Exception as e:
        print(f"\nError connecting to API: {e}")

if __name__ == "__main__":
    test_senior_trader()