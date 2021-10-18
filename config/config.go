package config

type Configuration struct {
	DNS  DNSConfiguration
	HTTP HTTPConfiguration
}

type DNSConfiguration struct {
	Port    uint16
	Host    string
	Domains []string
}

type HTTPConfiguration struct {
	Port           uint16
	Host           string
	Socket         string
	SocketFileMode uint32
}

var (
	Config Configuration

	defaultConfig = Configuration{
		DNS: DNSConfiguration{
			Port:    5353,
			Host:    "0.0.0.0",
			Domains: []string{},
		},
		HTTP: HTTPConfiguration{
			Port:           8080,
			Host:           "127.0.0.1",
			Socket:         "",
			SocketFileMode: 0644,
		},
	}
)

func Load(configFilePath string) error {
	config, err := loadConfFile(configFilePath)
	Config = config
	return err
}
