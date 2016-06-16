package main

import (
	"flag"
	"os"
	"syscall"
	"time"

	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"github.com/goware/lg"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	conf, err := config.NewFromFile(*confFile, os.Getenv("CONFIG"))
	if err != nil {
		lg.Fatal(err)
	}

	//[db]
	if err := data.NewDBSession(conf.DB); err != nil {
		lg.Fatal(err)
	}

	//[connect]
	connect.Configure(conf.Connect)

	//[jwt]
	data.SetupJWTAuth(conf.Jwt.Secret)

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

	lg.Infof("API starting on %v", conf.Bind)

	router := web.New()
	if err := graceful.ListenAndServe(conf.Bind, router); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
