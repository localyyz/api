package sync

import (
	"context"
	"log"
	"os"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/config"
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
)

const CONFFILE = "../../config/api.conf" //path to configuration file

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

func TestProductCategory(t *testing.T) {
	t.Parallel()

	ConnectToDB()
	cache := CreateCategoryCache()
	ctx := context.WithValue(context.Background(), cacheKey, cache) //putting it in the context

	tests := []tagTest{
		{
			name:     "Dress",
			inputs:   []string{"Basic Dress in Light Gray Stine Ladefoged Basic Dress - LGHTGREY"},
			place:    placeUnisex,
			expected: cache["dress"],
		},
		{
			name:     "Perfume",
			inputs:   []string{"1 Million Prive Eau De Parfum Spray By Paco Rabanne"},
			place:    placeUnisex,
			expected: cache["eau-de-parfum"],
		},
		{
			name:     "Bag",
			inputs:   []string{"new 2017 hot sale fashion men bags, men famous brand design leather messenger bag, high quality man brand bag, wholesale price"},
			place:    placeMale,
			expected: cache["bag"],
		},
		{
			name:     "Sunglass",
			inputs:   []string{"Lacoste L829S Brown  Sunglasses RRP £102"},
			place:    placeUnisex,
			expected: cache["sunglass"],
		},
		{
			name:     "Vneck",
			inputs:   []string{"Fashion Maternity V-neck Short Sleeve Cotton Pregnancy Dress Elastic Waist Dresses"},
			place:    placeUnisex,
			expected: cache["v-neck"],
		},
		{
			name:     "Shoe",
			inputs:   []string{"Merkmak Fashion Camouflage Military Men Unisex Canvas Shoes Men Casual Shoes Autumn Breathable Camo Men Flats Chaussure Femme"},
			place:    placeUnisex,
			expected: cache["flat"],
		},
		{
			name:     "Handbag",
			inputs:   []string{"Xiniu Women's Messenger Bags Women Canvas Handbags Ladies Stripe Shoulder Bag bolsa feminina para mujer #GHYW"},
			place:    placeFemale,
			expected: cache["shoulder-bag"],
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
		{
			name:     "gender hint sexy is female. only if nothing else is detected",
			place:    placeUnisex,
			inputs:   []string{"something something sexy"},
			expected: &data.Category{Gender: data.ProductGenderFemale},
		},
		{
			name:     "gender men with 'sexy'",
			place:    placeUnisex,
			inputs:   []string{"mens sexy something"},
			expected: &data.Category{Gender: data.ProductGenderMale},
		},
		{
			name:     "mixed signals",
			place:    placeFemale,
			inputs:   []string{"Fashion Shirt Dress Black Lapel Long Sleeve Belted A Line Dress Elegant Floral Long Dress"},
			expected: &data.Category{Gender: data.ProductGenderFemale},
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
	if tt.expected.Type != 0 && actual.Type != tt.expected.Type {
		t.Errorf("test '%s': expected type '%s', got '%s'", tt.name, tt.expected.Type, actual.Type)
	}
	if tt.expected.Value != "" && actual.Value != tt.expected.Value {
		t.Errorf("test '%s': expected category '%s', got '%s'", tt.name, tt.expected.Value, actual.Value)
	}
}

/* connects to the DB by loading configuration file */
func ConnectToDB() {
	/* loading configuration */
	confFile := CONFFILE
	conf, err := config.NewFromFile(confFile, os.Getenv("CONFIG"))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error: Could not load configuration"))
	}

	/* creating new db session */
	if _, err = data.NewDBSession(&conf.DB); err != nil {
		log.Fatal(errors.Wrap(err, "Error: Could not connect to the database"))
	}
}

/* gets the categories from the database and puts them in a map */
func CreateCategoryCache() map[string]*data.Category {
	cache := make(map[string]*data.Category)
	if categories, _ := data.DB.Category.FindAll(nil); categories != nil {
		for _, category := range categories {
			cache[category.Value] = category
		}
	}
	return cache
}
