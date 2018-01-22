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
	expected *data.Category
}

var (
	placeMale   = &data.Place{Gender: data.PlaceGenderMale}
	placeFemale = &data.Place{Gender: data.PlaceGenderFemale}
	placeUnisex = &data.Place{Gender: data.PlaceGenderUnisex}
)

func TestProductGender(t *testing.T) {
	t.Parallel()
	cache := map[string]*data.Category{
		"drake":       &data.Category{Gender: data.ProductGenderMale, Type: data.CategoryApparel, Value: "drake"},
		"beyonce":     &data.Category{Gender: data.ProductGenderFemale, Type: data.CategoryHandbag, Value: "beyonce"},
		"brucejenner": &data.Category{Gender: data.ProductGenderUnisex, Type: data.CategoryAccessory, Value: "brucejenner"},
		"shoe":        &data.Category{Gender: data.ProductGenderUnisex, Type: data.CategoryShoe, Value: "shoe"},
		"lace-up":     &data.Category{Weight: 1, Gender: data.ProductGenderUnisex, Type: data.CategoryShoe, Value: "lace-up"},

		"shirt":   &data.Category{Weight: 1, Gender: data.ProductGenderMale, Type: data.CategoryApparel, Value: "shirt"},
		"t-shirt": &data.Category{Weight: 1, Gender: data.ProductGenderUnisex, Type: data.CategoryApparel, Value: "t-shirt"},
	}
	ctx := context.WithValue(context.Background(), cacheKey, cache)

	tests := []tagTest{
		{
			name:     "male category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"Drake is best man singer"},
			expected: &data.Category{Value: "drake", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"females love drake"},
			expected: &data.Category{Value: "drake", Type: data.CategoryApparel, Gender: data.ProductGenderFemale},
		},
		{
			name:     "male category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love Drake"},
			expected: &data.Category{Value: "drake", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "female category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"beyonce is the greatest woman singer of all time"},
			expected: &data.Category{Value: "beyonce", Type: data.CategoryHandbag, Gender: data.ProductGenderFemale},
		},
		{
			name:     "female category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"all men should listen to at least one beyonce song"},
			expected: &data.Category{Value: "beyonce", Type: data.CategoryHandbag, Gender: data.ProductGenderMale},
		},
		{
			name:     "female category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love beyonce"},
			expected: &data.Category{Value: "beyonce", Type: data.CategoryHandbag, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was a man"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"brucejenner became a woman"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was on vanity fair cover"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderUnisex},
		},
		// place gender
		{
			name:     "unisex category with place gender male",
			place:    placeMale,
			inputs:   []string{"brucejenner was an olympian"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with place gender male",
			place:    placeMale,
			inputs:   []string{"drake is from forest hill"},
			expected: &data.Category{Value: "drake", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with place gender female",
			place:    placeFemale,
			inputs:   []string{"brucejenner was an athelete"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with place gender unisex",
			place:    placeUnisex,
			inputs:   []string{"brucejenner is kylies dad"},
			expected: &data.Category{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderUnisex},
		},
		{
			name:     "hyphonated compound category",
			place:    placeUnisex,
			inputs:   []string{"mens cool t-shirt"},
			expected: &data.Category{Value: "t-shirt", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "hyphonated compound category with female gender hint and higher weighted category",
			place:    placeUnisex,
			inputs:   []string{"Lace-up Warm Cotton Shoes Female"},
			expected: &data.Category{Value: "lace-up", Type: data.CategoryShoe, Gender: data.ProductGenderFemale},
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

func (tt tagTest) compare(t *testing.T, actual data.Category) {
	if actual.Gender != tt.expected.Gender {
		t.Errorf("test '%s': expected gender '%v', got '%v'", tt.name, tt.expected.Gender, actual.Gender)
	}
	if actual.Type != tt.expected.Type {
		t.Errorf("test '%s': expected type '%s', got '%s'", tt.name, tt.expected.Type, actual.Type)
	}
	if actual.Value != tt.expected.Value {
		t.Errorf("test '%s': expected category '%s', got '%s'", tt.name, tt.expected.Value, actual.Value)
	}
}
