package tests

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"bitbucket.org/liamstask/goose/lib/goose"
	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/token"
	"bitbucket.org/moodie-app/moodie-api/web"
	"github.com/stretchr/testify/require"
)

type Env struct {
	T          *testing.T
	DB         *data.Database
	URL        string
	HttpServer *httptest.Server
	Debug      bool
}

// Close must be called after each test to ensure the Env is properly destroyed.
func (env *Env) Close() {
	env.HttpServer.Close()
	env.DB.Close()
}

func SetupEnv(t *testing.T) *Env {
	config, err := config.NewFromFile("", os.Getenv("CONFIG"))
	require.NoError(t, err, "CONFIG env must be set")

	// initiate database
	cmd := exec.Command("/bin/sh", os.Getenv("DBSCRIPTS"), "reset", config.DB.Database)
	require.NoError(t, cmd.Run(), "DB failed to reset")

	// start migration
	mdir := os.Getenv("MIGRATIONDIR")
	gooseConf, err := goose.NewDBConf(mdir, "testing", "")
	require.NoError(t, err, "DB failed to setup goose")

	target, err := goose.GetMostRecentDBVersion(mdir)
	require.NoError(t, err, "DB failed to find goose target")

	require.NoError(t, goose.RunMigrations(gooseConf, mdir, target), "DB failed to migrate")

	// initialize database session
	db, err := data.NewDBSession(&config.DB)
	require.NoError(t, err, "DB failed to connect")

	// TODO: other initialization
	//  connect shopify/slack/facebook etc

	// stripe
	connect.SetupStripe(config.Connect.Stripe)

	// jwt
	token.SetupJWTAuth(config.Jwt.Secret)

	w := web.New(db)
	h := httptest.NewServer(w.Routes())

	return &Env{
		DB:         db,
		URL:        h.URL,
		HttpServer: h,
		Debug:      true,
	}
}
