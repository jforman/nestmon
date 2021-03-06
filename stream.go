package nestmon

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	flagConfigPath    string
	StreamingResponse NestAPIStreamingResponse
)

func StreamingStatusLoop(c chan NestAPIStreamingResponse, accessToken string) {
	u, _ := url.ParseRequestURI(NestAPIURL)
	urlStr := u.String()
	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	customClient := http.Client{
		CheckRedirect: func(redirRequest *http.Request, via []*http.Request) error {
			redirRequest.Header = req.Header
			if len(via) >= 10 {
				return errors.New("Stopped after 10 redirects")
			}
			return nil
		},
		// TODO: Make actual timeouts work where if the stream dies
		// the connection is dropped and retried.
		// Timeout: 300 * time.Second,
	}

	resp, _ := customClient.Do(req)
	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()
	for scanner.Scan() {
		st := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error in reading Nest HTTP Response: %v.\n", err)
			continue
		}
		response, err := getNestAPIResponse(st)
		if err != nil {
			fmt.Printf("Error from getNestAPIResponse: %v.\n", err)
			continue
		}
		if response.Data != nil {
			c <- response
		}
	}
}

func GetNestStructData(d string) NestAPIStreamingResponse {
	// Given string JSON of Nest API response, return NestAPIStreamingResponse object
	err := json.Unmarshal([]byte(d), &StreamingResponse)
	if err != nil {
		fmt.Printf("Error in unmarshalling NestAPIResponse JSON: %v.\n", err)
	}
	return StreamingResponse
}

func getNestAPIResponse(b string) (NestAPIStreamingResponse, error) {
	// Given a string of bytes, return either nil or an NestAPIResponse struct
	var emptyResponse NestAPIStreamingResponse
	httpData := strings.SplitN(b, ":", 2)
	if len(httpData) == 1 {
		// Empty line
		return emptyResponse, nil
	}
	value := strings.TrimSpace(httpData[1])
	switch prefix := strings.TrimSpace(httpData[0]); prefix {
	case "event":
		if value == "keep-alive" {
			// TODO: Handle lack of keep-alives.
		}
		return emptyResponse, nil
	case "data":
		if value != "null" {
			StreamingResponse = GetNestStructData(strings.TrimSpace(httpData[1]))
		}
	}
	return StreamingResponse, nil
}
