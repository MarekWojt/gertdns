package dns

import (
	"errors"
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

var domains []*domain = make([]*domain, 0)

func Init() {
	for _, currentDomain := range config.Config.DNS.Domains {
		log.Printf("Added domain root: %s\n", currentDomain)
		domains = append(domains, &domain{
			Root: currentDomain,
			Ipv4: make(map[string]string),
			Ipv6: make(map[string]string),
		})
	}
}

func Run() (*dns.Server, error) {
	// attach request handler func
	for _, currentDomain := range domains {
		dns.HandleFunc(currentDomain.Root, handleDnsRequest(currentDomain))
	}

	// start server
	server := &dns.Server{Addr: ":" + strconv.Itoa(int(config.Config.DNS.Port)), Net: "udp"}
	log.Printf("Starting DNS at %d\n", config.Config.DNS.Port)
	err := server.ListenAndServe()
	if err != nil {
		server.Shutdown()
		return server, err
	}

	return server, nil
}

func UpdateIpv6(domain string, ipv6 string) error {
	for _, currentDomain := range domains {
		if dns.IsSubDomain(currentDomain.Root, domain) {
			log.Printf("Updating domain %s AAAA %s\n", domain, ipv6)
			currentDomain.Mutv6.Lock()
			currentDomain.Ipv6[domain] = ipv6
			currentDomain.Mutv6.Unlock()
			return nil
		}
	}

	return errors.New("no root found")
}

func UpdateIpv4(domain string, ipv4 string) (err error) {
	for _, currentDomain := range domains {
		if dns.IsSubDomain(currentDomain.Root, domain) {
			log.Printf("Updating domain %s A %s\n", domain, ipv4)
			currentDomain.Mutv4.Lock()
			currentDomain.Ipv4[domain] = ipv4
			currentDomain.Mutv4.Unlock()
			return nil
		}
	}

	return errors.New("no root found")
}

func Get() []*domain {
	return domains
}

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
