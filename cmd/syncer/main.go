package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"syscall"
	"time"

	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/syncer"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
)

func main() {
	flags.Parse(os.Args[1:])

	conf, err := config.NewFromFile(*confFile, os.Getenv("CONFIG"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "invalid config file"))
	}

	// initialize seed
	rand.Seed(time.Now().Unix())

	//[db]
	db, err := data.NewDBSession(&conf.DB)
	if err != nil {
		lg.Fatal(errors.Wrap(err, "database connection failed"))
	}

	//[connect]
	connect.SetupSlack(conf.Connect.Slack)

	// new web handler
	h := syncer.New(db, (conf.Environment == "development"))

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

	lg.Warnf("Syncer starting on %v", conf.Bind)

	if err := graceful.ListenAndServe(conf.Bind, h.Routes()); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
