CREATE TABLE eu_energy_db.silver_master_energy_data
WITH (
  format = 'PARQUET',
  external_location = 's3://eu-energy-raw-ireland-sj/silver/master_data/'
) AS
WITH clean_prices AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(price AS DOUBLE)) AS price_eur_per_mwh
    FROM eu_energy_db.bronze_prices
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
clean_actual_load AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(actual_load AS DOUBLE)) AS actual_load_mw
    FROM eu_energy_db.bronze_actual_load
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
clean_forecast_load AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(forecast_load AS DOUBLE)) AS forecasted_load_mw
    FROM eu_energy_db.bronze_forecast_load
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
clean_actual_generation AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(actual_generation AS DOUBLE)) AS actual_generation_mw
    FROM eu_energy_db.bronze_actual_generation
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
clean_forecast_wind_solar AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(forecast_wind_solar AS DOUBLE)) AS forecasted_wind_solar_mw
    FROM eu_energy_db.bronze_forecast_wind_solar
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
clean_weather AS (
    SELECT 
        date_trunc('hour', TRY_CAST(timestamp AS TIMESTAMP)) AS utc_timestamp,
        regexp_extract("$path", '/([^/_]+)_[^/]*\.csv$', 1) AS country,
        AVG(TRY_CAST(temperature AS DOUBLE)) AS weather_temperature_celsius,
        AVG(TRY_CAST(wind_speed AS DOUBLE)) AS weather_wind_speed_kmh
    FROM eu_energy_db.bronze_weather
    WHERE timestamp IS NOT NULL AND timestamp NOT IN ('Timestamp', 'timestamp')
    GROUP BY 1, 2
),
time_country_spine AS (
    SELECT utc_timestamp, country FROM clean_prices
    UNION SELECT utc_timestamp, country FROM clean_actual_load
    UNION SELECT utc_timestamp, country FROM clean_forecast_load
    UNION SELECT utc_timestamp, country FROM clean_actual_generation
    UNION SELECT utc_timestamp, country FROM clean_forecast_wind_solar
    UNION SELECT utc_timestamp, country FROM clean_weather
)
SELECT 
    spine.utc_timestamp,
    spine.country,
    p.price_eur_per_mwh,
    al.actual_load_mw,
    fl.forecasted_load_mw,
    ag.actual_generation_mw,
    fws.forecasted_wind_solar_mw,
    w.weather_temperature_celsius,
    w.weather_wind_speed_kmh
FROM time_country_spine spine
LEFT JOIN clean_prices p ON spine.utc_timestamp = p.utc_timestamp AND spine.country = p.country
LEFT JOIN clean_actual_load al ON spine.utc_timestamp = al.utc_timestamp AND spine.country = al.country
LEFT JOIN clean_forecast_load fl ON spine.utc_timestamp = fl.utc_timestamp AND spine.country = fl.country
LEFT JOIN clean_actual_generation ag ON spine.utc_timestamp = ag.utc_timestamp AND spine.country = ag.country
LEFT JOIN clean_forecast_wind_solar fws ON spine.utc_timestamp = fws.utc_timestamp AND spine.country = fws.country
LEFT JOIN clean_weather w ON spine.utc_timestamp = w.utc_timestamp AND spine.country = w.country
WHERE spine.utc_timestamp IS NOT NULL
  AND spine.country IN ('Austria', 'Belgium', 'Bulgaria', 'Croatia', 'CzechRepublic', 'Denmark', 'Estonia', 
  'Finland', 'France', 'Germany', 'Greece', 'Hungary', 'Ireland', 'Italy', 'Latvia', 'Lithuania', 'Netherlands', 
  'Norway', 'Poland', 'Portugal', 'Romania', 'Slovakia', 'Slovenia', 'Spain', 'Sweden', 'Switzerland');