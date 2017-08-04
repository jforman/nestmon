package main

// https://developers.nest.com/documentation/cloud/how-to-read-data
import (
	"flag"
	"fmt"
	"github.com/jforman/nestmon"
	"time"
)

func main() {
	var (
		flagConfigPath = flag.String("config", "", "JSON config containing Nest API access parameters.")
		queryInterval  = flag.Duration("query_interval", 3*time.Minute, "Interval between Nest API queries.")
		config         nestmon.NestmonConfig
	)
	flag.Parse()
	nestmon.ParseConfig(*flagConfigPath, &config)
	fmt.Println("Let's get this show on the road.")
	fmt.Printf("In main, after parse, config: %+v.\n", config.AccessToken)

	nestmon.StartNestmonLoop(queryInterval, &config)
}
