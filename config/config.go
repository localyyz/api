package config

import (
	"errors"
	"os"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"

	"github.com/BurntSushi/toml"
)

var (
	ErrNoConfig = errors.New("no configuration file")
)

type Config struct {
	Bind string `toml:"bind"`

	// [db]
	DB data.DBConf `toml:"db"`

	// [map]
	Map struct { // google maps api key
		ApiKey string `toml:"api_key"`
	} `toml:"map"`

	// [connect]
	Connect connect.Configs `toml:"connect"`

	// [jwt]
	Jwt struct {
		Secret string `toml:"secret"`
	} `toml:"jwt"`
}

func NewFromFile(fileConfig, envConfig string) (*Config, error) {

	file := fileConfig
	if file == "" {
		file = envConfig
	}

	var conf Config
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
