package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joseph-gunnarsson/solar-cast/internals/clients"
	"github.com/joseph-gunnarsson/solar-cast/internals/solar"
	"github.com/ringsaturn/tzf"
)

type BaseHandler struct {
	solarPanelData   map[string]solar.SolarPanelData
	defaultPanelData map[string]solar.SolarPanelData
}

func NewBaseHandler(solarPanelData map[string]solar.SolarPanelData) *BaseHandler {
	defaults := map[string]solar.SolarPanelData{
		"Mono-Default-400": {
			ModelNo:                    "Mono-Default-400",
			MaximumPowerPmax:           400,
			TemperatureCoefficientPmax: -0.0035,
			NOCT_Temp:                  45,
		},
		"Poly-Default-340": {
			ModelNo:                    "Poly-Default-340",
			MaximumPowerPmax:           340,
			TemperatureCoefficientPmax: -0.0040,
			NOCT_Temp:                  45,
		},
		"Thin-Default-150": {
			ModelNo:                    "Thin-Default-150",
			MaximumPowerPmax:           150,
			TemperatureCoefficientPmax: -0.0025,
			NOCT_Temp:                  47,
		},
	}

	return &BaseHandler{
		solarPanelData:   solarPanelData,
		defaultPanelData: defaults,
	}
}

func (h *BaseHandler) solarPanelAutoCompleteHandler(rw http.ResponseWriter, r *http.Request) {
	response := []string{}
	query := r.PathValue("panel")
	if query == "" {
		http.Error(rw, "Query parameter 'panel' is required", http.StatusBadRequest)
		return
	}
	for modelName := range h.solarPanelData {
		if strings.Contains(strings.ToLower(modelName), strings.ToLower(query)) {
			response = append(response, modelName)
			if len(response) >= 5 {
				break
			}
		}
	}
	writeJSON(rw, http.StatusOK, response)
}

func (h *BaseHandler) getSolarPanel(rw http.ResponseWriter, r *http.Request) {
	query := r.PathValue("panel")
	panel, exists := h.solarPanelData[query]
	if !exists {
		http.Error(rw, "Panel not found", http.StatusNotFound)
		return
	}
	writeJSON(rw, http.StatusOK, panel)
}

// ---- New handlers ----

func (h *BaseHandler) locationAutocompleteHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	results, err := clients.GeocodeCity(r.Context(), q)
	if err != nil {
		http.Error(w, "geocoding failed", http.StatusBadGateway)
		return
	}
	if len(results) > 5 {
		results = results[:5]
	}
	writeJSON(w, http.StatusOK, results)
}

type estimateReq struct {
	Panel    string  `json:"panel"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
	Timezone *string `json:"timezone,omitempty"`
}

func (h *BaseHandler) estimateHandler(w http.ResponseWriter, r *http.Request) {
	var req estimateReq
	var p solar.SolarPanelData
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	p, ok := h.solarPanelData[req.Panel]
	if !ok {
		p, ok = h.defaultPanelData[req.Panel]
		if !ok {
			http.Error(w, "unknown panel", http.StatusBadRequest)
			return
		}
	}

	if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
		http.Error(w, "lat/lon out of range", http.StatusBadRequest)
		return
	}
	tz := "UTC"
	if req.Timezone != nil && *req.Timezone != "" {
		tz = *req.Timezone
	} else {
		tz = timeZoneFinder(req.Lat, req.Lon)
	}
	log.Printf("Timezone for %f, %f: %s", req.Lat, req.Lon, tz)
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	nowLocal := time.Now().In(loc)
	day := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)

	wp, err := clients.FetchHourlyWeather(r.Context(), req.Lat, req.Lon, day, tz)
	if err != nil {
		http.Error(w, "weather fetch failed", http.StatusBadGateway)
		return
	}

	points, totalBase, totalLow, totalHigh, err := solar.
		CalculateHourlyOutputFromWeatherWithRange(p, wp, req.Lat)
	if err != nil {
		http.Error(w, "calc failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"panel":       req.Panel,
		"lat":         req.Lat,
		"lon":         req.Lon,
		"timezone":    wp.Timezone,
		"date":        day.Format("2006-01-02"),
		"totalWh":     totalBase,
		"totalLowWh":  totalLow,
		"totalHighWh": totalHigh,
		"points":      points,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func timeZoneFinder(lat, lon float64) string {
	finder, err := tzf.NewDefaultFinder()
	if err != nil {
		return "Etc/UTC"
	}
	loc := finder.GetTimezoneName(lon, lat)
	return loc
}
