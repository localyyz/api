package main

import (
	"flag"
	"log"
	"os"
	"syscall"
	"time"

	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/merchant"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("merchant", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	conf, err := config.NewFromFile(*confFile, os.Getenv("CONFIG"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "invalid config file"))
	}

	//[db]
	if _, err = data.NewDBSession(&conf.DB); err != nil {
		lg.Fatal(errors.Wrap(err, "database connection failed"))
	}

	//[jwt]
	token.SetupJWTAuth(conf.Jwt.Secret)

	//[connect]
	connect.Configure(conf.Connect)

	graceful.AddSignal(syscall.SIGINT, syscall.SIGTERM)
	graceful.Timeout(10 * time.Second) // Wait timeout for handlers to finish.
	graceful.PreHook(func() {
		lg.Info("waiting for requests to finish..")
	})
	graceful.PostHook(func() {
		lg.Info("finishing up...")
		if err := data.DB.Close(); err != nil {
			lg.Alert(err)
		}
	})

	lg.Infof("Merchant site starting on %v", conf.Bind)
	router := merchant.New(
		&merchant.Handler{
			Debug:       (conf.Environment == "development"),
			SH:          connect.SH,
			SL:          connect.SL,
			ApiURL:      conf.Api.ApiURL,
			Environment: conf.Environment,
		},
	)
	if err := graceful.ListenAndServe(conf.Bind, router); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
