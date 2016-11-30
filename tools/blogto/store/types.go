package store

type Store struct {
	Address                 string        `json:"address"`
	AppearsInBestOfLists    bool          `json:"appears_in_best_of_lists"`
	Coordinates             Coordinate    `json:"coordinates"`
	DatePublished           string        `json:"date_published"`
	DefaultNeighborhood     Neighbourhood `json:"default_neighborhood"`
	DinesafeEstablishmentID string        `json:"dinesafe_establishment_id"`
	ID                      int           `json:"id"`
	ImageURL                string        `json:"image_url"`
	Name                    string        `json:"name"`
	Phone                   string        `json:"phone"`
	Rating                  float64       `json:"rating"`
	ShareURL                string        `json:"share_url"`
	SubType                 Category      `json:"sub_type"`
	Type                    Category      `json:"type"`
	Website                 string        `json:"website"`
}

type Neighbourhood struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type Category struct {
	ID       CategoryType `json:"id"`
	Name     string       `json:"name"`
	ShareURL string       `json:"share_url"`
}

type Coordinate struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Primary Category type
type CategoryType uint32
type SubcategoryType uint32

const (
	_ CategoryType = iota
	CategoryTypeRestaurant
	CategoryTypeBar
	CategoryTypeCafe
	CategoryTypeDesign
	CategoryTypeFashion
	CategoryTypeGrocery
	CategoryTypeGallerie
	CategoryTypeBookstore
	CategoryTypeBakerie
	CategoryTypeFitness
	CategoryTypeHotel
	CategoryTypeService
)
