package api

import (
	"log"
	"net/http"
	"os"

	"github.com/joseph-gunnarsson/solar-cast/internals/solar"
	"github.com/redis/go-redis/v9"
)

func Router(solarPanelData map[string]solar.SolarPanelData, redisClient *redis.Client) *http.ServeMux {
	mux := http.NewServeMux()
	h := NewBaseHandler(solarPanelData, redisClient)
	adminToken := os.Getenv("ADMIN_TOKEN_SECRET")
	mux.HandleFunc("GET /api/solar-panels/search/{panel}", h.solarPanelAutoCompleteHandler)
	mux.HandleFunc("GET /api/solar-panels/{panel}", h.getSolarPanel)

	mux.HandleFunc("GET /api/location/autocomplete", h.locationAutocompleteHandler)

	mux.HandleFunc("POST /api/solar/estimate", h.estimateHandler)

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("POST /api/admin/reload", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("adminToken: %s, x-Admin-Token: %s", adminToken, r.Header.Get("X-Admin-Token"))
		if adminToken == "" || r.Header.Get("X-Admin-Token") != adminToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		newData, err := solar.LoadSolarPanelData()
		if err != nil {
			http.Error(w, "reload failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		h.swapData(newData) // atomic snapshot swap
		log.Println("Solar panel data reloaded successfully.")
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
