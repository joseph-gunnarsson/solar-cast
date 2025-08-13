package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/joseph-gunnarsson/solar-cast/api"
	"github.com/joseph-gunnarsson/solar-cast/internals/scraping"
	"github.com/joseph-gunnarsson/solar-cast/internals/solar"
	red "github.com/joseph-gunnarsson/solar-cast/redis"
)

func main() {
	scrape := flag.Bool("scrape", false, "run web scraping to collect panel data")
	serve := flag.Bool("serve", false, "start the HTTP server")
	pages := flag.Int("pages", 1, "number of pages to scrape from ENF Solar listing")
	flag.Parse()

	loadEnvIfLocal()

	var scraped map[string]solar.SolarPanelData
	var err error

	if *scrape {
		scraped, err = runScrape(*pages)
		if err != nil {
			log.Fatalf("scrape failed: %v", err)
		}
		log.Printf("scrape complete (%d models).", len(scraped))
	}

	if *serve {

		data := scraped
		if len(data) == 0 {
			data, err = solar.LoadSolarPanelData()
			if err != nil {
				log.Fatalf("load panel data failed: %v", err)
			}
		}
		if err := runServer(data); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
		return
	}

	if !*scrape && !*serve {
		flag.Usage()
	}
}

func runScrape(pages int) (map[string]solar.SolarPanelData, error) {
	log.Printf("Starting scrape for %d page(s)…", pages)
	urls := scraping.GetProductURLs(pages)
	if len(urls) == 0 {
		log.Println("No product URLs found.")
		return map[string]solar.SolarPanelData{}, nil
	}

	data := scraping.GatherSolarPanelData(urls)
	if len(data) == 0 {
		log.Println("Scrape returned 0 panels.")
		return data, nil
	}

	if err := solar.SaveSolarPanelDataToFile(data); err != nil {
		return nil, err
	}
	log.Printf("Saved %d panels to data/solar_panel_data.json", len(data))
	return data, nil
}

func runServer(panelData map[string]solar.SolarPanelData) error {
	log.Printf("Loaded solar panel data for %d models.", len(panelData))
	redisClient := red.GetRedisConnection()
	mux := api.Router(panelData, redisClient)
	port := os.Getenv("backend_port")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on :%s", port)
	return srv.ListenAndServe()
}

func loadEnvIfLocal() {
	inDocker := false
	if _, ok := os.LookupEnv("DOCKER_CONTAINER"); ok {
		inDocker = true
	}
	if _, err := os.Stat("/.dockerenv"); err == nil {
		inDocker = true
	}
	if !inDocker {
		_ = godotenv.Load()
	}
}
