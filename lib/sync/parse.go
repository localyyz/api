package sync

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/gedex/inflector"
	"github.com/gosimple/slug"
	set "gopkg.in/fatih/set.v0"
)

const cacheKey = `category.cache`
const cacheKeyBlacklist = `category.blacklist`

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

type aggregateCategory []*data.Category

func (s aggregateCategory) Len() int {
	return len(s)
}

func (s aggregateCategory) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s aggregateCategory) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

func hasNoLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func tokenize(tagStr string, optTags ...string) []string {
	tagStr = strings.ToLower(tagStr)
	tt := tagRegex.Split(tagStr, -1)
	tagSet := set.New()

	slugTagStr := slug.Make(strings.Join(tt, " "))
	for _, t := range tagSpecial {
		if strings.Index(slugTagStr, t) != -1 {
			tagSet.Add(t)
		}
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

// filter tags for categories
func parseCategory(ctx context.Context, tokens []string) []string {
	categoryCache := ctx.Value(cacheKey).(map[string]*data.Category)
	var categories []string
	for _, t := range tokens {
		if _, found := categoryCache[t]; found {
			categories = append(categories, t)
		}
	}
	return categories
}

// filter tags for genders
func parseGender(ctx context.Context, tokens []string) data.ProductGender {
	var maybeGender data.ProductGender
	for _, t := range tokens {
		switch t {
		case "man", "male", "gentleman":
			return data.ProductGenderMale
		case "woman", "female", "lady", "sexy":
			maybeGender = data.ProductGenderFemale
			return data.ProductGenderFemale
		case "kid":
			return data.ProductGenderUnisex
		}
	}

	if maybeGender != data.ProductGenderUnknown {
		return maybeGender
	}

	return data.ProductGenderUnisex
}

func ParseProduct(ctx context.Context, inputs ...string) data.Category {
	categoryCache, _ := ctx.Value(cacheKey).(map[string]*data.Category)
	if categoryCache == nil {
		// if by chance category cache is not given, generate it here
		if categories, _ := data.DB.Category.FindAll(nil); categories != nil {
			categoryCache = make(map[string]*data.Category, len(categories))
			for _, c := range categories {
				categoryCache[c.Value] = c
			}
		}
		ctx = context.WithValue(ctx, cacheKey, categoryCache)
	}
	place := ctx.Value("sync.place").(*data.Place)

	var (
		parsed     = data.Category{Gender: data.ProductGender(place.Gender)}
		categories = set.New()
	)
	for _, s := range inputs {
		tokens := tokenize(s)
		// parse tokens for gender hint
		if parsed.Gender == data.ProductGenderUnisex {
			parsed.Gender = parseGender(ctx, tokens)
		}
		for _, c := range parseCategory(ctx, tokens) {
			categories.Add(c)
		}
	}

	aggregates := make(aggregateCategory, categories.Size())
	for i, s := range set.StringSlice(categories) {
		aggregates[i] = categoryCache[s]
	}
	// sort categories by weight
	sort.Sort(aggregates)

	if len(aggregates) > 0 {
		// use the parsed out category (sorted with the highest weighting) and insert as value
		parsed.Value = aggregates[0].Value
		parsed.Type = aggregates[0].Type
		// if gender is still unspecified, choose it here
		if parsed.Gender == data.ProductGenderUnisex {
			parsed.Gender = aggregates[0].Gender
		}
	}
	return parsed
}

/*
	Searches the blacklist for the range of inputs given
	Returns true if found
	Returns false if not found
*/
func SearchBlackList(ctx context.Context, inputs ...string) bool {

	blackListCache, _ := ctx.Value(cacheKeyBlacklist).(map[string]*data.Blacklist)
	if blackListCache == nil {
		//blacklist not in context generate it here
		blacklist, _ := data.DB.Blacklist.FindAll(nil)
		if blacklist != nil {
			blackListWordMap := make(map[string]*data.Blacklist, len(blacklist))
			for _, word := range blacklist {
				blackListWordMap[word.Word] = word
			}
			blackListCache = blackListWordMap
		} else {
			return false
		}
		ctx = context.WithValue(ctx, cacheKeyBlacklist, blackListCache)
	}

	// iterating over each input
	for _, s := range inputs {
		tokens := tokenize(s)
		// searching if any of the tokens is in the blacklist
		for _, token := range tokens {
			if _, found := blackListCache[token]; found {
				return true
				break
			}
		}
	}
	return false
}
