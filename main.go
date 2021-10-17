package main

import (
	"log"

	"github.com/MarekWojt/gertdns/config"
	"github.com/MarekWojt/gertdns/dns"
)

func main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %s\n ", err.Error())
	}

	err = dns.Run()
	if err != nil {
		log.Fatalf("Failed to start DNS server: %s\n ", err.Error())
	}
}
