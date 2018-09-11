package sync

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/gedex/inflector"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	set "gopkg.in/fatih/set.v0"
)

var (
	tagRegex = regexp.MustCompile("[^a-zA-Z0-9-]+")
)

var (
	ErrBlacklisted = errors.New("blacklisted")
	ErrNoCategory  = errors.New("no category")
)

var (
	genderMatchScore = 2
)

type aggregateCategory []*data.Whitelist

func (s aggregateCategory) Len() int {
	return len(s)
}

func (s aggregateCategory) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s aggregateCategory) Less(i, j int) bool {
	return (s[i].Weight > s[j].Weight)
}

func hasNoLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func (p *parser) tokenize(tagStr string, optTags ...string) []string {
	tagStr = strings.ToLower(tagStr)
	tt := tagRegex.Split(tagStr, -1)
	tagSet := set.New()

	// if the passed in string is a product handle
	// parse it by spliting it up by dashes
	if len(tt) == 1 && strings.Count(tagStr, "-") > 1 {
		tt = strings.Split(tagStr, "-")
	}

	// word tags
	slugTagStr := slug.Make(strings.Join(tt, " "))
	for k, v := range p.whitelist {
		if v[0].IsSpecial && strings.Index(slugTagStr, k) != -1 {
			tagSet.Add(k)
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

type parser struct {
	gender     data.ProductGender
	categories map[string]data.Whitelist

	whitelist whitelist
	blacklist blacklist
}

// filter tags for categories
func (p *parser) parseCategory(tokens []string) {
	for _, t := range tokens {
		if w, found := p.whitelist[t]; found {
			if len(w) == 2 {
				// if the gender is ambigious... add a unisex category
				key := fmt.Sprintf("%s-%s", data.ProductGenderUnisex, w[0].Value) // use value + gender as the key
				if _, ok := p.categories[key]; !ok {
					p.categories[key] = data.Whitelist{
						Type:   w[0].Type,
						Value:  w[0].Value,
						Weight: w[0].Weight,
						Gender: data.ProductGenderUnisex,
					}
				}
			}
			// iterate over the genders
			for _, x := range w {
				key := fmt.Sprintf("%s-%s", x.Gender, x.Value) // use value + gender as the key
				y, ok := p.categories[key]                     // check if already exists
				if !ok {
					y = data.Whitelist{
						Type:       x.Type,
						Value:      x.Value,
						Gender:     x.Gender,
						CategoryID: x.CategoryID,
					}
				}
				// either way, increment the weighting
				y.Weight = x.Weight
				p.categories[key] = y
			}
		}
	}
}

// filter tags for genders
func (p *parser) parseGender(tokens []string) {
	if p.gender != data.ProductGenderUnisex {
		return
	}

	var maybeGender data.ProductGender
	for _, t := range tokens {
		switch t {
		case "man", "male", "gentleman", "guy", "boy":
			p.gender = data.ProductGenderMale
			return
		case "woman", "female", "lady", "gal", "girl":
			p.gender = data.ProductGenderFemale
			return
		case "kid":
			p.gender = data.ProductGenderKid
			return
		case "sexy":
			// maybe female.
			if maybeGender == data.ProductGenderUnknown {
				maybeGender = data.ProductGenderFemale
			}
		}
	}

	if maybeGender != data.ProductGenderUnknown {
		p.gender = maybeGender
		return
	}

	p.gender = data.ProductGenderUnisex
}

func newParser(ctx context.Context) *parser {
	place := ctx.Value("sync.place").(*data.Place)
	return &parser{
		// assuming place Gender is (at least) Unisex
		gender:     data.ProductGender(place.Gender),
		categories: map[string]data.Whitelist{},
		whitelist:  whitelistCache,
		blacklist:  blacklistCache,
	}
}

func ParseProduct(ctx context.Context, inputs ...string) (data.Whitelist, error) {
	p := newParser(ctx)

	// search blacklist first.  if blacklisted, just return
	// NOTE:
	//  of course. this is pretty aggressive. ie. if we catch anything inside of
	//  the blacklist we basically throw the product out.
	//
	//  cases ilike... `iPhone print tshirt` wouldn't get a chance here
	//
	//  create a logic map
	//   - x blacklist + o category -> reject
	//   - x blacklist + x category -> pending?
	//   - o blacklist + o category -> pending?
	//   - o blacklist + x category -> good
	var err error
	if p.searchBlackList(inputs...) {
		err = ErrBlacklisted
	}
	return p.searchWhiteList(inputs...), err
}

func (p *parser) searchWhiteList(inputs ...string) data.Whitelist {
	for _, s := range inputs {
		tokens := p.tokenize(s)
		p.parseGender(tokens)
		p.parseCategory(tokens)

		log.Println(tokens)
		log.Println(p.gender)
	}

	var aggregates aggregateCategory
	// increase category weight for more gender occurences
	genderCount := map[data.ProductGender]int32{}
	typeCount := map[data.ProductCategoryType]int32{}
	for _, w := range p.categories {
		// copy incase the scope leaks
		a := w

		// keep track number of times gender category popedup
		genderCount[a.Gender] += int32(1)
		// don't over count unisex categories
		if a.Gender != data.ProductGenderUnisex {
			typeCount[a.Type] += int32(1)
		}

		if p.gender != data.ProductGenderUnisex &&
			a.Gender != p.gender && a.Gender != data.ProductGenderUnisex {
			// skip the matched whitelist entry if the
			// product gender and the whitelist gender
			// do not match up

			// for example, if parsed gender is "male",
			// skip the "female" categories. but do not skip the "unisex"
			// categories
			continue
		}
		// if the category matches the product gender, boost the weight
		if p.gender == a.Gender {
			a.Weight += int32(genderMatchScore)
			log.Printf("match: adding %d to %s => %d", genderMatchScore, a.Value, a.Weight)
		}
		aggregates = append(aggregates, &a)
	}

	// if the gender is not determined. increase the
	// category weight based on # of occurences
	if p.gender == data.ProductGenderUnisex {
		for _, a := range aggregates {
			a.Weight += genderCount[a.Gender]
			log.Printf("gender: adding %d to %s => %d", genderCount[a.Gender], a.Value, a.Weight)
		}
	}

	// if there are multiple detect product types (ie.. apparel vs shoes)
	// add weight to the highest occuring ones
	if len(typeCount) > 1 {
		for _, a := range aggregates {
			a.Weight += typeCount[a.Type]
			log.Printf("type: adding %d to %s => %d", typeCount[a.Type], a.Value, a.Weight)
		}
	}

	// sort categories by weight
	sort.Sort(aggregates)

	for _, a := range aggregates {
		log.Printf("aggregate: %v", a)
	}

	parsed := data.Whitelist{
		// inherit from the parser
		Gender: p.gender,
	}

	if len(aggregates) > 0 {
		// use the parsed out category (sorted with the highest weighting) and insert as value
		parsed.Value = aggregates[0].Value
		parsed.Type = aggregates[0].Type
		parsed.CategoryID = aggregates[0].CategoryID
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
func (p *parser) searchBlackList(inputs ...string) bool {
	// iterating over each input
	for _, s := range inputs {
		tokens := p.tokenize(s)
		// searching if any of the tokens is in the blacklist
		for _, token := range tokens {
			if _, found := p.blacklist[token]; found {
				return true
				break
			}
		}
	}
	return false
}

type shopifyCategorySyncer struct {
	product *data.Product
	place   *data.Place
}

func (s *shopifyCategorySyncer) Sync(inputs ...string) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "sync.place", s.place)

	// find product category + gender
	parsedData, err := ParseProduct(ctx, inputs...)
	if err != nil {
		// see parse product comment for logic on blacklisting product
		// throw away the product. and continue on
		// TODO: keep track of how many products are rejected

		//  create a logic map
		//   - x blacklist + o category -> reject
		//   - x blacklist + x category -> pending?
		//   - o blacklist + o category -> pending?
		//   - o blacklist + x category -> good
		if err == ErrBlacklisted {
			if len(parsedData.Value) == 0 {
				// no category and blacklisted -> return rejected
				return err
			}
			// blacklisted but has category. set the product as pending
			// and do not yet return.
			s.product.Status = data.ProductStatusPending
		}
	}
	if len(parsedData.Value) == 0 {
		// not blacklisted but did not find category
		s.product.Status = data.ProductStatusPending
	}
	s.product.Gender = parsedData.Gender
	s.product.Category = data.ProductCategory{
		Type:  parsedData.Type,
		Value: parsedData.Value,
	}
	s.product.CategoryID = parsedData.CategoryID

	return nil
}
