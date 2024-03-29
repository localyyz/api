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
	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	xchange "bitbucket.org/moodie-app/moodie-api/lib/xchanger"
	"bitbucket.org/moodie-app/moodie-api/web"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
	pemFile  = flags.String("pem", "", "path to apns pem file")
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

	//[stash]
	st, err := stash.NewStash(conf.Stash.Host)
	if err != nil {
		if conf.Environment == "production" {
			lg.Fatal(errors.Wrap(err, "stash connection failed"))
		} else {
			lg.Warn(errors.Wrap(err, "stash connection failed"))
		}
	}
	_ = st

	// new web handler
	h := web.New(db)
	h.Debug = (conf.Environment == "development")

	//[connect]
	connect.Configure(conf.Connect)

	//[rates]
	go func() {
		xchange, err := xchange.New()
		if err != nil {
			lg.Fatalf("failed to load currency rates: %v", err)
		}
		lg.Infof("xchanger: loaded %d rates", len(xchange.Rates))
	}()

	//[jwt]
	token.SetupJWTAuth(conf.Jwt.Secret)

	// pusher
	if pemFile != nil && *pemFile != "" {
		if err := pusher.Setup(*pemFile, conf.Pusher.Topic, conf.Environment); err != nil {
			lg.Fatal(errors.Wrap(err, "invalid pem file"))
		}
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
	})

	lg.Warnf("API starting on %v", conf.Bind)

	if err := graceful.ListenAndServe(conf.Bind, h.Routes()); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
