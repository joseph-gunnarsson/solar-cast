package solar

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/joseph-gunnarsson/solar-cast/internals/clients"
)

type HourlyPoint struct {
	Time           time.Time `json:"time"`
	Ambient        float64   `json:"ambient"`
	GHI            float64   `json:"ghi"`
	EnergyWh       float64   `json:"energyWh"`
	EnergyWhLow    float64   `json:"energyWhLow"`
	EnergyWhHigh   float64   `json:"energyWhHigh"`
	CumulativeWh   float64   `json:"cumulativeWh"`
	CumulativeLow  float64   `json:"cumulativeLow"`
	CumulativeHigh float64   `json:"cumulativeHigh"`
}

const lowBuffer = 0.9

func EstimatedCellTemperature(panel SolarPanelData, ambientTemp, irradiance float64) (float64, error) {
	if ambientTemp < -40 || ambientTemp > 85 {
		return 0, errors.New("ambient temperature out of range")
	}
	if irradiance < 0 || irradiance > 1200 {
		return 0, errors.New("irradiance out of range")
	}

	cellTemp := ambientTemp + ((irradiance / 800.0) * (panel.NOCT_Temp - 20.0))
	return cellTemp, nil
}
func CalculateSolarPanelOutputByHour(
	panel SolarPanelData,
	ambientTemp, irradiance float64,
) (float64, error) {
	Tc, err := EstimatedCellTemperature(panel, ambientTemp, irradiance)
	if err != nil {
		return 0, err
	}

	Pinst := panel.MaximumPowerPmax *
		(irradiance / 1000.0) *
		(1.0 + panel.TemperatureCoefficientPmax*(Tc-25.0))
	if Pinst < 0 {
		Pinst = 0
	}

	energyWh := Pinst * 1.0
	return energyWh, nil
}

var boostEquatorial = [...]float64{
	1.04, 1.03, 1.02, 1.02, 1.02, 1.02, 1.02, 1.02, 1.02, 1.02, 1.03, 1.04,
}
var boostMidLat = [...]float64{
	1.30, 1.25, 1.15, 1.10, 1.05, 1.02, 1.03, 1.08, 1.12, 1.18, 1.25, 1.32,
}
var boostHighLat = [...]float64{
	1.45, 1.35, 1.22, 1.12, 1.06, 1.02, 1.03, 1.09, 1.18, 1.28, 1.38, 1.48,
}

func TiltBoostFactor(lat float64, ts time.Time) float64 {
	m := int(ts.Month()) - 1

	if lat < 0 {
		m = (m + 6) % 12
	}

	alat := math.Abs(lat)

	eq := val(boostEquatorial, m)
	mid := val(boostMidLat, m)
	hi := val(boostHighLat, m)

	switch {
	case alat <= 15:
		return eq
	case alat <= 35:
		w := (alat - 15) / 20
		return lerp(eq, mid, w)
	case alat <= 55:
		return mid
	case alat <= 70:
		w := (alat - 55) / 15
		return lerp(mid, hi, w)
	default:
		return hi
	}
}

func val(arr [12]float64, i int) float64 { return arr[i%12] }
func lerp(a, b, t float64) float64       { return a + (b-a)*clamp01(t) }
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func CalculateHourlyOutputFromWeatherWithRange(
	panel SolarPanelData,
	wp clients.WeatherPack,
	lat float64,
) ([]HourlyPoint, float64, float64, float64, error) {
	points := make([]HourlyPoint, 0, len(wp.Hours))

	var totalBase, totalLow, totalHigh float64

	for _, h := range wp.Hours {

		baseWh, err := CalculateSolarPanelOutputByHour(panel, h.AmbientTemp, h.IrradianceGHI)
		if err != nil {
			return nil, 0, 0, 0, fmt.Errorf("hour %s: %w", h.Time.Format(time.RFC3339), err)
		}

		lowWh := baseWh * lowBuffer
		boost := TiltBoostFactor(lat, h.Time)
		highWh := baseWh * boost

		totalBase += baseWh
		totalLow += lowWh
		totalHigh += highWh

		points = append(points, HourlyPoint{
			Time:           h.Time,
			Ambient:        h.AmbientTemp,
			GHI:            h.IrradianceGHI,
			EnergyWh:       baseWh,
			EnergyWhLow:    lowWh,
			EnergyWhHigh:   highWh,
			CumulativeWh:   totalBase,
			CumulativeLow:  totalLow,
			CumulativeHigh: totalHigh,
		})
	}

	return points, totalBase, totalLow, totalHigh, nil
}
