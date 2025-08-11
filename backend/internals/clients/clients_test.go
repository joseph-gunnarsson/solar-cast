package clients

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestFetchHourlyWeather_Live(t *testing.T) {
	// London, today
	lat := 51.5074
	lon := -0.1278
	day := time.Now()

	wp, err := FetchHourlyWeather(context.Background(), lat, lon, day, "Europe/London")
	if err != nil {
		t.Fatalf("FetchHourlyWeather error: %v", err)
	}

	log.Printf("Timezone: %s", wp.Timezone)
	for _, h := range wp.Hours {
		log.Printf("%s  Temp: %.1f°C  Irradiance: %.1f W/m²",
			h.Time.Format(time.RFC3339),
			h.AmbientTemp,
			h.IrradianceGHI,
		)
	}
}
