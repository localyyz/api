package data

import "fmt"

// TODO: move this to table
type Category uint32

const (
	CategoryUnknown Category = iota
	CategoryBarberShop
	CategoryBeautyProduct
	CategoryCostume
	CategoryEyewear
	CategoryHairSalon
	CategoryHat
	CategoryJewellery
	CategoryLingerie
	CategoryClothing
	CategoryVintage
	CategoryShoe
	CategorySpa
	CategorySportswear
	CategoryTshirt
	CategoryTailorAndRepair
)

var (
	categories = []string{
		"all",
		"barber shops",
		"beauty products",
		"costume",
		"eyewear",
		"hair salons",
		"hats",
		"jewellery",
		"lingerie",
		"clothing",
		"vintage",
		"shoes",
		"spa",
		"sportswear",
		"t-shirts",
		"tailor and repair",
	}
	Categories = categories
)

// String returns the string value of the status.
func (s Category) String() string {
	return categories[s]
}

// MarshalText satisfies TextMarshaler
func (s Category) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText satisfies TextUnmarshaler
func (s *Category) UnmarshalText(text []byte) error {
	enum := string(text)
	for i := 0; i < len(categories); i++ {
		if enum == categories[i] {
			*s = Category(i)
			return nil
		}
	}
	return fmt.Errorf("unknown category %s", enum)
}
