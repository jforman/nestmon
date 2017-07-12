package main

// https://developers.nest.com/documentation/cloud/how-to-read-data
import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type NestmonConfig struct {
	AccessToken string `json:"accessToken"`
}

const (
	TokenURL   = "https://api.home.nest.com/oauth2/access_token"
	NestAPIURL = "https://developer-api.nest.com"
)

type NestResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func getNestData(c *NestmonConfig) {
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
	fmt.Printf("Response: %v.\n", resp)
	fmt.Printf("Response Body: %v.\n", string(bodyBytes))
}

func parseConfig(configPath string, c *NestmonConfig) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading config: %v.\n", err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal(raw, &c)
	if err != nil {
		fmt.Printf("Error in unmarshalling the JSON:: %v.\n", err)
		os.Exit(1)
	}
}

func main() {
	var (
		flagConfigPath = flag.String("config", "", "JSON config containing Nest API access parameters.")
		config         NestmonConfig
	)
	flag.Parse()
	parseConfig(*flagConfigPath, &config)
	fmt.Println("Let's get this show on the road.")
	fmt.Printf("In main, after parse, config: %+v.\n", config.AccessToken)

	for {
		t := time.Now()
		fmt.Printf("Requesting data at %v.\n", t.Format(time.RFC3339))
		getNestData(&config)
		time.Sleep(1 * time.Minute)
	}

}
