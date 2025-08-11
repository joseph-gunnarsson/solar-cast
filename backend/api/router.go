package api

import (
	"net/http"

	"github.com/joseph-gunnarsson/solar-cast/internals/solar"
)

func Router(solarPanelData map[string]solar.SolarPanelData) *http.ServeMux {
	mux := http.NewServeMux()
	h := NewBaseHandler(solarPanelData)

	// Solar panel search and details
	mux.HandleFunc("GET /api/solar-panels/search/{panel}", h.solarPanelAutoCompleteHandler)
	mux.HandleFunc("GET /api/solar-panels/{panel}", h.getSolarPanel)

	// Location autocomplete (lat, lon, timezone)
	mux.HandleFunc("GET /api/location/autocomplete", h.locationAutocompleteHandler)

	// Solar estimate (coords only)
	mux.HandleFunc("POST /api/solar/estimate", h.estimateHandler)

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return mux
}
