package main

import (
	"log"
	"os"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/moodie-app/moodie-api/cmd/blogto/locale"
	"bitbucket.org/moodie-app/moodie-api/cmd/blogto/store"
)

var (
	app = kingpin.New("blogto", "Blogto api")

	appLocale     = app.Command("locale", "Toronto neightbourhoods")
	localeWriteDB = appLocale.Command("load", "Load the database with cached neighbourhoods.")
	localeList    = appLocale.Command("list", "List the neighbourhoods with their BlogTo Id.")

	appStore     = app.Command("store", "Toronto stores")
	storeLocale  = app.Flag("locale", "neightbourhood shorthand").Short('l').Required().String()
	storeListing = appStore.Command("list", "List store in a neighbourhood")
	storeLoad    = appStore.Command("load", "Load stores into the database.")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "locale load":
		//doLoad()
	case "locale list":
		locale.List()
	case "store list":
		if err := store.GetListing(*storeLocale); err != nil {
			log.Fatal(err)
			return
		}
	case "store load":
		if err := store.LoadListing(*storeLocale); err != nil {
			log.Fatal(err)
			return
		}
	}
}
