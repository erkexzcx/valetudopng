package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/erkexzcx/valetudopng/pkg/config"
	"github.com/erkexzcx/valetudopng/pkg/server"
)

var (
	version string

	flagConfigFile = flag.String("config", "config.yml", "Path to configuration file")
	flagVersion    = flag.Bool("version", false, "prints version of the application")
)

func main() {
	flag.Parse()

	if *flagVersion {
		fmt.Println("Version:", version)
		return
	}

	c, err := config.NewConfig(*flagConfigFile)
	if err != nil {
		log.Fatalln("Failed to read configuration file:", err)
	}

	server.Start(c)
}
