package shopify

type CustomerAddress struct {
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Company      string `json:"company"`
	Country      string `json:"country"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Province     string `json:"province"`
	Zip          string `json:"zip"`
	CountryCode  string `json:"country_code,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
}

// TODO api
