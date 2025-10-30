// This entire file way ported from Python using ChatGPT. So it's probably
// shite, but it does seem to work. So I'm just going to use it as a black box
// and be done with it.

package weather

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//go:embed wmo_codes.json
var wmoCodesJSON string

// Position represents a geographical coordinate.
type Position struct {
	Latitude  float64
	Longitude float64
}

// WeatherInfo represents weather icon and description.
type WeatherInfo struct {
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

// WmoCode holds both day and night weather info.
type WmoCode struct {
	Day   WeatherInfo `json:"day"`
	Night WeatherInfo `json:"night"`
}

// wmoCodes is loaded from the JSON file at initialization.
var wmoCodes map[string]WmoCode

func init() {
	// Load wmo_codes.json
	err := json.Unmarshal([]byte(wmoCodesJSON), &wmoCodes)
	if err != nil {
		log.Printf("Error parsing wmo_codes.json: %v", err)
		wmoCodes = make(map[string]WmoCode)
	}
}

// friendlyDate formats a date in a friendly way.
// Adjust the format string as needed to match your friendly_date helper.
func friendlyDate(t time.Time) string {
	return t.Format("Mon Jan 2")
}

// windDirection converts a numeric wind direction into an arrow string.
func windDirection(direction float64) string {
	switch {
	case direction >= 0 && direction < 22.5:
		return "↓"
	case direction >= 22.5 && direction < 67.5:
		return "↙"
	case direction >= 67.5 && direction < 112.5:
		return "←"
	case direction >= 112.5 && direction < 157.5:
		return "↖"
	case direction >= 157.5 && direction < 202.5:
		return "↑"
	case direction >= 202.5 && direction < 247.5:
		return "↗"
	case direction >= 247.5 && direction < 292.5:
		return "→"
	case direction >= 292.5 && direction < 337.5:
		return "↘"
	case direction >= 337.5 && direction < 360:
		return "↓"
	default:
		return ""
	}
}

// FetchWeather retrieves the current weather at the given position.
func FetchWeather(position Position) (string, error) {
	baseURL := "https://api.open-meteo.com/v1/forecast"
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", position.Latitude))
	params.Set("longitude", fmt.Sprintf("%f", position.Longitude))

	// Add current weather parameters
	for _, p := range []string{
		"temperature_2m",
		"is_day",
		"precipitation",
		"weather_code",
		"wind_speed_10m",
		"wind_direction_10m",
	} {
		params.Add("current", p)
	}

	fullURL := baseURL + "?" + params.Encode()
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("could not reach the Open-Meteo server at this time: %d - %s", resp.StatusCode, string(body))
	}

	var weather map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return "", err
	}

	current, ok := weather["current"].(map[string]interface{})
	if !ok {
		return "", errors.New("no current weather data")
	}

	// Retrieve weather code and check day or night.
	codeVal := current["weather_code"]
	codeStr := ""
	switch v := codeVal.(type) {
	case float64:
		codeStr = strconv.Itoa(int(v))
	case string:
		codeStr = v
	}

	isDay := 1
	if v, ok := current["is_day"].(float64); ok {
		isDay = int(v)
	}

	var weatherInfo WeatherInfo
	if code, found := wmoCodes[codeStr]; found {
		if isDay == 1 {
			weatherInfo = code.Day
		} else {
			weatherInfo = code.Night
		}
	}

	icon := weatherInfo.Icon
	description := weatherInfo.Description
	temp := fmt.Sprintf("%v", current["temperature_2m"])

	// Retrieve units
	currentUnits, _ := weather["current_units"].(map[string]interface{})
	tempUnit := ""
	if currentUnits != nil {
		if v, ok := currentUnits["temperature_2m"].(string); ok {
			tempUnit = v
		}
	}
	precip := fmt.Sprintf("%v", current["precipitation"])
	precipUnit := ""
	if currentUnits != nil {
		if v, ok := currentUnits["precipitation"].(string); ok {
			precipUnit = v
		}
	}
	windSpeed := fmt.Sprintf("%v", current["wind_speed_10m"])
	windSpeedUnit := ""
	if currentUnits != nil {
		if v, ok := currentUnits["wind_speed_10m"].(string); ok {
			windSpeedUnit = v
		}
	}

	windDirFloat := 0.0
	if v, ok := current["wind_direction_10m"].(float64); ok {
		windDirFloat = v
	}
	windDir := windDirection(windDirFloat)

	// Format the result string
	result := fmt.Sprintf(
		"🌡️  %s%s\n%s  %s\n💧  %s%s\n🌬️  %s%s %s\n",
		temp, tempUnit,
		icon, description,
		precip, precipUnit,
		windSpeed, windSpeedUnit, windDir,
	)
	return result, nil
}

// FetchForecast retrieves the weather forecast for the given position.
func FetchForecast(position Position) (string, error) {
	baseURL := "https://api.open-meteo.com/v1/forecast"
	params := url.Values{}
	params.Set("latitude", fmt.Sprintf("%f", position.Latitude))
	params.Set("longitude", fmt.Sprintf("%f", position.Longitude))

	// Add daily forecast parameters
	for _, p := range []string{
		"weather_code",
		"temperature_2m_max",
		"temperature_2m_min",
		"precipitation_sum",
		"precipitation_probability_max",
		"wind_speed_10m_max",
		"wind_direction_10m_dominant",
	} {
		params.Add("daily", p)
	}
	params.Set("timezone", "auto")

	fullURL := baseURL + "?" + params.Encode()
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("could not reach the Open-Meteo server at this time: %d - %s", resp.StatusCode, string(body))
	}

	var forecast map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return "", err
	}

	daily, ok := forecast["daily"].(map[string]interface{})
	if !ok {
		return "", errors.New("no daily forecast data")
	}

	units, _ := forecast["daily_units"].(map[string]interface{})

	// Create a slice of maps (one per day) from the dictionary-of-arrays.
	timeArr, ok := daily["time"].([]interface{})
	if !ok {
		return "", errors.New("daily time data not found")
	}
	n := len(timeArr)
	structuredForecast := make([]map[string]string, n)
	for i := 0; i < n; i++ {
		structuredForecast[i] = make(map[string]string)
	}

	// Iterate over daily keys and populate each day’s data.
	for key, val := range daily {
		arr, ok := val.([]interface{})
		if !ok {
			continue
		}
		newKey := key
		if key == "time" {
			newKey = "day"
		}
		if key == "weather_code" {
			newKey = "icon"
		}
		for i, v := range arr {
			var valueStr string
			if newKey == "day" {
				// Parse date string and format it.
				if dateStr, ok := v.(string); ok {
					t, err := time.Parse("2006-01-02", dateStr)
					if err == nil {
						valueStr = friendlyDate(t)
					} else {
						valueStr = dateStr
					}
				}
			} else if newKey == "icon" {
				// Lookup weather code and set both icon and description.
				codeStr := ""
				switch cv := v.(type) {
				case float64:
					codeStr = strconv.Itoa(int(cv))
				case string:
					codeStr = cv
				}
				if code, found := wmoCodes[codeStr]; found {
					valueStr = code.Day.Icon
					structuredForecast[i]["description"] = code.Day.Description
				}
			} else if key == "wind_direction_10m_dominant" {
				// Convert numeric wind direction.
				if dir, ok := v.(float64); ok {
					valueStr = windDirection(dir)
				}
			} else {
				// Append unit if available.
				unit := ""
				if units != nil {
					if u, ok := units[key].(string); ok {
						unit = u
					}
				}
				valueStr = fmt.Sprintf("%v%s", v, unit)
			}
			structuredForecast[i][newKey] = valueStr
		}
	}

	// Build the forecast string (limit to 3 days if available)
	forecastStr := ""
	limit := 3
	if n < limit {
		limit = n
	}
	for i := 0; i < limit; i++ {
		day := structuredForecast[i]
		forecastStr += fmt.Sprintf("▬▬ %s ▬▬\n", day["day"])
		forecastStr += fmt.Sprintf("🌡️  %s / %s\n", day["temperature_2m_max"], day["temperature_2m_min"])
		forecastStr += fmt.Sprintf("%s  %s\n", day["icon"], day["description"])
		forecastStr += fmt.Sprintf("💧  %s %s\n", day["precipitation_sum"], day["precipitation_probability_max"])
		forecastStr += fmt.Sprintf("🌬️  %s %s\n\n", day["wind_speed_10m_max"], day["wind_direction_10m_dominant"])
	}

	return forecastStr, nil
}
