package main

// https://developers.nest.com/documentation/cloud/how-to-read-data
import (
	"flag"
	"log"
	"time"

	"github.com/jforman/nestmon"
)

func main() {
	var (
		flagConfigPath = flag.String("config", "", "JSON config containing Nest API access parameters.")
		queryInterval  = flag.Duration("query_interval", 3*time.Minute, "Interval between Nest API queries.")
		config         nestmon.NestmonConfig
	)
	flag.Parse()
	nestmon.ParseConfig(*flagConfigPath, &config)
	log.Println("Starting thermostat status nestmon. Type: Polling.")
	log.Printf("Config: %+v.\n", config)

	nestmon.StartNestmonLoop(*queryInterval, &config)
}
