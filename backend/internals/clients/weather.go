package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HourWeather struct {
	Time          time.Time
	AmbientTemp   float64
	IrradianceGHI float64
}

type WeatherPack struct {
	Timezone string `json:"timezone"`
	Hours    []HourWeather
}

type WeatherAPIResponse struct {
	Timezone string `json:"timezone"`
	Hourly   struct {
		Time               []string  `json:"time"`
		Temperature2m      []float64 `json:"temperature_2m"`
		ShortwaveRadiation []float64 `json:"shortwave_radiation"`
	} `json:"hourly"`
}

func FetchHourlyWeather(ctx context.Context, lat, lon float64, day time.Time, timezone string) (WeatherPack, error) {
	dayStr := day.Format("2006-01-02")

	q := url.Values{}
	q.Set("latitude", fmt.Sprintf("%.6f", lat))
	q.Set("longitude", fmt.Sprintf("%.6f", lon))
	q.Set("hourly", "temperature_2m,shortwave_radiation")
	q.Set("timezone", timezone)
	q.Set("start_date", dayStr)
	q.Set("end_date", dayStr)

	u := "https://api.open-meteo.com/v1/forecast?" + q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "solar-cast/1.0 (+contact@example.com)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return WeatherPack{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return WeatherPack{}, fmt.Errorf("open-meteo: HTTP %d", resp.StatusCode)
	}

	var apiResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return WeatherPack{}, err
	}

	n := len(apiResp.Hourly.Time)
	if n == 0 {
		return WeatherPack{}, fmt.Errorf("open-meteo: empty hourly data")
	}

	hours := make([]HourWeather, 0, n)
	for i := 0; i < n; i++ {
		locTime, err := parseOMTime(apiResp.Hourly.Time[i], apiResp.Timezone)
		if err != nil {
			return WeatherPack{}, err
		}
		hours = append(hours, HourWeather{
			Time:          locTime,
			AmbientTemp:   apiResp.Hourly.Temperature2m[i],
			IrradianceGHI: apiResp.Hourly.ShortwaveRadiation[i],
		})
	}

	return WeatherPack{
		Timezone: apiResp.Timezone,
		Hours:    hours,
	}, nil
}

func parseOMTime(s, tz string) (time.Time, error) {

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	if t, err := time.ParseInLocation("2006-01-02T15:04", s, loc); err == nil {
		return t, nil
	}

	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, loc); err == nil {
		return t, nil
	}
	if t, err := time.ParseInLocation("2006-01-02 15:04", s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q with tz %q", s, tz)
}
