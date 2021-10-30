package auth

import (
	"io/ioutil"
	"os"

	"github.com/gookit/color"
	"github.com/pelletier/go-toml/v2"
)

func loadAuthFile(authFilePath string) (map[string]*userRaw, error) {
	bytes, err := ioutil.ReadFile(authFilePath)

	if err != nil {
		color.Errorln(err.Error())
		color.Warnln("Creating new authentication file")
		return writeAuthFile(authFilePath, make(map[string]*userRaw))
	}

	var users map[string]*userRaw
	err = toml.Unmarshal(bytes, &users)
	if err != nil {
		color.Errorln(err.Error())
		color.Errorln("Authentication file " + authFilePath + " could not be read as TOML file")
		return users, err
	}
	color.Infoln("Loaded authentication file " + authFilePath)
	return users, err
}

func writeAuthFile(authFilePath string, users map[string]*userRaw) (map[string]*userRaw, error) {

	authBytes, err := toml.Marshal(users)
	if err != nil {
		color.Errorln(err.Error())
		color.Errorln("Default authentication struct is not TOML conform")
		return users, err
	}
	err = os.WriteFile(authFilePath, authBytes, 0644)
	if err != nil {
		color.Errorln(err.Error())
		color.Warnln("Using default authentication without file")
	}
	return users, err
}
