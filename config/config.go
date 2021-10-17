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

func Load(configFilePath string) error {
	config, err := loadConfFile(configFilePath)
	Config = config
	return err
}
