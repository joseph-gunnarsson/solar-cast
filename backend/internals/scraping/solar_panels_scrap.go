package scraping

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
}

func GetProductURLs() []string {
	var productURLs []string
	c := colly.NewCollector(
		colly.AllowedDomains("www.enfsolar.com", "enfsolar.com"),
	)

	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*enfsolar.*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to set rate limit:", err)
	}

	userAgentIndex := 0

	c.OnRequest(func(r *colly.Request) {

		r.Headers.Set("User-Agent", userAgents[userAgentIndex])

		r.Headers.Set("Referer", "https://www.enfsolar.com/pv/panel")
	})

	found404 := false
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("src"), "404billboard.jpg") {
			found404 = true
		}
	})

	c.OnHTML("a.enf-product-name", func(e *colly.HTMLElement) {
		fullURL := e.Request.AbsoluteURL(e.Attr("href"))
		productURLs = append(productURLs, fullURL)
	})

	for page := 1; page < 2; page++ {
		if page > 1 && (page-1)%5 == 0 {
			userAgentIndex = (userAgentIndex + 1) % len(userAgents)
			log.Println("----------------------------------------------------")
			log.Printf("Page %d: Rotating to next User-Agent (index %d).", page, userAgentIndex)
			log.Println("----------------------------------------------------")
		}

		pageURL := "https://www.enfsolar.com/pv/panel?page=" + strconv.Itoa(page)
		found404 = false
		log.Printf("Visiting page: %s (using UA index %d)", pageURL, userAgentIndex)
		err := c.Visit(pageURL)
		log.Printf("Length of productURLs: %d", len(productURLs))
		if err != nil && strings.Contains(err.Error(), "Not Found") {
			log.Println("Page not found:", pageURL)
			found404 = true
			break
		}

		if err != nil {
			log.Println("Visit failed:", err)
			time.Sleep(45 * time.Second)

			continue
		}

		if found404 {
			fmt.Println("404 detected on page", page, "- stopping.")
			break
		}
	}

	return productURLs
}

type SolarPanelData struct {
	NOCT_Temp                  float64 `json:"noct_temp"`
	ModelNo                    string  `json:"model_no"`
	TemperatureCoefficientPmax float64 `json:"temperature_coefficient_pmax"`
	MaximumPowerPmax           float64 `json:"maximum_power_pmax"`
}

func parseWatts(wattStr string) float64 {
	wattStr = strings.TrimSpace(strings.Replace(wattStr, "Wp", "", -1))
	val, _ := strconv.ParseFloat(wattStr, 64)
	return val
}

func parseTemperature(tempStr string) float64 {
	val, _ := strconv.ParseFloat(strings.TrimSpace(strings.Split(tempStr, "±")[0]), 64)
	return val
}

func parseTempCoeff(coeffStr string) float64 {

	coeffStr = strings.TrimSpace(strings.Replace(coeffStr, "%/°C", "", -1))
	val, _ := strconv.ParseFloat(coeffStr, 64)
	return val / 100.0
}

func GatherSolarPanelData(urls []string) map[string]SolarPanelData {

	solarPanelData := SolarPanelData{}
	solarPanelDataMap := make(map[string]SolarPanelData)

	c := colly.NewCollector(
		colly.AllowedDomains("www.enfsolar.com", "enfsolar.com"),
	)
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*enfsolar.*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})
	if err != nil {
		log.Fatal("Failed to set rate limit:", err)
	}
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		th := e.ChildText("th")
		if th == "Temperature" {
			td := strings.Trim(e.DOM.Find("td").First().Text(), " \n\r\t")

			solarPanelData.NOCT_Temp = parseTemperature(td)
			fmt.Println("NOCT Temperature:", solarPanelData.NOCT_Temp)
		}
		if th == "Model No." {
			td := strings.Trim(e.DOM.Find("td table tr td").First().Text(), " \n\r\t")

			solarPanelData.ModelNo = td
			fmt.Println("Model No:", solarPanelData.ModelNo)
		}
		if th == "Temperature Coefficient of Pmax" {
			td := strings.Trim(e.DOM.Find("td").First().Text(), " \n\r\t")

			solarPanelData.TemperatureCoefficientPmax = parseTempCoeff(td)
			fmt.Println("Temperature Coefficient of Pmax:", solarPanelData.TemperatureCoefficientPmax)
		}
		if th == "Maximum Power (Pmax)" {
			td := strings.Trim(e.DOM.Find("td table tr td").First().Text(), " \n\r\t")

			solarPanelData.MaximumPowerPmax = parseWatts(td)
			fmt.Println("Maximum Power Pmax:", solarPanelData.MaximumPowerPmax)
		}

	})

	for _, url := range urls {
		log.Println("Visiting product URL:", url)
		err := c.Visit(url)
		if err != nil {
			log.Println("Failed to visit product URL:", url, "Error:", err)
			break
		}
		solarPanelDataMap[solarPanelData.ModelNo] = solarPanelData
		solarPanelData = SolarPanelData{}
	}
	fmt.Println("Gathered solar panel data for", len(solarPanelDataMap), "models.")
	log.Println("Finished gathering solar panel data.")
	return solarPanelDataMap
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
