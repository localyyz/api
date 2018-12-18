package product

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
	db "upper.io/db.v3"
)

var (
	MaxRowNum    = 14
	MaxFavRowNum = 3

	IgnoreCategory = []string{
		"bodysuit",
		"boxer-brief",
		"boxer-trunk",
		"bra",
		"deodorant",
		"lingerie",
		"loungewear",
		"thong",
	}
)

func ListFeedV3Products(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	query := data.DB.
		Select("p.*").
		From("products p").
		Where(db.Cond{
			"p.status": data.ProductStatusApproved,
		}).
		OrderBy(
			db.Raw("row_number() over (partition by place_id)"),
			"-p.score",
			"-p.id",
		)
	query = filterSort.UpdateQueryBuilder(query)

	if filterSort.HasFilter() {
		w.Write([]byte{})
		return
	}

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}

}

func ListFeedV3Onsale(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)
	cursor := ctx.Value("cursor").(*api.Page)
	filterSort := ctx.Value("filter.sort").(*api.FilterSort)

	preferedPlaceIDs, _ := data.DB.PlaceMeta.GetPlacesFromPreference(user.Preference)
	preferedGender := user.GetPreferredGenders()
	cond := db.Cond{
		"p.status":                     data.ProductStatusApproved,
		"p.category_id":                db.IsNotNull(),
		"p.discount_pct":               db.Gte(0.5),
		"p.place_id":                   preferedPlaceIDs,
		"p.gender":                     preferedGender,
		db.Raw("p.category->>'value'"): db.NotIn(IgnoreCategory),
	}
	query := data.DB.
		Select("p.*").
		From("products p").
		Where(cond).
		OrderBy(
			"-p.score",
			"-p.id",
		)
	query = filterSort.UpdateQueryBuilder(query)

	if filterSort.HasFilter() {
		w.Write([]byte{})
		return
	}

	var products []*data.Product
	paginate := cursor.UpdateQueryBuilder(query)
	if err := paginate.All(&products); err != nil {
		render.Respond(w, r, err)
		return
	}
	cursor.Update(products)

	presented := presenter.NewProductList(ctx, products)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func ListFeedV3(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	if user.Preference == nil {
		render.Respond(w, r, []struct{}{})
		return
	}

	var (
		rows      []*data.Feed
		onSaleRow = &data.Feed{
			Products: []*data.Product{},
		}
		favouritesRow = &data.Feed{
			Title:     "Your favourite products",
			FetchPath: "products/favourite",
			Type:      data.FeedTypeFavourite,
			Order:     2,
			Products:  []*data.Product{},
		}
	)

	preferedGender := user.GetPreferredGenders()
	preferedPlaceIDs, _ := data.DB.PlaceMeta.GetPlacesFromPreference(user.Preference)
	{ // all items on sale -> greater than 50%
		query := data.DB.Product.
			Find(db.Cond{
				"status":                     data.ProductStatusApproved,
				"category_id":                db.IsNotNull(),
				"discount_pct":               db.Gte(0.5),
				"place_id":                   preferedPlaceIDs,
				"gender":                     preferedGender,
				db.Raw("category->>'value'"): db.NotIn(IgnoreCategory),
			}).
			OrderBy(db.Raw("random()")).
			Limit(10)
		var products []*data.Product
		query.All(&products)
		if len(products) > 0 {
			onSaleRow = &data.Feed{
				Title:     "Products on sale you may like.",
				Products:  products,
				FetchPath: "products/feedv3/onsale",
				Order:     1,
			}
		}
	}

	{
		// favourites
		favProductIDs, _ := data.DB.FavouriteProduct.GetProductIDsFromUser(user.ID)
		if len(favProductIDs) > 0 {
			var products []*data.Product
			data.DB.Select("*").
				From("products").
				Where(db.Cond{
					"status":      data.ProductStatusApproved,
					"category_id": db.IsNotNull(),
					"id":          favProductIDs,
				}).
				Limit(10).
				All(&products)

			favouritesRow.Products = products

			for i, product := range products {
				if i > MaxFavRowNum {
					break
				}

				// find some related products
				var related []*data.Product
				data.DB.Select("*").
					From("products").
					Where(
						db.Cond{
							"place_id":    preferedPlaceIDs,
							"gender":      product.Gender,
							"id":          db.NotEq(product.ID),
							"status":      data.ProductStatusApproved,
							"category_id": product.CategoryID,
							"price":       db.Between(product.Price-10, product.Price+10),
							db.Raw("category->>'value'"): product.Category.Value,
						}).
					OrderBy("score desc", "id desc").
					Limit(10).
					All(&related)

				if len(related) > 0 {
					rows = append(rows,
						&data.Feed{
							Title:     product.Title,
							Type:      data.FeedTypeRelated,
							Products:  related,
							FetchPath: fmt.Sprintf("products/%d/related", product.ID),
						})
				}
			}
		}
	}

	var preferences []data.Preference
	// make a matrix of user preferences
	for _, gd := range user.Preference.Gender {
		var categoryIDs []int64
		if gd == "man" {
			categoryIDs = []int64{
				// NOTE: break up apparel into sub categories
				11020, // activewear
				11040, // blazer
				11060, // coatjacket
				11080, // jeans
				11120, // shirt
				11160, // sweatshirt
				11180, // suit
				11200, // sweater
				11220, // tshirt

				12000,
				13000,
				14000,
				15000,
			}
		} else {
			categoryIDs = []int64{
				// NOTE: break up apparel into sub categories
				21020, // activewear
				21040, // coatjacket
				21060, // dress
				21100, // jeans
				21120, // jumpsuit
				21180, // pants
				21260, // sweater
				21280, // top

				22000,
				23000,
				24000,
				25000,
				26000,
				27000,
			}
		}
		// for every gender
		for _, pr := range user.Preference.Pricings {
			// for every pricing
			for _, st := range user.Preference.Styles {
				for _, cID := range categoryIDs {
					// for every style
					p := data.Preference{
						Gender:     gd,
						Pricing:    pr,
						Style:      st,
						CategoryID: cID,
					}
					preferences = append(preferences, p)
				}

			}
		}
	}

	// permutate + randomize the preferences slice
	for i := range preferences {
		j := rand.Intn(i + 1)
		preferences[i], preferences[j] = preferences[j], preferences[i]
	}

	selectedCategorySet := set.New()
	for _, prf := range preferences {
		s := time.Now()
		if len(rows) > MaxRowNum {
			break
		}

		// we've already selected the same category
		if selectedCategorySet.Has(prf.CategoryID) {
			continue
		}

		// get the products
		styleCol := "style_female"
		if prf.Gender == "man" {
			styleCol = "style_male"
		}

		var meta []data.PlaceMeta
		data.DB.PlaceMeta.
			Find(
				db.And(
					db.Or(
						db.Cond{"gender": prf.Gender},
						db.Cond{"gender": db.IsNull()},
					),
					db.Cond{
						"pricing": prf.Pricing,
						styleCol:  prf.Style,
					},
				),
			).
			All(&meta)
		var placeIDs []int64
		for _, p := range meta {
			placeIDs = append(placeIDs, p.PlaceID)
		}

		if len(placeIDs) == 0 {
			continue
		}

		gender := new(data.ProductGender)
		gender.UnmarshalText([]byte(prf.Gender))
		cond := db.Cond{
			"status": data.ProductStatusApproved,
			"gender": []data.ProductGender{
				*gender,
				data.ProductGenderUnisex,
			},
			"place_id": placeIDs,
		}

		switch prf.Pricing {
		case "low":
			cond["price"] = db.Lte(50)
		case "medium":
			cond["price"] = db.Between(50, 150)
		case "high":
			cond["price"] = db.Gte(150)
		}

		category, _ := data.DB.Category.FindByID(prf.CategoryID)
		descendantIDs, _ := data.DB.Category.FindDescendantIDs(category.ID)
		catCond := db.And(
			cond,
			db.Cond{
				"category_id": append(descendantIDs, category.ID),
			},
		)

		query := data.DB.Product.
			Find(db.And(catCond, db.Cond{"discount_pct": 0})).
			OrderBy(db.Raw("row_number() over (partition by category_id)")).
			Limit(10)
		if count, _ := query.Count(); count < 10 {
			lg.Debugf("skipping %s because count is %d", category.Label, count)
			// not enough products. skip
			continue
		}

		var products []*data.Product
		if err := query.All(&products); err != nil {
			lg.Debugf("skipping %s err %v", category.Label, err)
			continue
		}

		selectedCategorySet.Add(prf.CategoryID)
		rows = append(rows, &data.Feed{
			Preference: prf,
			Category:   category,
			Type:       data.FeedTypeRecommend,
			FetchPath:  "/products/feedv3/products",
			Products:   products,
		})

		// sale cat cond
		saleCatCond := db.And(catCond, db.Cond{"discount_pct": db.Gte(0.25)})
		var saleProducts []*data.Product
		data.DB.Product.
			Find(saleCatCond).
			OrderBy("-discount_pct", db.Raw("row_number() over (partition by category_id)")).
			Limit(5).
			All(&saleProducts)

		if len(saleProducts) == 5 {
			rows = append(rows, &data.Feed{
				Preference: prf,
				Category:   category,
				Type:       data.FeedTypeSale,
				FetchPath:  "/products/feedv3/products",
				Products:   saleProducts,
			})
		}

		lg.Debugf("iteration done: %s", time.Since(s))
	}

	// permutate + randomize the row slice
	for i := range rows {
		j := rand.Intn(i + 1)
		rows[i], rows[j] = rows[j], rows[i]
	}
	rows = append([]*data.Feed{favouritesRow}, rows...)
	if len(onSaleRow.Products) != 0 {
		rows = append([]*data.Feed{onSaleRow}, rows...)
	}

	presented := presenter.NewFeedList(ctx, rows)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}
