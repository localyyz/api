package shopify

type CustomerAddress struct {
	Address1     string `json:"address1"`
	Address2     string `json:"address2,omitempty"`
	City         string `json:"city"`
	Company      string `json:"company,omitempty"`
	Country      string `json:"country"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone,omitempty"`
	Province     string `json:"province,omitempty"`
	ProvinceCode string `json:"province_code,omitempty"`
	Zip          string `json:"zip"`
	CountryCode  string `json:"country_code,omitempty"`
}

// TODO api
