package presenter

import (
	"context"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type Category struct {
	Type     string      `json:"type"`
	Values   []*Category `json:"values"`
	ImageURL string      `json:"imageUrl"`

	ctx context.Context
}

type subCatCache map[string][]*data.Category

func fetchSubcategory(ctx context.Context, categoryTypes ...data.CategoryType) []*data.Category {
	cond := db.Cond{
		// for now, just fetch the 2ndary categories
		"mapping": db.NotEq(""),
		"type":    categoryTypes,
	}
	// fetch based on user's gender if available
	if sessionUser, ok := ctx.Value("session.user").(*data.User); ok {
		if sessionUser.Etc.Gender == data.UserGenderMale {
			cond["gender"] = data.ProductGenderMale
		} else if sessionUser.Etc.Gender == data.UserGenderFemale {
			cond["gender"] = data.ProductGenderFemale
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
