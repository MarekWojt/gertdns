package config

import (
	"io/ioutil"
	"os"

	"github.com/gookit/color"
	"github.com/pelletier/go-toml/v2"
)

func loadConfFile(configFilePath string) Configuration {
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
		panic(err)
	}
	color.Infoln("Loaded configuration file " + configFilePath)
	return cfg
}

func createNewConfig(configFilePath string) Configuration {
	confBytes, err := toml.Marshal(defaultConfig)
	if err != nil {
		color.Errorln(err.Error())
		color.Errorln("Default config struct is not TOML conform")
		panic(err)
	}
	err = os.WriteFile(configFilePath, confBytes, 0644)
	if err != nil {
		color.Errorln(err.Error())
		color.Warnln("Using default config without file")
	}
	return defaultConfig
}
