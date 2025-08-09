package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/joseph-gunnarsson/solar-cast/internals/scraping"
)

type BaseHandler struct {
	solarPanelData map[string]scraping.SolarPanelData
}

func NewBaseHandler(solarPanelData map[string]scraping.SolarPanelData) *BaseHandler {
	return &BaseHandler{
		solarPanelData: solarPanelData,
	}
}

func (h *BaseHandler) solarPanelAutoCompleteHandler(rw http.ResponseWriter, r *http.Request) {
	response := []string{}
	query := r.PathValue("panel")
	if query == "" {
		http.Error(rw, "Query parameter 'query' is required", http.StatusBadRequest)
		return
	}

	for model_name := range h.solarPanelData {
		if strings.Contains(strings.ToLower(model_name), strings.ToLower(query)) {
			response = append(response, model_name)
			if len(response) >= 5 {
				break
			}
		}
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
}

func (h *BaseHandler) getSolarPanel(rw http.ResponseWriter, r *http.Request) {
	query := r.PathValue("panel")
	panel, exists := h.solarPanelData[query]
	if !exists {
		http.Error(rw, "Panel not found", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(panel)
}
