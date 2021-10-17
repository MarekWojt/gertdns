package dns

import (
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/MarekWojt/gertdns/config"
	"github.com/miekg/dns"
)

type domain struct {
	Root  string
	Mutv4 sync.RWMutex
	Mutv6 sync.RWMutex
	Ipv4  map[string]string
	Ipv6  map[string]string
}

var Domains = []*domain{}

func parseQuery(m *dns.Msg, currentDomain *domain) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for A record of %s\n", q.Name)
			currentDomain.Mutv4.RLock()
			ip := currentDomain.Ipv4[q.Name]
			currentDomain.Mutv4.RUnlock()
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		case dns.TypeAAAA:
			log.Printf("Query for AAAA record of %s\n", q.Name)
			currentDomain.Mutv6.RLock()
			ip := currentDomain.Ipv6[q.Name]
			currentDomain.Mutv6.RUnlock()
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
}

func handleDnsRequest(currentDomain *domain) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Compress = false

		switch r.Opcode {
		case dns.OpcodeQuery:
			parseQuery(m, currentDomain)
		}

		w.WriteMsg(m)
	}
}

func Run() error {
	// attach request handler func
	for _, currentDomain := range Domains {
		dns.HandleFunc(currentDomain.Root, handleDnsRequest(currentDomain))
	}

	// start server
	port := config.Config.Port
	server := &dns.Server{Addr: ":" + strconv.Itoa(int(port)), Net: "udp"}
	log.Printf("Starting DNS at %d\n", port)
	err := server.ListenAndServe()
	if err != nil {
		server.Shutdown()
		return err
	}

	return nil
}
