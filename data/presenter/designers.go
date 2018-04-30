package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Designer struct {
	ID           string            `json:"id"`
	ImageURL     string            `json:"imageUrl"`
	ProductCount int               `json:"productCount"`
	Products     []render.Renderer `json:"products"`
}

type CountCache map[int64][]int

var (
	logoCache = map[string]string{
		"calvin klein":    "https://cdn.shopify.com/s/files/1/1976/6885/files/ck.jpg?16038229841572028414",
		"converse":        "https://cdn.shopify.com/s/files/1/1976/6885/files/converse.jpg?16038229841572028414",
		"the north face":  "https://cdn.shopify.com/s/files/1/1976/6885/files/north_face.jpg?16038229841572028414",
		"puma":            "https://cdn.shopify.com/s/files/1/1976/6885/files/puma.jpg?16038229841572028414",
		"timberland":      "https://cdn.shopify.com/s/files/1/1976/6885/files/timberland.jpg?16038229841572028414",
		"nike":            "https://cdn.shopify.com/s/files/1/1976/6885/files/nike.jpg?16038229841572028414",
		"michael kors":    "https://cdn.shopify.com/s/files/1/1976/6885/files/michael_kors.jpg?16038229841572028414",
		"prada":           "https://cdn.shopify.com/s/files/1/1976/6885/files/daew_8.jpg?16038229841572028414",
		"asics":           "https://cdn.shopify.com/s/files/1/1976/6885/files/asics.png?16038229841572028414",
		"bottega veneta":  "https://cdn.shopify.com/s/files/1/1976/6885/files/bottega_veneta.jpg?16038229841572028414",
		"gucci":           "https://cdn.shopify.com/s/files/1/1976/6885/files/gucci_logo.jpg?16038229841572028414",
		"valentino":       "https://cdn.shopify.com/s/files/1/1976/6885/files/valentino.jpg?1782811635631612473",
		"tissot":          "https://cdn.shopify.com/s/files/1/1976/6885/files/tissot.jpg?1782811635631612473",
		"coach":           "https://cdn.shopify.com/s/files/1/1976/6885/files/coachlogo.png?1782811635631612473",
		"levi's":          "https://cdn.shopify.com/s/files/1/1976/6885/files/levi_s.jpg?1782811635631612473",
		"adidas":          "https://cdn.shopify.com/s/files/1/1976/6885/files/adidas.jpg?1782811635631612473",
		"tommy hilfiger":  "https://cdn.shopify.com/s/files/1/1976/6885/files/tommy_hilfiger.jpg?1782811635631612473",
		"louis vuitton":   "https://cdn.shopify.com/s/files/1/1976/6885/files/louis_vuitton.jpg?1782811635631612473",
		"dolce & gabbana": "https://cdn.shopify.com/s/files/1/1976/6885/files/dolce.png?16613022696514256589",
		"diesel":          "https://cdn.shopify.com/s/files/1/1976/6885/files/diesel.png?10642812083174478034",
		"versace":         "https://cdn.shopify.com/s/files/1/1976/6885/files/versace.png?13325662696205199185",
		"guess":           "https://cdn.shopify.com/s/files/1/1976/6885/files/guess.png?10642812083174478034",
	}
)

func NewDesigner(ctx context.Context, product *data.Product) *Designer {
	d := &Designer{
		ID: product.Brand,
	}

	{ // product count
		cond := db.Cond{"deleted_at": nil}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["gender"] = gender
		}
		count, _ := data.DB.Product.Find(
			db.And(
				db.Raw(`tsv @@ plainto_tsquery(?)`, d.ID),
				cond,
			),
		).Count()
		d.ProductCount = int(count)
	}

	{ // product preview
		cond := db.Cond{"p.deleted_at": nil}
		if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
			cond["p.gender"] = gender
		}

		var products []*data.Product
		data.DB.Select("p.*").
			From("products p").
			Where(
				db.And(
					db.Raw(`tsv @@ plainto_tsquery(?)`, d.ID),
					cond,
				),
			).
			Limit(10).
			All(&products)

		if pp := NewProductList(ctx, products); len(pp) > 4 {
			d.Products = pp[:4]
		} else {
			d.Products = pp
		}
	}

	return d
}

func (d *Designer) Render(w http.ResponseWriter, r *http.Request) error {
	if img, ok := logoCache[d.ID]; ok {
		d.ImageURL = img
	}
	return nil
}

type DesignerList []*Designer

func NewDesignerList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}
	for _, p := range products {
		list = append(list, NewDesigner(ctx, p))
	}
	return list
}

func (l DesignerList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}
