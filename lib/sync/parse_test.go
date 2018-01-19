package sync

import (
	"context"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type tagTest struct {
	name     string
	inputs   []string
	place    *data.Place
	expected *data.ProductCategory
}

var (
	placeMale   = &data.Place{Gender: data.PlaceGenderMale}
	placeFemale = &data.Place{Gender: data.PlaceGenderFemale}
	placeUnisex = &data.Place{Gender: data.PlaceGenderUnisex}
)

func TestProductGender(t *testing.T) {
	t.Parallel()
	cache := map[string]*data.ProductCategory{
		"drake":       &data.ProductCategory{Gender: data.ProductGenderMale, Value: "drake"},
		"beyonce":     &data.ProductCategory{Gender: data.ProductGenderFemale, Value: "beyonce"},
		"brucejenner": &data.ProductCategory{Gender: data.ProductGenderUnisex, Value: "brucejenner"},
		"shoe":        &data.ProductCategory{Gender: data.ProductGenderUnisex, Value: "shoe"},
		"lace-up":     &data.ProductCategory{Weight: 1, Gender: data.ProductGenderUnisex, Value: "lace-up"},

		"shirt":   &data.ProductCategory{Weight: 1, Gender: data.ProductGenderMale, Value: "shirt"},
		"t-shirt": &data.ProductCategory{Weight: 1, Gender: data.ProductGenderUnisex, Value: "t-shirt"},
	}
	ctx := context.WithValue(context.Background(), cacheKey, cache)

	tests := []tagTest{
		{
			name:     "male category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"Drake is best man singer"},
			expected: &data.ProductCategory{Value: "drake", Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"females love drake"},
			expected: &data.ProductCategory{Value: "drake", Gender: data.ProductGenderFemale},
		},
		{
			name:     "male category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love Drake"},
			expected: &data.ProductCategory{Value: "drake", Gender: data.ProductGenderMale},
		},
		{
			name:     "female category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"beyonce is the greatest woman singer of all time"},
			expected: &data.ProductCategory{Value: "beyonce", Gender: data.ProductGenderFemale},
		},
		{
			name:     "female category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"all men should listen to at least one beyonce song"},
			expected: &data.ProductCategory{Value: "beyonce", Gender: data.ProductGenderMale},
		},
		{
			name:     "female category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love beyonce"},
			expected: &data.ProductCategory{Value: "beyonce", Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was a man"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"brucejenner became a woman"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was on vanity fair cover"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderUnisex},
		},
		// place gender
		{
			name:     "unisex category with place gender male",
			place:    placeMale,
			inputs:   []string{"brucejenner was an olympian"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with place gender male",
			place:    placeMale,
			inputs:   []string{"drake is from forest hill"},
			expected: &data.ProductCategory{Value: "drake", Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with place gender female",
			place:    placeFemale,
			inputs:   []string{"brucejenner was an athelete"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with place gender unisex",
			place:    placeUnisex,
			inputs:   []string{"brucejenner is kylies dad"},
			expected: &data.ProductCategory{Value: "brucejenner", Gender: data.ProductGenderUnisex},
		},
		{
			name:     "hyphonated compound category",
			place:    placeUnisex,
			inputs:   []string{"mens cool t-shirt"},
			expected: &data.ProductCategory{Value: "t-shirt", Gender: data.ProductGenderMale},
		},
		{
			name:     "hyphonated compound category with female gender hint and higher weighted category",
			place:    placeUnisex,
			inputs:   []string{"Lace-up Warm Cotton Shoes Female"},
			expected: &data.ProductCategory{Value: "lace-up", Gender: data.ProductGenderFemale},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = context.WithValue(ctx, "sync.place", tt.place)
			actual := ParseProduct(ctx, tt.inputs...)
			tt.compare(t, actual)
		})
	}
}

func (tt tagTest) compare(t *testing.T, actual data.ProductCategory) {
	if actual.Gender != tt.expected.Gender {
		t.Errorf("test '%s': expected gender '%v', got '%v'", tt.name, tt.expected.Gender, actual.Gender)
	}
	if actual.Value != tt.expected.Value {
		t.Errorf("test '%s': expected category '%s', got '%s'", tt.name, tt.expected.Value, actual.Value)
	}
}
