package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type Category struct {
	Type        string      `json:"type"`
	Values      []*Category `json:"values"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	ImageURL    string      `json:"imageUrl"`

	ctx context.Context
}

type subCatCache map[string][]*data.Category

func fetchSubcategory(ctx context.Context, categoryTypes ...data.CategoryType) []*data.Category {
	cond := db.Cond{
		// for now, just fetch the 2ndary categories
		"mapping": db.NotEq(""),
		"type":    categoryTypes,
	}

	if filterSort, ok := ctx.Value(api.FilterSortCtxKey).(*api.FilterSort); ok {
		// TODO: pull this out into the filtersort api
		for _, f := range filterSort.Filters {
			if f.Type == "gender" {
				v := new(data.ProductGender)
				if err := v.UnmarshalText([]byte(f.Value.(string))); err == nil {
					cond["gender"] = []data.ProductGender{
						*v,
						data.ProductGenderUnisex,
					}
				}
				break
			}
		}
	}

	// bulk fetch product categories
	// this is the 2nd level mapping
	//
	// category type -> mapping -> category values
	rows, err := data.DB.
		Select(db.Raw("distinct mapping"), "type", "image_url").
		From("product_categories").
		Where(cond).
		OrderBy("mapping").
		Query()
	if err != nil {
		lg.Warn(err)
		return nil
	}

	var subcategories []*data.Category
	for rows.Next() {
		var mapping string
		var typ data.CategoryType
		var imgUrl string
		if err := rows.Scan(&mapping, &typ, &imgUrl); err != nil {
			break
		}
		subcategories = append(
			subcategories,
			&data.Category{
				Type:     typ,
				Mapping:  mapping,
				ImageURL: imgUrl,
			},
		)
	}

	return subcategories
}

func NewCategory(ctx context.Context, categoryType data.CategoryType) *Category {
	category := &Category{
		Type: categoryType.String(),
		ctx:  ctx,
	}
	var values []*data.Category
	if subcats, ok := ctx.Value("subcat").(subCatCache); ok {
		values = subcats[categoryType.String()]
	} else {
		values = fetchSubcategory(ctx, categoryType)
	}

	for _, v := range values {
		category.Values = append(category.Values, &Category{
			Type:     v.Mapping,
			ImageURL: v.ImageURL,
		})
	}
	return category
}

func (c *Category) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type CategoryList []*Category

func NewCategoryList(ctx context.Context, categoryTypes []data.CategoryType) []render.Renderer {
	list := []render.Renderer{}

	// bulk fetch subcategories
	subcatMap := make(subCatCache)
	for _, c := range fetchSubcategory(ctx, categoryTypes...) {
		subcatMap[c.Type.String()] = append(subcatMap[c.Type.String()], c)
	}

	ctx = context.WithValue(ctx, "subcat", subcatMap)
	for _, c := range categoryTypes {
		var presented *Category
		switch c {
		case data.CategorySale:
			presented = &Category{
				Type: "sales",
				Values: []*Category{
					{
						Type:     "70% OFF",
						Title:    "70%+ OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/70.png?17957505310432019141",
					},
					{
						Type:     "50% OFF",
						Title:    "50%-70% OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/50.png?5115785919598170614",
					},
					{
						Type:     "20% OFF",
						Title:    "20%-50% OFF",
						ImageURL: "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/20.png?14969378164451378728",
					},
				},
			}
		case data.CategoryCollection:
			presented = &Category{
				Type: "collections",
				Values: []*Category{
					{
						Type:     "smart",
						Title:    "Under $50",
						ImageURL: "https://cdn.shopify.com/s/files/1/0835/3729/products/Oversized_Hoodies_-4_eda921cf-882d-479f-8d07-ed1c070b0a0a.jpg",
					},
					{
						Type:     "boutique",
						Title:    "$50 - $200",
						ImageURL: "https://cdn.shopify.com/s/files/1/1066/9348/products/UNG85206_red_0.jpg",
					},
					{
						Type:     "designer",
						Title:    "$200 plus",
						ImageURL: "https://cdn.shopify.com/s/files/1/0444/7969/products/mens-jackets-coats-hexagon-stitch-brother-jacket-1.jpg",
					},
				},
			}
		default:
			presented = NewCategory(ctx, c)
		}
		if len(presented.Values) > 0 {
			list = append(list, presented)
		}
	}
	return list
}

func (l CategoryList) Render(w http.ResponseWriter, r *http.Request) error {
	for _, v := range l {
		if err := v.Render(w, r); err != nil {
			return err
		}
	}
	return nil
}
