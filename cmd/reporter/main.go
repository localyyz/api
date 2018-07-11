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
	"bitbucket.org/moodie-app/moodie-api/data/stash"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/forgett"
	"bitbucket.org/moodie-app/moodie-api/reporter"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("reporter", flag.ExitOnError)
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
	if _, err := data.NewDBSession(&conf.DB); err != nil {
		lg.Fatal(errors.Wrap(err, "database connection failed"))
	}

	//[connect]
	connect.SetupSlack(conf.Connect.Slack)
	nats := connect.SetupNatsStream(conf.Connect.Nats)

	//[stash]
	_, err = stash.NewStash(conf.Stash.Host)
	if err != nil {
		lg.Fatal(errors.Wrap(err, "stash connection failed"))
	}

	//[forgett]
	_, err = forgett.NewForgett(stash.NewPool(stash.NewDialer(conf.Stash.Host)))
	if err != nil {
		lg.Fatal(errors.Wrap(err, "forgett setup failed"))
	}

	// new handler
	h := reporter.New(nats)
	if nats != nil {
		// subscribe to all the nats streams
		h.Subscribe(conf.Connect.Nats)
	}

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

		if nats != nil {
			if !conf.Connect.Nats.Durable {
				nats.UnsubscribeAll()
			}
			nats.Close()
		}
	})

	lg.Warnf("Reporter starting on %v", conf.Bind)

	if err := graceful.ListenAndServe(conf.Bind, h.Routes()); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
