package main

// https://developers.nest.com/documentation/cloud/how-to-read-data
import (
	"flag"
	"fmt"
	"github.com/jforman/nestmon"
)

func main() {
	var (
		flagConfigPath = flag.String("config", "", "JSON config containing Nest API access parameters.")
		config         nestmon.NestmonConfig
	)
	flag.Parse()
	nestmon.ParseConfig(*flagConfigPath, &config)
	n := make(chan nestmon.NestAPIStreamingResponse)
	go func() {
		nestmon.StreamingStatusLoop(n, config.AccessToken)
	}()

	for status := range n {
		printHomeStatus(status)
	}
}

func printHomeStatus(h nestmon.NestAPIStreamingResponse) {
	if h.Data.Devices != nil {
		for key, value := range h.Data.Devices.Thermostats {
			fmt.Printf("Streaming Thermostats key: %+v, value: %+v.\n", key, value)
		}
	}
	for key, value := range h.Data.Structures {
		fmt.Printf("Structures, key: %+v, value: %+v.\n", key, value)
	}

}
