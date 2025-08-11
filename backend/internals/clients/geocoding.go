package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type GeocodeResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

type geocodeResponse struct {
	Results []GeocodeResult `json:"results"`
}

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func GeocodeCity(ctx context.Context, name string) ([]GeocodeResult, error) {
	q := url.Values{}
	q.Set("name", name)
	q.Set("count", "5")
	q.Set("language", "en")
	q.Set("format", "json")

	u := "https://geocoding-api.open-meteo.com/v1/search?" + q.Encode()

	req, _ := http.NewRequestWithContext(ctx, "GET", u, nil)
	req.Header.Set("User-Agent", "solar-cast/1.0 (+contact@example.com)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("geocode: HTTP %d", resp.StatusCode)
	}

	var data geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Results, nil
}
