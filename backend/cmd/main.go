package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/joseph-gunnarsson/solar-cast/api"
	"github.com/joseph-gunnarsson/solar-cast/internals/scraping"
	solarcalc "github.com/joseph-gunnarsson/solar-cast/internals/solar_calc"
)

func main() {

	panel := scraping.SolarPanelData{
		ModelNo:                    "SKT410M10-108D4",
		NOCT_Temp:                  45.0,
		TemperatureCoefficientPmax: -0.0029,
		MaximumPowerPmax:           410.0,
	}
	ambient := 30.0     // °C
	irradiance := 850.0 // W/m²
	hours := 1.0        // equivalent full‑sun hours

	wh, err := solarcalc.CalculateSolarPanelOutputByHour(panel, ambient, irradiance, hours)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Energy in 1h: %.2f Wh\n", wh)

	_, inDocker := os.LookupEnv("DOCKER_CONTAINER")
	if _, err := os.Stat("/.dockerenv"); err == nil {
		inDocker = true
	}

	if !inDocker {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	solarPanelData, err := scraping.LoadSolarPanelData()
	if err != nil {
		log.Println("Error loading solar panel data:", err)
		return
	}
	fmt.Println("Loaded solar panel data for", len(solarPanelData), "models.")

	server := api.Router(solarPanelData)
	port := os.Getenv("backend_port")

	http.ListenAndServe(":"+port, server)
	log.Println("Server started on port", port)

}
