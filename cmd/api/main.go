package main

import (
	"flag"
	"os"
	"syscall"
	"time"

	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/pusher"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web"
	"bitbucket.org/moodie-app/moodie-api/workers"
	"github.com/goware/lg"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/zenazn/goji/graceful"
)

var (
	flags    = flag.NewFlagSet("api", flag.ExitOnError)
	confFile = flags.String("config", "", "path to config file")
	pemFile  = flags.String("pem", "", "path to apns pem file")
)

func main() {
	flags.Parse(os.Args[1:])

	// new handler context
	h := &web.Handler{}
	conf, err := config.NewFromFile(*confFile, os.Getenv("CONFIG"))
	if err != nil {
		lg.Fatal(err)
	}
	h.Debug = (conf.Environment == "development")

	//[db]
	if h.DB, err = data.NewDBSession(&conf.DB); err != nil {
		lg.Fatal(errors.Wrap(err, "database connection failed"))
	}

	//[connect]
	connect.Configure(conf.Connect)

	//[jwt]
	token.SetupJWTAuth(conf.Jwt.Secret)

	// pusher
	if pemFile != nil {
		if err := pusher.Setup(*pemFile, conf.Pusher.Topic, conf.Environment); err != nil {
			lg.Fatal(err)
		}
	}

	// cron worker
	c := cron.New()
	c.AddFunc("@every 1m", workers.PromoStartWorker)
	c.AddFunc("@every 1m", workers.PromoEndWorker)
	c.AddFunc("@every 1h", workers.ShopifyPuller)
	//c.AddFunc("0 0 0 * * *", worker.RefreshPromoWorker)
	c.Start()

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

	router := web.New(h)
	if err := graceful.ListenAndServe(conf.Bind, router); err != nil {
		lg.Fatal(err)
	}

	graceful.Wait()
}
