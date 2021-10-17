package config

var Config *Configuration = nil

type Configuration struct {
	Port    uint16
	Host    string
	Domains []string
}

func Load() error {
	Config = &Configuration{
		Port:    5353,
		Host:    "0.0.0.0",
		Domains: []string{},
	}
	return nil
}
