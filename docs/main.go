package main

import (
	"encoding/json"
	"log"
	"os"

	"bitbucket.org/moodie-app/moodie-api/web"
	"github.com/go-chi/docgen"
)

func main() {
	routes := web.New(&web.Handler{})

	// mkdown output
	mdDocFile, err := os.Create("./docs/routes.md")
	if err != nil {
		log.Fatal(err)
	}
	mdDoc := &docgen.MarkdownDoc{Router: routes}
	if err := mdDoc.Generate(); err != nil {
		log.Fatal(err)
	}
	mdDocFile.WriteString(mdDoc.String())
	mdDocFile.Close()

	// json output
	jsDocFile, err := os.Create("./docs/routes.json")
	if err != nil {
		log.Fatal(err)
	}
	encoder := json.NewEncoder(jsDocFile)
	encoder.SetIndent("", "  ")
	encoder.Encode(&mdDoc.Doc)

	jsDocFile.Close()
}
