package nestmon

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

var (
	NestResponse NestAPIResponse
)

const (
	TokenURL   = "https://api.home.nest.com/oauth2/access_token"
	NestAPIURL = "https://developer-api.nest.com"
)

func StartNestmonLoop(queryInterval *time.Duration, nc *NestmonConfig) {
	for {
		t := time.Now()
		log.Printf("Requesting data at %v.\n", t.Format(time.RFC3339))
		nestResponseData := getNestResponse(nc)
		processNestResponse(nestResponseData, nc)
		log.Printf("Sleeping for %v.\n", time.Duration(*queryInterval))
		time.Sleep(*queryInterval)
	}
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

	log.Printf("Request: %+v.\n", r)
	resp, _ := customClient.Do(r)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var prettyJson bytes.Buffer
	json.Indent(&prettyJson, bodyBytes, "=", "\t")
	log.Printf("Response: %v.\n", resp)
	log.Printf("Pretty JSON: %v.\n", prettyJson.String())

	err := json.Unmarshal(bodyBytes, &NestResponse)
	if err != nil {
		log.Printf("Error in unmarshalling NestAPIResponse JSON: %v.\n", err)
	}

	return NestResponse
}

func ParseConfig(configPath string, c *NestmonConfig) {
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
	// Given NestResponse, act on it.
	log.Printf("NestJson.Devices: %v.\n", nr.Devices.Thermostats)
	for key, value := range nr.Devices.Thermostats {
		log.Printf("Thermostats key: %+v, value: %+v.\n", key, value)
	}
	for key, value := range nr.Structures {
		log.Printf("Structures, key: %+v, value: %+v.\n", key, value)
	}
	insertNestAPIResponseIntoInfluxDB(nr, c)
}

func insertNestAPIResponseIntoInfluxDB(nr NestAPIResponse, nc *NestmonConfig) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     nc.DbHostUrl,
		Username: nc.DbUsername,
		Password: nc.DbPassword,
	})
	if err != nil {
		log.Fatal(err)
	}

	c.Close()

	// Create a new batch points
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  nc.DbName,
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
		log.Printf("tags: %+v; fields: %+v.\n", tags, fields)
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
}
