package config

import (
	"io/ioutil"
	"os"

	"github.com/gookit/color"
	"github.com/pelletier/go-toml/v2"
)

func loadConfFile(configFilePath string) (Configuration, error) {
	bytes, err := ioutil.ReadFile(configFilePath)

	if err != nil {
		color.Errorln(err.Error())
		color.Warnln("Creating new configuration file")
		return createNewConfig(configFilePath)
	}

	var cfg Configuration
	err = toml.Unmarshal(bytes, &cfg)
	if err != nil {
		color.Errorln(err.Error())
		color.Errorln("Configuration file " + configFilePath + " could not be read as TOML file")
		return defaultConfig, err
	}
	color.Infoln("Loaded configuration file " + configFilePath)
	return cfg, err
}

func createNewConfig(configFilePath string) (Configuration, error) {
	confBytes, err := toml.Marshal(defaultConfig)
	if err != nil {
		color.Errorln(err.Error())
		color.Errorln("Default config struct is not TOML conform")
		return defaultConfig, err
	}
	err = os.WriteFile(configFilePath, confBytes, 0644)
	if err != nil {
		color.Errorln(err.Error())
		color.Warnln("Using default config without file")
	}
	return defaultConfig, err
}
