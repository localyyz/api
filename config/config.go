package config

import (
	"os"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"

	"github.com/burntsushi/toml"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/sirupsen/logrus"
)

var (
	ErrNoConfig = errors.New("no configuration file")
)

type Config struct {
	Environment string `toml:"environment"`
	Bind        string `toml:"bind"`

	// [db]
	DB data.DBConf `toml:"db"`

	// [connect]
	Connect connect.Configs `toml:"connect"`

	// [pusher]
	Pusher struct {
		Topic string `tomp:"topic"`
	} `toml:"pusher"`

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
		return nil, errors.Wrap(err, "invalid config file given")
	}
	if _, err := toml.DecodeFile(file, &conf); err != nil {
		return nil, errors.Wrap(err, "unable to load config file")
	}

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	// If development, set lg to debug
	if conf.Environment == "development" {
		debugLv, _ := logrus.ParseLevel("debug")
		logger.SetLevel(debugLv)
	} else {
		// redirect all stdout to logrus logger
		lg.RedirectStdlogOutput(logger)
	}
	lg.DefaultLogger = logger

	return &conf, nil
}
