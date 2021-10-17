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
	root  string
	mutv4 sync.RWMutex
	mutv6 sync.RWMutex
	ipv4  map[string]string
	ipv6  map[string]string
}

var domains = []*domain{}

func parseQuery(m *dns.Msg, currentDomain *domain) {
	for _, q := range m.Question {
		switch q.Qtype {
		case dns.TypeA:
			log.Printf("Query for A record of %s\n", q.Name)
			currentDomain.mutv4.RLock()
			ip := currentDomain.ipv4[q.Name]
			currentDomain.mutv4.RUnlock()
			if ip != "" {
				rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
				if err == nil {
					m.Answer = append(m.Answer, rr)
				}
			}
		case dns.TypeAAAA:
			log.Printf("Query for AAAA record of %s\n", q.Name)
			currentDomain.mutv6.RLock()
			ip := currentDomain.ipv6[q.Name]
			currentDomain.mutv6.RUnlock()
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

func Load() {
	for _, currentDomain := range config.Config.Domains {
		domains = append(domains, &domain{
			root: currentDomain,
		})
	}
}

func Run() (*dns.Server, error) {
	// attach request handler func
	for _, currentDomain := range domains {
		dns.HandleFunc(currentDomain.root, handleDnsRequest(currentDomain))
	}

	// start server
	server := &dns.Server{Addr: ":" + strconv.Itoa(int(config.Config.Port)), Net: "udp"}
	log.Printf("Starting DNS at %d\n", config.Config.Port)
	err := server.ListenAndServe()
	if err != nil {
		server.Shutdown()
		return server, err
	}

	return server, nil
}

func UpdateIpv6(domain string, ipv6 string) {
	for _, currentDomain := range domains {
		if dns.IsSubDomain(currentDomain.root, domain) {
			currentDomain.mutv6.Lock()
			currentDomain.ipv6[domain] = ipv6
			currentDomain.mutv6.Unlock()
			break
		}
	}
}

func UpdateIpv4(domain string, ipv4 string) {
	for _, currentDomain := range domains {
		if dns.IsSubDomain(currentDomain.root, domain) {
			currentDomain.mutv4.Lock()
			currentDomain.ipv4[domain] = ipv4
			currentDomain.mutv4.Unlock()
			break
		}
	}
}
