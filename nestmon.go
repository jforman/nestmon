package nestmon

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
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
		fmt.Printf("Requesting data at %v.\n", t.Format(time.RFC3339))
		nestResponseData := getNestResponse(nc)
		processNestResponse(nestResponseData, nc)
		fmt.Printf("Sleeping for %v.\n", time.Duration(*queryInterval))
		time.Sleep(*queryInterval)
	}
}

func getNestResponse(c *NestmonConfig) NestAPIResponse {
	fmt.Println("Getting Nest Data.")
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

	fmt.Printf("Request: %+v.\n", r)
	resp, _ := customClient.Do(r)
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var prettyJson bytes.Buffer
	json.Indent(&prettyJson, bodyBytes, "=", "\t")
	fmt.Printf("Response: %v.\n", resp)
	fmt.Printf("Pretty JSON: %v.\n", prettyJson.String())

	err := json.Unmarshal(bodyBytes, &NestResponse)
	if err != nil {
		fmt.Printf("Error in unmarshalling NestAPIResponse JSON: %v.\n", err)
	}

	return NestResponse
}

func ParseConfig(configPath string, c *NestmonConfig) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config: %v.\n", err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(raw, &c)
	if err != nil {
		fmt.Printf("In PraseConfig, Error in unmarshalling the JSON: %v.\n", err)
		os.Exit(1)
	}
}

func processNestResponse(nr NestAPIResponse, c *NestmonConfig) {
	// Given NestResponse, act on it.
	fmt.Println("in processNestRespose")
	fmt.Printf("NestJson.Devices: %v.\n", nr.Devices.Thermostats)
	for key, value := range nr.Devices.Thermostats {
		fmt.Printf("Thermostats key: %+v, value: %+v.\n", key, value)
	}
	for key, value := range nr.Structures {
		fmt.Printf("Structures, key: %+v, value: %+v.\n", key, value)
	}
}
