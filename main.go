package main

import (
	"flag"
	"log"

	"github.com/MarekWojt/gertdns/config"
	"github.com/MarekWojt/gertdns/dns"
	"github.com/MarekWojt/gertdns/web"
	rootDns "github.com/miekg/dns"
)

var (
	configFile = flag.String("configFile", "conf.toml", "Path to configuration file")
)

type dnsResult struct {
	server *rootDns.Server
	err    error
}

func main() {
	flag.Parse()

	err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s\n ", err.Error())
	}

	dns.Init()
	web.Init()

	webChan := make(chan error)
	dnsChan := make(chan dnsResult)
	go func() {
		server, err := dns.Run()
		if err != nil {
			log.Fatalf("Failed to start DNS server: %s\n ", err.Error())
		}

		dnsChan <- dnsResult{
			server: server,
			err:    err,
		}
	}()

	go func() {
		err := web.RunSocket()
		if err != nil {
			log.Fatalf("Failed to start HTTP socket: %s\n ", err.Error())
		}

		webChan <- err
	}()

	go func() {
		err := web.RunHTTP()
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %s\n ", err.Error())
		}

		webChan <- err
	}()

	currentDnsResult := <-dnsChan
	defer currentDnsResult.server.Shutdown()
	<-webChan
	<-webChan
}
