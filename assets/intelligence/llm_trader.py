import os
import json
import feedparser
from datetime import datetime, timedelta, timezone
from email.utils import parsedate_to_datetime
from dotenv import load_dotenv
from groq import Groq

# Load environment variables (API Key)
load_dotenv()
client = Groq(api_key=os.environ.get("GROQ_API_KEY"))

def fetch_live_news(rss_url="https://energypost.eu/feed/", days_back=3):
    """
    Fetches headlines from an RSS feed published strictly within the last X days.
    """
    print(f"Fetching live intelligence from: {rss_url} (Last {days_back} Days)")
    feed = feedparser.parse(rss_url)
    
    news_summaries = []
    
    # Calculate the cutoff date (timezone aware)
    now = datetime.now(timezone.utc)
    cutoff_date = now - timedelta(days=days_back)
    
    for entry in feed.entries:
        title = entry.title
        
        # Parse the publication date of the article
        if hasattr(entry, 'published'):
            try:
                # RSS feeds use RFC 822 date formats, we parse it to a datetime object
                article_date = parsedate_to_datetime(entry.published)
                
                # Make sure it's timezone aware for accurate comparison
                if article_date.tzinfo is None:
                    article_date = article_date.replace(tzinfo=timezone.utc)
                
                # Check if the article was published within our timeframe
                if article_date >= cutoff_date:
                    formatted_date = article_date.strftime("%Y-%m-%d")
                    news_summaries.append(f"- [{formatted_date}] {title}")
            except Exception as e:
                print(f"Warning: Could not parse date for article '{title}'. Error: {e}")
                
    if not news_summaries:
        return "No major energy news published in the specified timeframe."
        
    return "\n".join(news_summaries)

def get_market_sentiment_multiplier(baseline_prediction=85.50, model_mae=21.10):
    """
    Passes the live news and baseline context to Llama 3.3 and returns the calculated sentiment multiplier.
    """
    # 1. Get the real news
    live_news = fetch_live_news()
    print("\nToday's Headlines:")
    print(live_news)
    
    # 2. Define the Agent's Rules
    system_prompt = f"""
    You are a Senior Energy Trader in the European power market. 
    Your objective is to read recent energy news and adjust a machine learning baseline price prediction.
    
    CONTEXT:
    - The quantitative baseline model predicts tomorrow's price will be EUR {baseline_prediction}/MWh.
    - The model has a historical Mean Absolute Error (MAE) of +/- EUR {model_mae}/MWh.
    - This means the true price is highly likely to fall between EUR {baseline_prediction - model_mae:.2f} and EUR {baseline_prediction + model_mae:.2f}.
    
    RULES:
    1. You will receive a summary of recent news headlines.
    2. You must determine if the overall market sentiment is Bullish (drives prices up), Bearish (drives prices down), or Neutral.
    3. You must output a mathematical multiplier between 0.95 and 1.05.
       - 1.05 = Maximum panic / severe supply shortage
       - 1.00 = Neutral / business as usual
       - 0.95 = Massive oversupply / demand crash
    4. You MUST output your response in strict JSON format. Do not include any other text.
    
    JSON FORMAT:
    {{
        "reasoning": "A one sentence explanation of your market sentiment.",
        "sentiment": "Bullish",
        "multiplier": 1.02
    }}
    """
    
    print("\nConsulting Senior Trader (Llama 3.3) for Sentiment Analysis...")
    try:
        response = client.chat.completions.create(
            messages=[
                {"role": "system", "content": system_prompt},
                {"role": "user", "content": f"Here is the recent news:\n{live_news}"}
            ],
            model="llama-3.3-70b-versatile", 
            temperature=0.0, 
            response_format={"type": "json_object"} 
        )
        
        # 3. Parse the JSON response
        result = json.loads(response.choices[0].message.content)
        return result
        
    except Exception as e:
        print(f"\nError connecting to Groq API: {e}")
        # If the API fails, return a safe neutral multiplier so the pipeline doesn't crash
        return {"reasoning": "API Failure. Defaulting to neutral.", "sentiment": "Neutral", "multiplier": 1.00}

if __name__ == "__main__":
    # Execute the function with our test parameters and print the final JSON
    trader_output = get_market_sentiment_multiplier(baseline_prediction=85.50, model_mae=21.10)
    print("\nFinal LLM Output:")
    print(json.dumps(trader_output, indent=4))