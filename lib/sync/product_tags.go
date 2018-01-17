package sync

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/gedex/inflector"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	set "gopkg.in/fatih/set.v0"
)

type aggregateCategory struct {
	gender     data.ProductGender
	categories []*data.ProductCategory
}

func (s *aggregateCategory) Len() int {
	return len(s.categories)
}

func (s *aggregateCategory) Swap(i, j int) {
	s.categories[i], s.categories[j] = s.categories[j], s.categories[i]
}

func (s *aggregateCategory) Less(i, j int) bool {
	return s.categories[i].Weight > s.categories[j].Weight
}

func ShopifyProductTagsCreate(ctx context.Context, product *data.Product, p *shopify.ProductList) error {
	place := ctx.Value("sync.place").(*data.Place)

	// bulk parse and insert product tags
	q := data.DB.InsertInto("product_tags").
		Columns("product_id", "place_id", "value", "type").
		Amend(func(query string) string {
			return query + ` ON CONFLICT DO NOTHING`
		})
	b := q.Batch(5)
	go func() {
		defer b.Done()

		// Flag if gender was found in any of the given fields
		parsedCategories := []aggregateCategory{
			parseCategory(ctx, p.Title),
			parseCategory(ctx, p.Tags),
			parseCategory(ctx, p.ProductType),
		}
		f := aggregateCategory{}

		for _, a := range parsedCategories {
			// flatten gender
			if f.gender == data.ProductGenderUnknown &&
				(a.gender == data.ProductGenderMale || a.gender == data.ProductGenderFemale) {
				f.gender = a.gender
			}
		}
		for _, a := range parsedCategories {
			// flatten category, but throw out ones that doesn't match
			// products gender
			for _, fc := range a.categories {
				if f.gender != data.ProductGenderUnknown &&
					fc.Gender != data.ProductGenderUnisex &&
					fc.Gender != f.gender {
					continue
				}
				f.categories = append(f.categories, fc)
			}
		}

		sort.Sort(&f)

		if f.gender != data.ProductGenderUnknown &&
			product.Gender == data.ProductGenderUnisex {
			// if the product gender is unisex (inherited from place),
			// use the hinted category genders found both category and gender
			product.Gender = f.gender
			data.DB.Product.Save(product)
		}

		if len(f.categories) == 0 {
			lg.Warnf("parseTags: pl(%s) pr(%s) g(%s) without category", place.Name, p.Handle, f.gender)
		} else {
			// use the parsed out category (sorted with the highest weighting) and insert as value
			b.Values(product.ID, place.ID, f.categories[0].Value, data.ProductTagTypeCategory)
		}

		// Product Vendor/Brand
		b.Values(product.ID, place.ID, strings.ToLower(p.Vendor), data.ProductTagTypeBrand)
	}()
	if err := b.Wait(); err != nil {
		return errors.Wrap(err, "failed to create product tags")
	}

	return nil
}

func parseGender(t string) data.ProductGender {
	if t == "man" || t == "gentleman" {
		return data.ProductGenderMale
	}
	if t == "woman" || t == "lady" {
		return data.ProductGenderFemale
	}
	if t == "kid" {
		return data.ProductGenderUnisex
	}
	return data.ProductGenderUnknown
}

func parseCategory(ctx context.Context, s string) aggregateCategory {
	tags := ParseTags(s)
	categoryCache := ctx.Value("category.cache").(map[string]*data.ProductCategory)
	place := ctx.Value("sync.place").(*data.Place)

	gender := data.ProductGenderUnknown
	bestCat := aggregateCategory{}
	for _, t := range tags {
		// if we got a gender hint in the product -> check detected category and
		// see if we can pull up something that matches the gender
		gender = parseGender(t)
		if gender != data.ProductGenderUnknown {
			bestCat.gender = gender
			break
		}
	}

	for _, t := range tags {
		cat, found := categoryCache[t]
		if !found {
			continue
		}
		if place.Gender != data.PlaceGender(data.ProductGenderUnisex) &&
			cat.Gender != data.ProductGender(place.Gender) {
			// if place has a gender baise, only pick the categories
			// with the same gender
			continue
		}

		// if found gender in category, set it as such
		if cat.Gender != data.ProductGenderUnisex &&
			gender == data.ProductGenderUnknown {
			bestCat.gender = cat.Gender
		}
		bestCat.categories = append(bestCat.categories, cat)
	}

	return bestCat
}

var (
	tagRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")
	// TODO: load up tag specials from product categories
	tagSpecial = []string{
		"bomber-jacket",
		"boxer-brief",
		"boxer-trunk",
		"eau-de-toilette",
		"eau-de-parfum",
		"pocket-square",
		"shoulder-bag",
		"sleeping-bag",
		"t-shirt",
		"track-pant",
		"v-neck",
		"air-jordan",
		"lace-up",
		"slip-on",
	}
)

func hasNoLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func ParseTags(tagStr string, optTags ...string) []string {
	tagStr = strings.ToLower(tagStr)
	tt := tagRegex.Split(tagStr, -1)
	tagSet := set.New()

	slugTagStr := slug.Make(strings.Join(tt, " "))
	for _, t := range tagSpecial {
		if strings.Index(slugTagStr, t) != -1 {
			tagSet.Add(t)
		}
	}

	// if tagSet at this point is not empty, found a special case
	// return right away
	if tagSet.Size() > 0 {
		return set.StringSlice(tagSet)
	}

	for _, t := range tt {
		if hasNoLetter(t) {
			// skip if not contain any alphanum letter
			continue
		}
		tt := inflector.Singularize(t)
		for {
			if tt == t {
				break
			}
			t = tt
			tt = inflector.Singularize(t)
		}
		if t == "" {
			continue
		}
		tagSet.Add(t)
	}
	for _, t := range optTags {
		tagSet.Add(t)
	}

	return set.StringSlice(tagSet)
}
