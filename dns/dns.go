package dns

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MarekWojt/gertdns/config"
	"github.com/MarekWojt/gertdns/util"
	"github.com/gookit/color"
	"github.com/miekg/dns"
)

var domains []*domain = make([]*domain, 0)
var saveTicker *time.Ticker = time.NewTicker(time.Second)
var saving sync.Mutex = sync.Mutex{}
var currentDataPath string

const (
	IPV4_FILE = "v4.csv"
	IPV6_FILE = "v6.csv"
)

type domain struct {
	Root         string
	Mutv4        sync.RWMutex
	Mutv6        sync.RWMutex
	Mutv4Changed sync.RWMutex
	Mutv6Changed sync.RWMutex
	Ipv4         map[string]string
	Ipv4Changed  bool
	Ipv6         map[string]string
	Ipv6Changed  bool
}

func (currentDomain *domain) IsV4Changed() bool {
	currentDomain.Mutv4Changed.RLock()
	result := currentDomain.Ipv4Changed
	currentDomain.Mutv4Changed.RUnlock()

	return result
}

func (currentDomain *domain) IsV6Changed() bool {
	currentDomain.Mutv6Changed.RLock()
	result := currentDomain.Ipv6Changed
	currentDomain.Mutv6Changed.RUnlock()

	return result
}

func (currentDomain *domain) MarkV6Changed(changed bool) {
	currentDomain.Mutv6Changed.Lock()
	currentDomain.Ipv6Changed = changed
	currentDomain.Mutv6Changed.Unlock()
}

func (currentDomain *domain) MarkV4Changed(changed bool) {
	currentDomain.Mutv4Changed.Lock()
	currentDomain.Ipv4Changed = changed
	currentDomain.Mutv4Changed.Unlock()
}

func (currentDomain *domain) SetV4(domain string, ipv4 string) {
	currentDomain.Mutv4.Lock()
	currentDomain.Ipv4[domain] = ipv4
	currentDomain.Mutv4.Unlock()
}

func (currentDomain *domain) SetV6(domain string, ipv6 string) {
	currentDomain.Mutv6.Lock()
	currentDomain.Ipv6[domain] = ipv6
	currentDomain.Mutv6.Unlock()
}

func loadFile(ty string, currentDomain *domain) {
	if ty != IPV4_FILE && ty != IPV6_FILE {
		panic("type passed to loadFile must be either IPV4_FILE or IPV6_FILE")
	}

	filePath := path.Join(currentDataPath, currentDomain.Root+ty)
	f, err := os.Open(filePath)
	if err != nil {
		color.Warnf("Could not load file for domain %s: %s\n", currentDomain.Root, err)
		return
	}
	defer f.Close()

	log.Printf("Reading file: %s", filePath)
	scanner := bufio.NewScanner(f)

	lineCounter := 0
	for scanner.Scan() {
		lineCounter++
		currentLine := scanner.Text()
		cols := strings.Split(currentLine, "\t")
		if len(cols) < 2 {
			color.Warnf("Error reading line %d of ipv4 addresses for domain %s: too few columns\n", lineCounter, currentDomain.Root)
			continue
		}

		if ty == IPV4_FILE {
			currentDomain.Ipv4[cols[0]] = cols[1]
		} else if ty == IPV6_FILE {
			currentDomain.Ipv6[cols[0]] = cols[1]
		}
	}
	color.Infof("Read file: %s\n", filePath)
}

func Init(dataPath string) {
	currentDataPath = dataPath
	for _, currentDomain := range config.Config.DNS.Domains {
		currentDomain = util.ParseDomain(currentDomain)
		log.Printf("Added domain root: %s\n", currentDomain)

		domainObj := &domain{
			Root: currentDomain,
			Ipv4: make(map[string]string),
			Ipv6: make(map[string]string),
		}
		domains = append(domains, domainObj)
		loadFile(IPV4_FILE, domainObj)
		loadFile(IPV6_FILE, domainObj)
	}

	go func() {
		for {
			<-saveTicker.C
			Save()
		}
	}()
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

			if !currentDomain.IsV6Changed() {
				currentDomain.MarkV6Changed(true)
			}

			currentDomain.SetV6(domain, ipv6)
			return nil
		}
	}

	return errors.New("no root found")
}

func UpdateIpv4(domain string, ipv4 string) (err error) {
	for _, currentDomain := range domains {
		log.Printf("%s sub of %s ?\n", domain, currentDomain.Root)
		if dns.IsSubDomain(currentDomain.Root, domain) {
			log.Printf("Updating domain %s A %s\n", domain, ipv4)

			if !currentDomain.IsV4Changed() {
				currentDomain.MarkV4Changed(true)
			}

			currentDomain.SetV4(domain, ipv4)
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
				rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN A %s", q.Name, ip))
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
				rr, err := dns.NewRR(fmt.Sprintf("%s 300 IN AAAA %s", q.Name, ip))
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

func Save() (errs []error) {
	saving.Lock()
	for _, domain := range domains {
		if domain.IsV4Changed() {
			ipv4Data := ""
			domain.Mutv4.RLock()
			for key, val := range domain.Ipv4 {
				ipv4Data += key + "\t" + val + "\n"
			}
			domain.Mutv4.RUnlock()
			err := os.WriteFile(path.Join(currentDataPath, domain.Root+IPV4_FILE), []byte(ipv4Data), 0644)
			if err != nil {
				errs = append(errs, err)
				color.Errorf("Failed to save ipv4 data for domain %s: %s\n", domain.Root, err)
			} else {
				// did successfully save, so mark as saved
				domain.MarkV4Changed(false)
			}
		}

		if domain.IsV6Changed() {
			ipv6Data := ""
			domain.Mutv6.RLock()
			for key, val := range domain.Ipv6 {
				ipv6Data += key + "\t" + val + "\n"
			}
			domain.Mutv6.RUnlock()
			err := os.WriteFile(path.Join(currentDataPath, domain.Root+IPV6_FILE), []byte(ipv6Data), 0644)
			if err != nil {
				errs = append(errs, err)
				color.Errorf("Failed to save ipv6 data for domain %s: %s\n", domain.Root, err)
			} else {
				// did successfully save, so mark as saved
				domain.MarkV6Changed(false)
			}
		}
	}
	saving.Unlock()

	errLen := len(errs)
	if errLen > 0 {
		color.Errorf("%d errors occurred while trying to save\n", errLen)
	}
	return
}

func Shutdown() {
	saveTicker.Stop()
	Save()
}
