package api

import (
	"net/http"

	"github.com/joseph-gunnarsson/solar-cast/internals/scraping"
)

func Router(solar_panel_data map[string]scraping.SolarPanelData) *http.ServeMux {
	mux := http.NewServeMux()
	bh := NewBaseHandler(solar_panel_data)

	mux.HandleFunc("GET /api/solar-panels/{panel}", bh.getSolarPanel)

	mux.HandleFunc("GET /api/solar-panels/search/{panel}", bh.solarPanelAutoCompleteHandler)

	return mux
}
