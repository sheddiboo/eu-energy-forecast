CREATE TABLE eu_energy_db.gold_ml_features
WITH (
  format = 'PARQUET',
  external_location = 's3://eu-energy-raw-ireland-sj/gold/ml_features/'
) AS
WITH imputed_data AS (
    SELECT 
        utc_timestamp,
        country,
        LAST_VALUE(price_eur_per_mwh) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as price_eur_per_mwh,
        LAST_VALUE(actual_load_mw) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as actual_load_mw,
        LAST_VALUE(forecasted_load_mw) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as forecasted_load_mw,
        LAST_VALUE(actual_generation_mw) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as actual_generation_mw,
        LAST_VALUE(forecasted_wind_solar_mw) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as forecasted_wind_solar_mw,
        LAST_VALUE(weather_temperature_celsius) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as weather_temperature_celsius,
        LAST_VALUE(weather_wind_speed_kmh) IGNORE NULLS OVER (PARTITION BY country ORDER BY utc_timestamp) as weather_wind_speed_kmh
    FROM eu_energy_db.silver_master_energy_data
),
features AS (
    SELECT 
        *,
        EXTRACT(HOUR FROM utc_timestamp) AS hour_of_day,
        EXTRACT(DOW FROM utc_timestamp) AS day_of_week,
        LAG(price_eur_per_mwh, 24) OVER (PARTITION BY country ORDER BY utc_timestamp) AS price_24h_ago,
        LAG(price_eur_per_mwh, 168) OVER (PARTITION BY country ORDER BY utc_timestamp) AS price_168h_ago,
        (actual_load_mw - forecasted_load_mw) AS load_forecast_error,
        (actual_generation_mw - forecasted_wind_solar_mw) AS wind_solar_error,
        AVG(weather_temperature_celsius) OVER (PARTITION BY country ORDER BY utc_timestamp ROWS BETWEEN 24 PRECEDING AND CURRENT ROW) AS temp_24h_rolling
    FROM imputed_data
)
SELECT * FROM features
WHERE price_24h_ago IS NOT NULL;