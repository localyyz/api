package sync

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type tagTest struct {
	name     string
	inputs   []string
	place    *data.Place
	expected data.Whitelist
}

var (
	placeMale   = &data.Place{Gender: data.PlaceGenderMale}
	placeFemale = &data.Place{Gender: data.PlaceGenderFemale}
	placeUnisex = &data.Place{Gender: data.PlaceGenderUnisex}
)

func TestTokenizeSpecial(t *testing.T) {
	t.Parallel()

	cache := whitelist{
		"jeans": {
			data.Whitelist{
				Value: "jeans",
			},
		},
		"jean": {
			data.Whitelist{
				Value: "jean",
			},
		},
		"jean-paul-gaultier": {
			data.Whitelist{
				Value:     "jean-paul-gaultier",
				IsSpecial: true,
				Weight:    -1,
			},
		},
		"versace-jeans": {
			data.Whitelist{
				Value:     "versace-jeans",
				IsSpecial: true,
				Weight:    -1,
			},
		},
		"carrera-jeans": {
			data.Whitelist{
				Value:     "carrera-jeans",
				IsSpecial: true,
				Weight:    -1,
			},
		},
		"ankle-boot": {
			data.Whitelist{
				Value:     "ankle-boot",
				IsSpecial: true,
			},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "versace jeans",
			input:    "versace jeans women grey crossbody bags",
			expected: []string{"woman", "grey", "crossbody", "bag"},
		},
		{
			name:     "carrera jeans ankle boot",
			input:    "carrera jeans tennesse men brown ankle boots",
			expected: []string{"tennesse", "man", "brown", "ankle-boot"},
		},
		{
			name:     "jean paul gaultier",
			input:    "jean paul gaultier pencil skirt",
			expected: []string{"pencil", "skirt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(context.WithValue(context.Background(), "sync.place", &data.Place{}))
			p.whitelist = cache

			a := p.tokenize(tt.input)
			e := tt.expected

			// sort
			sort.Strings(a)
			sort.Strings(e)

			if !reflect.DeepEqual(a, e) {
				t.Fatalf("test '%s' tokenize not equal. e: %v got %v", tt.name, e, a)
			}
		})
	}
}

func TestWhitelistSpecial(t *testing.T) {
	t.Parallel()

	cache := whitelist{
		"jeans": {
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryApparel,
				Value:  "jeans",
				Weight: 1,
			},
			data.Whitelist{
				Gender: data.ProductGenderFemale,
				Type:   data.CategoryApparel,
				Value:  "jeans",
				Weight: 1,
			},
		},
		"jean-paul": {
			data.Whitelist{
				Weight:    -1,
				Value:     "jean-paul",
				IsSpecial: true,
			},
		},
		"versace-jeans": {
			data.Whitelist{
				Weight:    -1,
				Value:     "versace-jeans",
				IsSpecial: true,
			},
		},
		"carrera-jeans": {
			data.Whitelist{
				Weight:    -1,
				Value:     "carrera-jeans",
				IsSpecial: true,
			},
		},
		"ankle-boot": {
			data.Whitelist{
				Weight:    2,
				Value:     "ankle-boot",
				Gender:    data.ProductGenderFemale,
				Type:      data.CategoryShoe,
				IsSpecial: true,
			},
			data.Whitelist{
				Weight:    2,
				Value:     "ankle-boot",
				Gender:    data.ProductGenderMale,
				Type:      data.CategoryShoe,
				IsSpecial: true,
			},
		},
	}

	tests := []tagTest{
		{
			name:   "versace jeans",
			inputs: []string{"versace jeans women grey crossbody bags"},
			place:  placeUnisex,
		},
		{
			name:     "carrera jeans",
			inputs:   []string{"carrera jeans tennesse men brown ankle boots"},
			place:    placeUnisex,
			expected: cache["ankle-boot"][1],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(context.WithValue(context.Background(), "sync.place", tt.place))
			p.whitelist = cache

			actual := p.searchWhiteList(tt.inputs...)
			if actual.Value == "jeans" {
				t.Errorf("test '%s' unexpectedly categorized as 'jeans'", tt.name)
			}
			if tt.expected.Value != "" {
				tt.compare(t, actual)
			}
		})
	}
}

func TestWhitelistOxfordShoe(t *testing.T) {
	t.Parallel()

	cache := whitelist{
		"oxford": {
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryShoe,
				Value:  "oxford",
				Weight: 2,
			},
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryApparel,
				Value:  "oxford",
				Weight: 2,
			},
		},
		"shirt": {
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryApparel,
				Value:  "shirt",
				Weight: 1,
			},
		},
		"shoe": {
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryShoe,
				Value:  "shoe",
				Weight: 0,
			},
		},
	}

	tests := []tagTest{
		{
			name:     "oxford shirt",
			inputs:   []string{"j lindeberg daniel stretch oxford shirt (grey)"},
			place:    placeUnisex,
			expected: cache["oxford"][1],
		},
		{
			name:     "oxford shoes",
			inputs:   []string{"whole cut plain oxford dress shoes"},
			place:    placeUnisex,
			expected: cache["oxford"][0],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(context.WithValue(context.Background(), "sync.place", tt.place))
			p.whitelist = cache

			actual := p.searchWhiteList(tt.inputs...)
			tt.compare(t, actual)
		})
	}
}

func TestWhitelistBathingSuit(t *testing.T) {
	t.Parallel()

	cache := whitelist{
		"suit": {
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryApparel,
				Value:  "suit",
				Weight: 2,
			},
		},
		"bathing-suit": {
			data.Whitelist{
				Gender:    data.ProductGenderFemale,
				Type:      data.CategoryApparel,
				Value:     "bathing-suit",
				Weight:    2 + 1,
				IsSpecial: true,
			},
		},
	}

	tests := []tagTest{
		{
			name:     "bathing suit",
			inputs:   []string{"High Cut Summer Bathing Suit"},
			place:    placeUnisex,
			expected: cache["bathing-suit"][0],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(context.WithValue(context.Background(), "sync.place", tt.place))
			p.whitelist = cache

			actual := p.searchWhiteList(tt.inputs...)
			tt.compare(t, actual)
		})
	}
}

func TestWhitelist(t *testing.T) {
	t.Parallel()

	cache := whitelist{
		"dress": {
			data.Whitelist{
				Gender: data.ProductGenderFemale,
				Type:   data.CategoryApparel,
				Value:  "dress",
				Weight: 1,
			},
		},
		"bag": {
			data.Whitelist{
				Gender: data.ProductGenderUnisex,
				Type:   data.CategoryBag,
				Value:  "bag",
				Weight: 0,
			},
		},
		"backpack": {
			data.Whitelist{
				Gender: data.ProductGenderFemale,
				Type:   data.CategoryBag,
				Value:  "backpack",
				Weight: 2,
			},
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryBag,
				Value:  "backpack",
				Weight: 2,
			},
		},
		"jean": {
			data.Whitelist{
				Gender: data.ProductGenderFemale,
				Type:   data.CategoryApparel,
				Value:  "jean",
				Weight: 2,
			},
			data.Whitelist{
				Gender: data.ProductGenderMale,
				Type:   data.CategoryApparel,
				Value:  "jean",
				Weight: 2,
			},
		},
		"jean-paul": {
			data.Whitelist{
				Gender:   data.ProductGenderUnisex,
				Type:     data.CategoryApparel,
				Value:    "jean-paul",
				IsIgnore: true,
			},
		},
	}

	tests := []tagTest{
		{
			name:     "dress",
			inputs:   []string{"Basic Dress in Light Gray Stine Ladefoged Basic Dress - LGHTGREY"},
			place:    placeUnisex,
			expected: cache["dress"][0],
		},
		{
			name:   "backpack",
			inputs: []string{"Original NIKE Training Backpacks Sports Bags"},
			place:  placeUnisex,
			expected: data.Whitelist{
				Gender: data.ProductGenderUnisex,
				Type:   data.CategoryBag,
				Value:  "backpack",
			},
		},
		{
			name:     "womens backpack",
			inputs:   []string{"Original NIKE 'womens' Training Backpacks Sports Bags"},
			place:    placeUnisex,
			expected: cache["backpack"][0],
		},
		{
			name:     "mens backpack",
			inputs:   []string{"Original NIKE 'mens' Training Backpacks Sports Bags"},
			place:    placeUnisex,
			expected: cache["backpack"][1],
		},
		{
			name:     "mens backpack",
			inputs:   []string{"Original NIKE 'mens' Training Backpacks Sports Bags"},
			place:    placeUnisex,
			expected: cache["backpack"][1],
		},
		//{
		//name:     "jean paul",
		//inputs:   []string{"JEAN PAUL GAULTIER Size L Burgundy Sheer Opaque Tatto Print Mini Dress"},
		//place:    placeUnisex,
		//expected: cache["dress"][0],
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(context.WithValue(context.Background(), "sync.place", tt.place))
			p.whitelist = cache

			actual := p.searchWhiteList(tt.inputs...)
			tt.compare(t, actual)
		})
	}
}

func TestProductGender(t *testing.T) {
	t.Parallel()
	cache := whitelist{
		"lovedrake": {
			data.Whitelist{Gender: data.ProductGenderMale, Type: data.CategoryApparel, Value: "lovedrake"},
			data.Whitelist{Gender: data.ProductGenderFemale, Type: data.CategoryApparel, Value: "lovedrake"},
		},
		"eminem":      {data.Whitelist{Gender: data.ProductGenderMale, Type: data.CategoryApparel, Value: "eminem"}},
		"beyonce":     {data.Whitelist{Gender: data.ProductGenderFemale, Type: data.CategoryHandbag, Value: "beyonce"}},
		"brucejenner": {data.Whitelist{Gender: data.ProductGenderUnisex, Type: data.CategoryAccessory, Value: "brucejenner"}},
		"shoe":        {data.Whitelist{Gender: data.ProductGenderUnisex, Type: data.CategoryShoe, Value: "shoe"}},
		"lace-up":     {data.Whitelist{Weight: 1, Gender: data.ProductGenderUnisex, Type: data.CategoryShoe, Value: "lace-up"}},

		"shirt":   {data.Whitelist{Weight: 1, Gender: data.ProductGenderMale, Type: data.CategoryApparel, Value: "shirt"}},
		"t-shirt": {data.Whitelist{Weight: 1, Gender: data.ProductGenderUnisex, Type: data.CategoryApparel, Value: "t-shirt"}},
	}

	tests := []tagTest{
		{
			name:     "male category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"eminem is best man singer"},
			expected: data.Whitelist{Value: "eminem", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"females lovedrake"},
			expected: data.Whitelist{Value: "lovedrake", Type: data.CategoryApparel, Gender: data.ProductGenderFemale},
		},
		{
			name:     "male category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love EMINEM"},
			expected: data.Whitelist{Value: "eminem", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "female category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"beyonce is the greatest woman singer of all time"},
			expected: data.Whitelist{Value: "beyonce", Type: data.CategoryHandbag, Gender: data.ProductGenderFemale},
		},
		{
			name:     "female category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"I love beyonce"},
			expected: data.Whitelist{Value: "beyonce", Type: data.CategoryHandbag, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with gender keyword male",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was a man"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with gender keyword female",
			place:    placeUnisex,
			inputs:   []string{"brucejenner became a woman"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with no gender keyword",
			place:    placeUnisex,
			inputs:   []string{"brucejenner was on vanity fair cover"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderUnisex},
		},
		{
			name:     "unisex category with place gender male",
			place:    placeMale,
			inputs:   []string{"brucejenner was an olympian"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderMale},
		},
		{
			name:     "male category with place gender male",
			place:    placeMale,
			inputs:   []string{"eminem is rap god"},
			expected: data.Whitelist{Value: "eminem", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "unisex category with place gender female",
			place:    placeFemale,
			inputs:   []string{"brucejenner was an athelete"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderFemale},
		},
		{
			name:     "unisex category with place gender unisex",
			place:    placeUnisex,
			inputs:   []string{"brucejenner is kylies dad"},
			expected: data.Whitelist{Value: "brucejenner", Type: data.CategoryAccessory, Gender: data.ProductGenderUnisex},
		},
		{
			name:     "hyphonated compound category",
			place:    placeUnisex,
			inputs:   []string{"mens cool t-shirt"},
			expected: data.Whitelist{Value: "t-shirt", Type: data.CategoryApparel, Gender: data.ProductGenderMale},
		},
		{
			name:     "hyphonated compound category with female gender hint and higher weighted category",
			place:    placeUnisex,
			inputs:   []string{"Lace-up Warm Cotton Shoes Female"},
			expected: data.Whitelist{Value: "lace-up", Type: data.CategoryShoe, Gender: data.ProductGenderFemale},
		},
		{
			name:     "gender hint sexy is female. only if nothing else is detected",
			place:    placeUnisex,
			inputs:   []string{"something something sexy"},
			expected: data.Whitelist{Gender: data.ProductGenderFemale},
		},
		{
			name:     "gender men with 'sexy'",
			place:    placeUnisex,
			inputs:   []string{"mens sexy something"},
			expected: data.Whitelist{Gender: data.ProductGenderMale},
		},
		{
			name:     "mixed signals",
			place:    placeFemale,
			inputs:   []string{"Fashion Shirt Dress Black Lapel Long Sleeve Belted A Line Dress Elegant Floral Long Dress"},
			expected: data.Whitelist{Gender: data.ProductGenderFemale},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "sync.place", tt.place)
			p := newParser(ctx)
			p.whitelist = cache

			actual := p.searchWhiteList(tt.inputs...)
			tt.compare(t, actual)
		})
	}
}

func TestProductBlacklist(t *testing.T) {
	t.Parallel()

	cache := map[string]data.Blacklist{
		"iphone":    data.Blacklist{Word: "iphone"},
		"phone":     data.Blacklist{Word: "phone"},
		"lcd":       data.Blacklist{Word: "lcd"},
		"diy":       data.Blacklist{Word: "diy"},
		"electric":  data.Blacklist{Word: "electric"},
		"car":       data.Blacklist{Word: "car"},
		"equipment": data.Blacklist{Word: "equipment"},
		"hdd":       data.Blacklist{Word: "hdd"},
		"canon":     data.Blacklist{Word: "canon"},
		"tray":      data.Blacklist{Word: "tray"},
		"3d":        data.Blacklist{Word: "3d"},
		"bicycle":   data.Blacklist{Word: "bicycle"},
		"mug":       data.Blacklist{Word: "mug"},
		"passport":  data.Blacklist{Word: "passport"},
		"card":      data.Blacklist{Word: "card"},
		"brush":     data.Blacklist{Word: "brush"},
		"book":      data.Blacklist{Word: "book"},
	}

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
			p := &parser{blacklist: cache}
			blacklisted := p.searchBlackList(tt)
			if !blacklisted {
				t.Error("Fail: ", tt)
			}
		})
	}
}

func (tt tagTest) compare(t *testing.T, actual data.Whitelist) {
	if actual.Gender != tt.expected.Gender {
		t.Errorf("test '%s': expected gender '%v', got '%v'", tt.name, tt.expected.Gender, actual.Gender)
	}
	if tt.expected.Type != 0 && actual.Type != tt.expected.Type {
		t.Errorf("test '%s': expected type '%s', got '%s'", tt.name, tt.expected.Type, actual.Type)
	}
	if tt.expected.Value != "" && actual.Value != tt.expected.Value {
		t.Errorf("test '%s': expected value '%s', got '%s'", tt.name, tt.expected.Value, actual.Value)
	}
}
