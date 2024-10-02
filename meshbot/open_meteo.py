import requests
import json
from datetime import datetime

from .meshwrapper.time_helper import friendly_date


wmo_codes = json.loads(open("./meshbot/wmo_codes.json").read())


def fetch_weather(position) -> str | None:
    try:
        params = {
            "latitude": position[0],
            "longitude": position[1],
            "current": [
                "temperature_2m",
                "is_day",
                "precipitation",
                "weather_code",
                "wind_speed_10m",
                "wind_direction_10m",
            ],
        }
        result = requests.get("https://api.open-meteo.com/v1/forecast", params=params)

        if not result.ok:
            print(
                f"Could not reach the Open-Meteo server at this time: {result.status_code} - {result.text}"
            )
            return None

        weather = result.json()
        weather_code = wmo_codes.get(
            str(weather.get("current", {}).get("weather_code", None)), {}
        ).get(
            "day" if weather.get("current", {}).get("is_day", 1) == 1 else "night", {}
        )

        icon = weather_code.get("icon", "")
        description = weather_code.get("description", "")
        temp = weather.get("current", {}).get("temperature_2m", "")
        temp_unit = weather.get("current_units", {}).get("temperature_2m", "")
        precip = weather.get("current", {}).get("precipitation", "")
        precip_unit = weather.get("current_units", {}).get("precipitation", "")
        wind_speed = weather.get("current", {}).get("wind_speed_10m", "")
        wind_speed_unit = weather.get("current_units", {}).get("wind_speed_10m", "")
        wind_dir = wind_direction(
            weather.get("current", {}).get("wind_direction_10m", None)
        )

        return f"""🌡️  {temp}{temp_unit}
{icon}  {description}
💧  {precip}{precip_unit}
🌬️  {wind_speed}{wind_speed_unit} {wind_dir}
"""
    except Exception as e:
        print(e)
        return None


def fetch_forecast(position) -> str | None:
    try:
        params = {
            "latitude": position[0],
            "longitude": position[1],
            "daily": [
                "weather_code",
                "temperature_2m_max",
                "temperature_2m_min",
                "precipitation_sum",
                "precipitation_probability_max",
                "wind_speed_10m_max",
                "wind_direction_10m_dominant",
            ],
            "timezone": "auto",
        }
        result = requests.get("https://api.open-meteo.com/v1/forecast", params=params)

        if not result.ok:
            print(
                f"Could not reach the Open-Meteo server at this time: {result.status_code} - {result.text}"
            )
            return None

        forecast = result.json()
        daily = forecast.get("daily", None)
        units = forecast.get("daily_units", None)

        # Rewrite dictionary of arrays to array of dictionaries, rename some
        # things, add some units. In short, do all the pre-processing.
        structured_forecast = {}
        for key, value in daily.items():
            if key == "time":
                key = "day"
            if key == "weather_code":
                key = "icon"
            if type(value) == list:
                for i, v in enumerate(value):
                    if i not in structured_forecast:
                        structured_forecast[i] = {}
                    if key == "day":
                        v = friendly_date(datetime.strptime(v, "%Y-%m-%d"))
                    if key == "icon":
                        weather_code = wmo_codes.get(str(v), {}).get("day", {})
                        v = weather_code.get("icon", "")
                        structured_forecast[i]["description"] = weather_code.get(
                            "description", ""
                        )
                    if key == "wind_direction_10m_dominant":
                        v = wind_direction(v)
                    else:
                        v = f"{v}{units.get(key, '')}"
                    structured_forecast[i][key] = v

        forecast_string = ""
        for day in list(structured_forecast.values())[:6]:
            forecast_string += f"""▬▬ {day["day"]} ▬▬
🌡️  {day["temperature_2m_max"]} / {day["temperature_2m_min"]}
{day["icon"]}  {day["description"]}
💧  {day["precipitation_sum"]} {day["precipitation_probability_max"]}
🌬️  {day["wind_speed_10m_max"]} {day["wind_direction_10m_dominant"]}

"""

        return forecast_string
    except Exception as e:
        print(e)
        return None


def wind_direction(direction) -> str:
    match direction:
        case dir if 0 <= dir < 22.5:
            return "↓"
        case dir if 22.5 <= dir < 67.5:
            return "↙"
        case dir if 67.5 <= dir < 112.5:
            return "←"
        case dir if 112.5 <= dir < 157.5:
            return "↖"
        case dir if 157.5 <= dir < 202.5:
            return "↑"
        case dir if 202.5 <= dir < 247.5:
            return "↗"
        case dir if 247.5 <= dir < 292.5:
            return "→"
        case dir if 292.5 <= dir < 337.5:
            return "↘"
        case dir if 337.5 <= dir < 360:
            return "↓"
        case _:
            return ""
