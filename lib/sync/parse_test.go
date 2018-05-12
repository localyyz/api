package sync

import (
	"context"
	"fmt"
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

func TestProductCategory(t *testing.T) {
	t.Parallel()

	cache := map[string]*data.Category{
		"dress": &data.Category{
			Gender:  data.ProductGenderFemale,
			Type:    data.CategoryApparel,
			Value:   "dress",
			Mapping: "dresses",
		},
		"eau-de-parfum": &data.Category{
			Gender:  data.ProductGenderUnisex,
			Type:    data.CategoryFragrance,
			Value:   "eau-de-parfum",
			Mapping: "perfume",
		},
		"bag": &data.Category{
			Gender:  data.ProductGenderUnisex,
			Type:    data.CategoryBag,
			Value:   "bag",
			Mapping: "",
		},
		"sunglass": &data.Category{
			Gender:  data.ProductGenderUnisex,
			Type:    data.CategoryAccessory,
			Value:   "sunglass",
			Mapping: "sunglasses",
		},
		"v-neck": &data.Category{
			Gender:  data.ProductGenderUnisex,
			Type:    data.CategoryApparel,
			Value:   "v-neck",
			Mapping: "",
		},
		"face-mask": &data.Category{
			Gender:  data.ProductGenderUnisex,
			Type:    data.CategoryCosmetic,
			Value:   "face-mask",
			Mapping: "face mask",
		},
		"eau-de-toilette": &data.Category{
			Gender:  data.ProductGenderMale,
			Type:    data.CategoryFragrance,
			Value:   "eau-de-toilette",
			Mapping: "cologne",
		},
	}

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
			expected: &data.Category{Gender: data.ProductGenderMale, Type: data.CategoryBag, Value: "bag", Mapping: ""},
		},
		{
			name:     "Sunglass",
			inputs:   []string{"Lacoste L829S Brown  Sunglasses RRP £102"},
			place:    placeUnisex,
			expected: cache["sunglass"],
		},
		//{ TODO: undeterministic could be vneck or dress
		//name:     "V-neck dress",
		//inputs:   []string{"Fashion Maternity V-neck Short Sleeve Cotton Pregnancy Dress Elastic Waist Dresses"},
		//place:    placeFemale,
		//expected: &data.Category{Gender: data.ProductGenderFemale, Type: data.CategoryApparel, Value: "dress", Mapping: ""},
		//},
		{
			name:     "Face mask",
			inputs:   []string{"Apple, Aloe & Avocado Face Mask"},
			place:    placeFemale,
			expected: &data.Category{Gender: data.ProductGenderFemale, Type: data.CategoryCosmetic, Value: "face-mask", Mapping: "face mask"},
		},

		{
			name:     "Eau De Toillet",
			inputs:   []string{"Burberry Sport Ice by Burberry Eau De Toilette Spray 2.5 oz"},
			place:    placeUnisex,
			expected: cache["eau-de-toilette"],
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

func TestProductBlacklist(t *testing.T) {
	t.Parallel()

	cache := map[string]*data.Blacklist{
		"iphone":    &data.Blacklist{Word: "iphone"},
		"phone":     &data.Blacklist{Word: "phone"},
		"lcd":       &data.Blacklist{Word: "lcd"},
		"diy":       &data.Blacklist{Word: "diy"},
		"electric":  &data.Blacklist{Word: "electric"},
		"car":       &data.Blacklist{Word: "car"},
		"equipment": &data.Blacklist{Word: "equipment"},
		"hdd":       &data.Blacklist{Word: "hdd"},
		"canon":     &data.Blacklist{Word: "canon"},
		"tray":      &data.Blacklist{Word: "tray"},
		"3d":        &data.Blacklist{Word: "3d"},
		"bicycle":   &data.Blacklist{Word: "bicycle"},
		"mug":       &data.Blacklist{Word: "mug"},
		"passport":  &data.Blacklist{Word: "passport"},
		"card":      &data.Blacklist{Word: "card"},
		"brush":     &data.Blacklist{Word: "brush"},
		"book":      &data.Blacklist{Word: "book"},
	}
	ctx := context.WithValue(context.Background(), cacheKeyBlacklist, cache)

	/* all values from db */
	var products = []string{
		"Original Apple iPhone 7 2GB",
		"Fashion  Phone Case For iPhone 6s Plu Luxurious Simple Love Heart Soft TPU Blue Pink  For iPhone 7 8 7 Plus X",
		"High Speed 1m 2m HDMI 2.0 Cable HDMI Male To HDMI Male Cabo For HD TV LCD Laptop PC PS3 Projector Displayer Cable V2 4k 3D 1080p",
		"Wall stickers big 3d decor large mirror pattern surface DIY wall sticker",
		"5500W instantaneous water heater tap water heater instant water heater electric shower free shipping",
		"2PCS Screen Protector Glass Huawei Honor 5C Tempered Glass For Huawei Honor 5C Glass Anti-scratch Phone Film Honor 5C WolfRule [",
		"Car Air Ozonizer - Air Purifier",
		"3M Bearing Skip Rope Cord Speed Fitness Aerobic Jumping Exercise Equipment Adjustable Boxing Skipping Sport Jump Ropes",
		"18cm Molex to SATA HDD Hard Drive Power Cord",
		"FALCONEYES Honeycomb Grip with 10 color flash gels set for Canon Nikon YONGNUO Metz Nissin Flash Gun Speedlites",
		"Garlic Press  Very Sharp Stainless Steel Blades, Inbuilt Clear Plastic Tray, Green",
		"[SHIJUEHEZI] Outer Space Planets 3D Wall Stickers Cosmic Galaxy Wall Decals for Kids Room Baby Bedroom Ceiling Floor Decoration",
		"360 Degree Rotation Cycling Bike Bicycle Flashlight Torch Mount LED Head Front Light Holder Clip Bicycle Accessories",
		"Cafe Is Life Mug",
		"Kids DIY Garden Starter Growing Kit - Teach a child how to grow a home grown garden",
		"Royal Blue Passport Cover",
		"6 stick 3D colorful patterns makeup brush a set",
		"I Really Fucked That Up Letterpress Card",
		"Sperm Whale Brush",
		"Cinderella Picture Book - Free Shipping",
	}

	for i, tt := range products {
		t.Run(fmt.Sprintf("blacklist %d", i), func(t *testing.T) {
			blacklisted := SearchBlackList(ctx, tt)
			if !blacklisted {
				t.Error("Fail: ", tt)
			}
		})
	}
}

func TestFinalizeProductStatus(t *testing.T) {
	t.Parallel()

	cache := map[string]*data.Blacklist{
		"blacklist": &data.Blacklist{
			Word: "blacklist",
		},
	}
	ctx := context.WithValue(context.Background(), cacheKeyBlacklist, cache)

	tests := []struct {
		name        string
		hasCategory bool
		input       string
		expected    data.ProductStatus
	}{
		{"approved", true, "ok product", data.ProductStatusApproved},
		{"pending", true, "blacklist", data.ProductStatusPending},
		{"rejected", false, "blacklist", data.ProductStatusRejected},
		{"pending no category or blacklist", false, "product", data.ProductStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := finalizeStatus(ctx, tt.hasCategory, tt.input)
			if tt.expected != actual {
				t.Errorf("test '%s': expected type '%s', got '%s'", tt.name, tt.expected, actual)
			}
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
