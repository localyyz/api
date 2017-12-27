package merchant

import "github.com/flosch/pongo2"

var (
	indexTmpl = pongo2.Must(pongo2.FromFile("/merchant/index.html"))
)
