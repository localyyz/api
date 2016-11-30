package main

import (
	"log"
	"net/http"
	"os"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/tools/blogto/locale"

	_ "github.com/lib/pq"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	//"bitbucket.org/moodie-app/moodie-api/cmd/blogto/store"
)

var (
	app = kingpin.New("blogto", "Blogto api")

	appLocale     = app.Command("locale", "Toronto neightbourhoods")
	localeWriteDB = appLocale.Command("load", "Load the database with cached neighbourhoods.")
	localeList    = appLocale.Command("list", "List the neighbourhoods with their BlogTo Id.")
	localeShow    = appLocale.Command("show", "Show neighbourhoods in a web interface <port 8080>")

	appStore     = app.Command("store", "Toronto stores")
	storeLocale  = appStore.Flag("locale", "neightbourhood shorthand").Short('l').Required().String()
	storeListing = appStore.Command("list", "List store in a neighbourhood")
	storeLoad    = appStore.Command("load", "Load stores into the database. Open tunnel with 'ssh -L <port>:localhost:5432 -N ubuntu@moodie'")
	loadHost     = storeLoad.Flag("host", "Tunneled host string <host>:<port>.").Short('h').Default("localhost").String()
)

func main() {
	conf := &data.DBConf{
		Database: "localyyz",
		Hosts:    []string{*loadHost},
		Username: "localyyz",
	}
	if err := data.NewDBSession(conf); err != nil {
		log.Fatalf("db err: %s. Check ssh tunnel.", err)
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "locale show":
		http.HandleFunc("/", locale.LocaleHandler)
		log.Println("Listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	case "locale load":
		locale.LoadLocale()
	case "locale list":
		locale.List()
	case "store list":
		//if err := store.GetListing(*storeLocale); err != nil {
		//log.Fatal(err)
		//return
		//}
	case "store load":
		//if err := store.LoadListing(*storeLocale); err != nil {
		//log.Fatal(err)
		//return
		//}
	}
}
