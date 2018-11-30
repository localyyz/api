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
	"bitbucket.org/moodie-app/moodie-api/lib/sync"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
	"bitbucket.org/moodie-app/moodie-api/syncer"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("syncer", flag.ExitOnError)
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
	connect.SetupNatsStream(conf.Connect.Nats)
	connect.SetupZapier(conf.Connect.Zapier)

	// new web handler
	h := syncer.New(db)
	h.Debug = (conf.Environment == "development")

	// cache
	if err := sync.SetupCache(); err != nil {
		lg.Fatal(err)
	}

	//[rates]
	go func() {
		xchange, err := xchange.New()
		if err != nil {
			lg.Fatalf("failed to load currency rates: %v", err)
		}
		lg.Infof("xchanger: loaded %d rates", len(xchange.Rates))
	}()

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
