package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/clementd64/mangad/pkg/mangad"
)

var endpoint = flag.String("url", "", "Tachidesk endpoint")
var configFile = flag.String("config", "config.yaml", "Config file")
var sleepTime = flag.Duration("interval", 0, "Update interval (0 to disable)")
var waitForIt = flag.Bool("wait-for-it", false, "Wait for tachidesk to be running")

func main() {
	flag.Parse()

	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}

	t := mangad.New(*endpoint)

	if *waitForIt {
		log.Print("Waiting for tachidesk to be running")
		for t.Ping() != nil {
			time.Sleep(1 * time.Second)
		}
		log.Print("Tachidesk is running")
	}

	conf, err := mangad.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	if *sleepTime == time.Duration(0) {
		t.Run(conf)
	} else {
		for {
			t.Run(conf)
			time.Sleep(*sleepTime)
		}
	}
}
