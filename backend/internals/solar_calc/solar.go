package solarcalc

import (
	"errors"

	"github.com/joseph-gunnarsson/solar-cast/internals/scraping"
)

func EstimatedCellTemperature(panel scraping.SolarPanelData, ambientTemp, irradiance float64) (float64, error) {
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
	panel scraping.SolarPanelData,
	ambientTemp, irradiance, sunlightHours float64,
) (float64, error) {
	// 1) Estimate cell temp
	Tc, err := EstimatedCellTemperature(panel, ambientTemp, irradiance)
	if err != nil {
		return 0, err
	}

	// 2) Instantaneous power [W]:
	//    P = P_STC * (G/1000) * [1 + γ*(Tc - 25)]
	Pinst := panel.MaximumPowerPmax *
		(irradiance / 1000.0) *
		(1.0 + panel.TemperatureCoefficientPmax*(Tc-25.0))

	// 3) Multiply by hours to get Wh
	energyWh := Pinst * sunlightHours

	return energyWh, nil
}
