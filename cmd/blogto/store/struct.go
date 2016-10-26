package store

type Store struct {
	Address   string `json:"Address"`
	Latitude  string `json:"Latitude"`
	Longitude string `json:"Longitude"`
	Name      string `json:"Name"`
	Phone     string `json:"Phone"`
	URLPath   string `json:"URL"`
	ListingID int    `json:"ListingID"`
}

type StoreDetail struct {
	Address        string `json:"address"`
	Author         string `json:"author"`
	Bestofmentions []struct {
		ListID    string `json:"list_id"`
		ListTitle string `json:"list_title"`
		ListURL   string `json:"list_url"`
		Rank      string `json:"rank"`
	} `json:"bestofmentions"`
	Category          string        `json:"category"`
	Closed            string        `json:"closed"`
	CommentCount      int           `json:"comment_count"`
	Comments          []interface{} `json:"comments"`
	DinesafeURL       string        `json:"dinesafe_url"`
	Distance          int           `json:"distance"`
	EntryID           string        `json:"entry_id"`
	Excerpt           string        `json:"excerpt"`
	ID                string        `json:"id"`
	Image             string        `json:"image"`
	Images            []string      `json:"images"`
	Latitude          string        `json:"latitude"`
	Longitude         string        `json:"longitude"`
	Name              string        `json:"name"`
	Neighborhood      string        `json:"neighborhood"`
	NeighbourhoodID   string        `json:"neighbourhood_id"`
	NeighbourhoodName string        `json:"neighbourhood_name"`
	Phone             string        `json:"phone"`
	Popularity        string        `json:"popularity"`
	PublishDate       string        `json:"publish_date"`
	Rating            int           `json:"rating"`
	RatingVotes       interface{}   `json:"rating_votes"`
	RecentlyOpened    string        `json:"recently_opened"`
	Review            string        `json:"review"`
	ReviewDate        string        `json:"review_date"`
	Reviewed          string        `json:"reviewed"`
	Status            string        `json:"status"`
	Thumbnail         string        `json:"thumbnail"`
	TypeID            string        `json:"type_id"`
	TypeName          string        `json:"type_name"`
	TypeShorthand     string        `json:"type_shorthand"`
	TypeSlug          string        `json:"type_slug"`
	URL               string        `json:"url"`
	Website           string        `json:"website"`
}
