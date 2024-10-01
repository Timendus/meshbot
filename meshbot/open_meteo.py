import requests
from datetime import datetime

from .meshwrapper.time_helper import friendly_date


wmo_codes = {
    "0": {
        "day": {
            "description": "Sunny",
            "image": "http://openweathermap.org/img/wn/01d@2x.png",
            "icon": "☀️",
        },
        "night": {
            "description": "Clear",
            "image": "http://openweathermap.org/img/wn/01n@2x.png",
            "icon": "🌙",
        },
    },
    "1": {
        "day": {
            "description": "Mainly Sunny",
            "image": "http://openweathermap.org/img/wn/01d@2x.png",
            "icon": "☀️",
        },
        "night": {
            "description": "Mainly Clear",
            "image": "http://openweathermap.org/img/wn/01n@2x.png",
            "icon": "🌙",
        },
    },
    "2": {
        "day": {
            "description": "Partly Cloudy",
            "image": "http://openweathermap.org/img/wn/02d@2x.png",
            "icon": "⛅️",
        },
        "night": {
            "description": "Partly Cloudy",
            "image": "http://openweathermap.org/img/wn/02n@2x.png",
            "icon": "☁️",
        },
    },
    "3": {
        "day": {
            "description": "Cloudy",
            "image": "http://openweathermap.org/img/wn/03d@2x.png",
            "icon": "☁️",
        },
        "night": {
            "description": "Cloudy",
            "image": "http://openweathermap.org/img/wn/03n@2x.png",
            "icon": "☁️",
        },
    },
    "45": {
        "day": {
            "description": "Foggy",
            "image": "http://openweathermap.org/img/wn/50d@2x.png",
            "icon": "🌫️",
        },
        "night": {
            "description": "Foggy",
            "image": "http://openweathermap.org/img/wn/50n@2x.png",
            "icon": "🌫️",
        },
    },
    "48": {
        "day": {
            "description": "Rime Fog",
            "image": "http://openweathermap.org/img/wn/50d@2x.png",
            "icon": "🌫️",
        },
        "night": {
            "description": "Rime Fog",
            "image": "http://openweathermap.org/img/wn/50n@2x.png",
            "icon": "🌫️",
        },
    },
    "51": {
        "day": {
            "description": "Light Drizzle",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Light Drizzle",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "53": {
        "day": {
            "description": "Drizzle",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Drizzle",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "55": {
        "day": {
            "description": "Heavy Drizzle",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Heavy Drizzle",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "56": {
        "day": {
            "description": "Light Freezing Drizzle",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Light Freezing Drizzle",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌨️",
        },
    },
    "57": {
        "day": {
            "description": "Freezing Drizzle",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Freezing Drizzle",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌨️",
        },
    },
    "61": {
        "day": {
            "description": "Light Rain",
            "image": "http://openweathermap.org/img/wn/10d@2x.png",
            "icon": "🌦️",
        },
        "night": {
            "description": "Light Rain",
            "image": "http://openweathermap.org/img/wn/10n@2x.png",
            "icon": "🌧️",
        },
    },
    "63": {
        "day": {
            "description": "Rain",
            "image": "http://openweathermap.org/img/wn/10d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Rain",
            "image": "http://openweathermap.org/img/wn/10n@2x.png",
            "icon": "🌧️",
        },
    },
    "65": {
        "day": {
            "description": "Heavy Rain",
            "image": "http://openweathermap.org/img/wn/10d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Heavy Rain",
            "image": "http://openweathermap.org/img/wn/10n@2x.png",
            "icon": "🌧️",
        },
    },
    "66": {
        "day": {
            "description": "Light Freezing Rain",
            "image": "http://openweathermap.org/img/wn/10d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Light Freezing Rain",
            "image": "http://openweathermap.org/img/wn/10n@2x.png",
            "icon": "🌨️",
        },
    },
    "67": {
        "day": {
            "description": "Freezing Rain",
            "image": "http://openweathermap.org/img/wn/10d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Freezing Rain",
            "image": "http://openweathermap.org/img/wn/10n@2x.png",
            "icon": "🌨️",
        },
    },
    "71": {
        "day": {
            "description": "Light Snow",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Light Snow",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "73": {
        "day": {
            "description": "Snow",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Snow",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "75": {
        "day": {
            "description": "Heavy Snow",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Heavy Snow",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "77": {
        "day": {
            "description": "Snow Grains",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Snow Grains",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "80": {
        "day": {
            "description": "Light Showers",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Light Showers",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "81": {
        "day": {
            "description": "Showers",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Showers",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "82": {
        "day": {
            "description": "Heavy Showers",
            "image": "http://openweathermap.org/img/wn/09d@2x.png",
            "icon": "🌧️",
        },
        "night": {
            "description": "Heavy Showers",
            "image": "http://openweathermap.org/img/wn/09n@2x.png",
            "icon": "🌧️",
        },
    },
    "85": {
        "day": {
            "description": "Light Snow Showers",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Light Snow Showers",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "86": {
        "day": {
            "description": "Snow Showers",
            "image": "http://openweathermap.org/img/wn/13d@2x.png",
            "icon": "🌨️",
        },
        "night": {
            "description": "Snow Showers",
            "image": "http://openweathermap.org/img/wn/13n@2x.png",
            "icon": "🌨️",
        },
    },
    "95": {
        "day": {
            "description": "Thunderstorm",
            "image": "http://openweathermap.org/img/wn/11d@2x.png",
            "icon": "🌩️",
        },
        "night": {
            "description": "Thunderstorm",
            "image": "http://openweathermap.org/img/wn/11n@2x.png",
            "icon": "🌩️",
        },
    },
    "96": {
        "day": {
            "description": "Light Thunderstorms With Hail",
            "image": "http://openweathermap.org/img/wn/11d@2x.png",
            "icon": "⛈️",
        },
        "night": {
            "description": "Light Thunderstorms With Hail",
            "image": "http://openweathermap.org/img/wn/11n@2x.png",
            "icon": "⛈️",
        },
    },
    "99": {
        "day": {
            "description": "Thunderstorm With Hail",
            "image": "http://openweathermap.org/img/wn/11d@2x.png",
            "icon": "⛈️",
        },
        "night": {
            "description": "Thunderstorm With Hail",
            "image": "http://openweathermap.org/img/wn/11n@2x.png",
            "icon": "⛈️",
        },
    },
}


def fetch_weather(location) -> str | None:
    try:
        params = {
            "latitude": location[0],
            "longitude": location[1],
            "current": [
                "temperature_2m",
                "is_day",
                "precipitation",
                "weather_code",
            ],
        }
        result = requests.get("https://api.open-meteo.com/v1/forecast", params=params)

        if not result.ok:
            print(
                f"Could not reach the Open-Meteo server at this time: {result.status_code} - {result.text}"
            )
            return None

        weather = result.json()
        temp = weather.get("current", {}).get("temperature_2m", None)
        temp_unit = weather.get("current_units", {}).get("temperature_2m", None)
        precip = weather.get("current", {}).get("precipitation", None)
        precip_unit = weather.get("current_units", {}).get("precipitation", None)

        weather_code = wmo_codes.get(
            str(weather.get("current", {}).get("weather_code", None)), None
        ).get(
            "day" if weather.get("current", {}).get("is_day", 1) == 1 else "night", None
        )
        icon = weather_code.get("icon", None)
        description = weather_code.get("description", None)

        return f"{icon + ' ' if icon else ''}It's {temp}{temp_unit}. {description + '.' if description else ''}{' '+ str(precip) + precip_unit if precip > 0 else ''}"
    except Exception as e:
        print(e)
        return None


def fetch_forecast(location) -> str | None:
    try:
        params = {
            "latitude": location[0],
            "longitude": location[1],
            "daily": [
                "weather_code",
                "temperature_2m_max",
                "temperature_2m_min",
                "precipitation_sum",
                "precipitation_probability_max",
                "wind_speed_10m_max",
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

        forecast_string = ""
        for i, date in enumerate(daily.get("time", [])):
            date = datetime.strptime(date, "%Y-%m-%d")
            forecast_string += f"""{friendly_date(date)} - {wmo_codes.get(str(daily.get("weather_code", "")[i]), {}).get("day", {}).get("icon", "")}
Temp: Min {daily.get("temperature_2m_min", "")[i]}{units.get("temperature_2m_min", "")} - Max {daily.get("temperature_2m_max", "")[i]}{units.get("temperature_2m_max", "")}
Precip: {daily.get("precipitation_sum", "")[i]}{units.get("precipitation_sum", "")} - {daily.get("precipitation_probability_max", "")[i]}{units.get("precipitation_probability_max", "")} chance
Wind: {daily.get("wind_speed_10m_max", "")[i]}{units.get("wind_speed_10m_max", "")}

"""

        return forecast_string
    except Exception as e:
        print(e)
        return None
