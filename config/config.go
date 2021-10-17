package config

type Configuration struct {
	Port    uint16
	Host    string
	Domains []string
}

var (
	Config Configuration

	defaultConfig = Configuration{
		Port:    5353,
		Host:    "0.0.0.0",
		Domains: []string{},
	}
)

func Load(configFilePath string) (*Configuration, error) {
	Config, err := loadConfFile(configFilePath)
	return &Config, err
}
