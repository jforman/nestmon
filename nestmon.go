package nestmon

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

var (
	NestResponse         NestAPIResponse
	enableNest           = flag.Bool("enableNest", false, "Enable checking for Nest data and inserting into database.")
	enableWeather        = flag.Bool("enableWeather", false, "Enable checking the local weather and inserting into database")
	weatherQueryInterval = time.Duration(1 * time.Minute)
)

const (
	TokenURL   = "https://api.home.nest.com/oauth2/access_token"
	NestAPIURL = "https://developer-api.nest.com"
)

func WeatherLoop(nc *NestmonConfig) {
	log.Printf("Weather data enabled for location: %v.\n", nc.OWMZip)
	for {
		log.Printf("Requesting weather data")
		weather, err := getWeatherData(nc.OWMZip, nc.OWMAppId)
		if err != nil {
			log.Printf("Error retreiving weather data: %s.\n", err)
			return
		}
		log.Println("Finished retrieving weather.")
		err = insertWeatherDataIntoDatabase(weather, nc)
		if err != nil {
			log.Printf("Error inserting weather data into database: %s.\n", err)
			return
		}
		log.Printf("Weather data retrieval sleeping for %v.\n",
			time.Duration(weatherQueryInterval))
		time.Sleep(weatherQueryInterval)
	}
}

func NestLoop(queryInterval time.Duration, nc *NestmonConfig) {
	for {
		log.Println("Requesting Nest data.")
		nestResponseData := getNestResponse(nc)
		processNestResponse(nestResponseData, nc)
		log.Printf("Nest data retrieval sleeping for %v.\n", time.Duration(queryInterval))
		time.Sleep(queryInterval)
	}
}

func StartNestmonLoop(queryInterval time.Duration, nc *NestmonConfig) {
	var wg sync.WaitGroup
	if *enableWeather {
		wg.Add(1)
		go func() {
			defer wg.Done()
			WeatherLoop(nc)
		}()
	}

	if *enableNest {
		wg.Add(1)
		go func() {
			defer wg.Done()
			go NestLoop(queryInterval, nc)
		}()
	}
	wg.Wait()
}

func getNestResponse(c *NestmonConfig) NestAPIResponse {
	log.Println("Getting Nest Data.")
	u, _ := url.ParseRequestURI(NestAPIURL)
	urlStr := u.String()
	r, _ := http.NewRequest("GET", urlStr, nil)
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+c.AccessToken)

	customClient := http.Client{
		CheckRedirect: func(redirRequest *http.Request, via []*http.Request) error {
			redirRequest.Header = r.Header

			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	//log.Printf("Request: %+v.\n", r)
	resp, _ := customClient.Do(r)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var prettyJson bytes.Buffer
	json.Indent(&prettyJson, bodyBytes, "=", "\t")
	// log.Printf("Response: %v.\n", resp)
	// log.Printf("Pretty JSON: %v.\n", prettyJson.String())

	err := json.Unmarshal(bodyBytes, &NestResponse)
	if err != nil {
		log.Printf("Error in unmarshalling NestAPIResponse JSON: %v.\n", err)
	}

	return NestResponse
}

func ParseConfig(configPath string, c *NestmonConfig) {
	// Read Nestmon JSON and set NestmonConfig variable
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error reading config: %v.\n", err.Error())
	}
	err = json.Unmarshal(raw, &c)
	if err != nil {
		log.Fatalf("In PraseConfig, Error in unmarshalling the JSON: %v.\n", err)
	}
}

func processNestResponse(nr NestAPIResponse, c *NestmonConfig) {
	insertNestAPIResponseIntoInfluxDB(nr, c)
}

func insertNestAPIResponseIntoInfluxDB(nr NestAPIResponse, nc *NestmonConfig) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     nc.DbHostUrl,
		Username: nc.NestDbUsername,
		Password: nc.NestDbPassword,
	})
	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	// Create a new batch points
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  nc.NestDbName,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Loop over thermostats and add their current state to the batch
	for key, value := range nr.Devices.Thermostats {
		structureName := nr.Structures[value.StructureID].Name
		tags := map[string]string{
			"name":      value.Name,
			"key":       key,
			"structure": structureName,
		}
		fields := map[string]interface{}{
			"AmbientTemperatureF": value.AmbientTemperatureF,
			"Humidity":            value.Humidity,
			"HvacState":           value.HvacState,
			"SoftwareVersion":     value.SoftwareVersion,
		}
		// log.Printf("tags: %+v; fields: %+v.\n", tags, fields)
		pt, err := client.NewPoint("thermostat", tags, fields, time.Now())
		if err != nil {
			log.Printf("Error when adding NewPoint: %+v.\n", err)
		} else {
			bp.AddPoint(pt)
		}

		// Write the batch to Influx database
		if err := c.Write(bp); err != nil {
			log.Printf("Error writing batch: %+v.\n", err)
		}

		// Close client resources
		if err := c.Close(); err != nil {
			log.Printf("Error closing client resources: %+v.\n", err)
		}
	}
	log.Println("Successfully inserted Nest data into DB.")
}

func getWeatherData(zip string, appID string) (OWMResponse, error) {
	log.Println("Getting Weather Data.")
	var owmr OWMResponse
	weatherURL := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?units=imperial&zip=%s&APPID=%s",
		zip, appID)

	req, err := http.Get(weatherURL)
	if err != nil {
		log.Printf("Error retrieving weather data at %s: %s.\n",
			weatherURL, err)
		return owmr, err
	}
	resp, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		log.Printf("Error reading response body: %s.\n", err)
		return owmr, err
	}
	err = json.Unmarshal(resp, &owmr)
	if err != nil {
		log.Printf("Error in reading OpenWeatherMap Response JSON: %s.\n", err)
		return owmr, err
	}
	return owmr, nil
}

func insertWeatherDataIntoDatabase(w OWMResponse, nc *NestmonConfig) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     nc.DbHostUrl,
		Username: nc.WeatherDbUsername,
		Password: nc.WeatherDbPassword,
	})
	if err != nil {
		log.Printf("Error creating NewHTTPClient: %v.\n", err)
		return err
	}

	defer c.Close()

	// Create a new batch points
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  nc.WeatherDbName,
		Precision: "s",
	})
	if err != nil {
		log.Printf("Error when creating NewBatchPoints: %v.\n", err)
		return err
	}
	tags := map[string]string{
		"location": nc.OWMZip,
	}
	fields := map[string]interface{}{
		"TemperatureF":    w.Main.Temp,
		"HumidityPercent": w.Main.Humidity,
		"WindSpeed":       w.Wind.Speed,
	}

	pt, err := client.NewPoint("weather", tags, fields, time.Now())
	if err != nil {
		log.Printf("Error when adding NewPoint: %+v.\n", err)
	} else {
		bp.AddPoint(pt)
	}

	// Write the batch to Influx database
	if err := c.Write(bp); err != nil {
		log.Printf("Error writing batch: %+v.\n", err)
		return err
	}
	//
	// Close client resources
	if err := c.Close(); err != nil {
		log.Printf("Error closing client resources: %+v.\n", err)
		return err
	}
	return nil
}
