package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	db "upper.io/db.v3"
)

type Category struct {
	Type   data.CategoryType `json:"type"`
	Values []string          `json:"values"`

	ctx context.Context
}

type subCatCache map[data.CategoryType][]string

func NewCategory(ctx context.Context, categoryType data.CategoryType) *Category {
	category := &Category{
		Type: categoryType,
		ctx:  ctx,
	}
	if subcats, ok := ctx.Value("subcat").(subCatCache); ok {
		category.Values = subcats[categoryType]
	} else {
		cond := db.Cond{
			"type":       categoryType,
			"mapping !=": "",
		}
		var categories []*data.Category
		data.DB.
			Select(db.Raw("distinct mapping")).
			From("product_categories").
			Where(cond).
			OrderBy("mapping").
			All(&categories)

		category.Values = []string{}
		for _, c := range categories {
			category.Values = append(category.Values, c.Mapping)
		}

	}
	return category
}

func (c *Category) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

type CategoryList []*Category

func NewCategoryList(ctx context.Context, categories []data.CategoryType) []render.Renderer {
	list := []render.Renderer{}

	cond := db.Cond{
		"mapping": db.NotEq(""),
		"type":    categories,
	}
	if sessionUser, ok := ctx.Value("session.user").(*data.User); ok {
		if sessionUser.Etc.Gender == data.UserGenderMale {
			cond["gender"] = data.ProductGenderMale
		} else if sessionUser.Etc.Gender == data.UserGenderFemale {
			cond["gender"] = data.ProductGenderFemale
		}
	}
	// bulk fetch product categories
	rows, _ := data.DB.
		Select(db.Raw("distinct mapping"), "type").
		From("product_categories").
		Where(cond).
		OrderBy("mapping").
		Query()

	subcatMap := subCatCache{}
	for rows.Next() {
		var mapping string
		var typ data.CategoryType
		if err := rows.Scan(&mapping, &typ); err != nil {
			break
		}
		subcatMap[typ] = append(subcatMap[typ], mapping)
	}
	ctx = context.WithValue(ctx, "subcat", subcatMap)

	for _, c := range categories {
		list = append(list, NewCategory(ctx, c))
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
