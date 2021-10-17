package main

import (
	"flag"
	"log"

	"github.com/MarekWojt/gertdns/config"
	"github.com/MarekWojt/gertdns/dns"
)

var (
	configFile = flag.String("configFile", "conf.toml", "Path to configuration file")
)

func main() {
	flag.Parse()

	_, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s\n ", err.Error())
	}

	err = dns.Run()
	if err != nil {
		log.Fatalf("Failed to start DNS server: %s\n ", err.Error())
	}
}
