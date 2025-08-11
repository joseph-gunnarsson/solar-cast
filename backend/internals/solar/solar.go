package solar

import (
	"encoding/json"
	"os"
)

type SolarPanelData struct {
	NOCT_Temp                  float64 `json:"noct_temp"`
	ModelNo                    string  `json:"model_no"`
	TemperatureCoefficientPmax float64 `json:"temperature_coefficient_pmax"`
	MaximumPowerPmax           float64 `json:"maximum_power_pmax"`
}

func LoadSolarPanelData() (map[string]SolarPanelData, error) {
	data := make(map[string]SolarPanelData)
	jsonData, err := os.ReadFile("data/solar_panel_data.json")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SaveSolarPanelDataToFile(data map[string]SolarPanelData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile("data/solar_panel_data.json", jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}
