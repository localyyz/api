package presenter

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
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
		if cache, ok := ctx.Value("count.cache").(map[string]int); ok {
			d.ProductCount = cache[d.ID]
		} else {
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
	}

	{
		if productCache, ok := ctx.Value("product.cache").(map[string][]render.Renderer); ok {
			d.Products = productCache[d.ID]
		} else {
			// product preview
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

var splitRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")

func fetchDesignerProducts(ctx context.Context, query []string) []render.Renderer {
	genderCond := []data.ProductGender{
		data.ProductGenderUnisex,
		data.ProductGenderMale,
		data.ProductGenderFemale,
	}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		genderCond = []data.ProductGender{data.ProductGender(gender)}
	}

	// bulk fetch the preview products
	var p []*data.Product
	err := data.DB.Select("*").
		From(db.Raw(`(
				SELECT *, row_number() over (partition by lower(brand)) as n
				FROM products
				WHERE tsv @@ to_tsquery(?)
				AND deleted_at IS NULL
				AND gender IN ?
				) x`,
			strings.Join(query, "|"),
			genderCond,
		)).
		Where("n < 4").
		All(&p)

	if err != nil {
		return nil
	}
	return NewProductList(ctx, p)
}

func fetchDesignerCount(ctx context.Context, parsed []string) map[string]int {
	// bulk fetch product counts
	cond := db.Cond{"deleted_at": nil}
	if gender, ok := ctx.Value("session.gender").(data.UserGender); ok {
		cond["gender"] = gender
	}
	rows, err := data.DB.Select(
		db.Raw("lower(brand)"),
		db.Raw("count(1)")).
		From("products").
		Where(
			db.And(
				db.Raw(`tsv @@ to_tsquery(?)`, strings.Join(parsed, "|")),
				cond,
			),
		).
		GroupBy(db.Raw("lower(brand)")).
		Query()
	if err != nil {
		return nil
	}
	defer rows.Close()

	cache := make(map[string]int)
	for rows.Next() {
		var brand string
		var count int

		if err := rows.Scan(&brand, &count); err != nil {
			break
		}

		cache[brand] = count
	}

	return cache
}

func NewDesignerList(ctx context.Context, products []*data.Product) []render.Renderer {
	list := []render.Renderer{}

	brands := []string{}
	for _, p := range products {
		brands = append(brands, p.Brand)
	}
	// assemble the tsquery from designers
	var parsed []string
	for _, d := range brands {
		tt := splitRegex.Split(d, -1)
		parsed = append(parsed, strings.Join(tt, "&"))
	}

	productCache := make(map[string][]render.Renderer)
	for _, p := range fetchDesignerProducts(ctx, parsed) {
		pp := p.(*Product)
		b := strings.ToLower(pp.Brand)
		if _, ok := productCache[b]; ok {
			productCache[b] = []render.Renderer{}
		}
		productCache[b] = append(productCache[b], p)
	}
	lg.Warnf("%+v", productCache)
	ctx = context.WithValue(ctx, "product.cache", productCache)

	if cache := fetchDesignerCount(ctx, parsed); cache != nil {
		ctx = context.WithValue(ctx, "count.cache", cache)
	}

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
