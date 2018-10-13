package category

import (
	"errors"
	"fmt"
	"net/http"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type stylesRequest struct {
	Gender  []string `json:"gender"`
	Pricing []string `json:"pricing"`
}

func (s *stylesRequest) Bind(r *http.Request) error {
	if len(s.Gender) == 0 {
		err := errors.New("missing gender")
		return api.ErrInvalidRequest(err)
	}
	if len(s.Pricing) == 0 {
		err := errors.New("missing pricing")
		return api.ErrInvalidRequest(err)
	}
	return nil
}

type Style struct {
	ID       string `json:"id"`
	Value    string `json:"value"`
	Label    string `json:"label"`
	ImageURL string `json:"imageUrl"`
}

var styleImage = map[string]string{
	"artsy-man":           "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Adam-Katz-Sinding-Simon-Rasmussen-Paris-Fashion-Week-Spring-Summer-2019_AKS4454-1.jpg?2212093451777825601",
	"artsy-woman":         "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Artsy.jpg?15479008772991519334",
	"american-man":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/American.jpg?15479008772991519334",
	"american-woman":      "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/American_6be033c2-c61e-4b68-8d2a-5df4995bf1ae.jpg?15479008772991519334",
	"athletic-man":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Athletic.jpg?15479008772991519334",
	"athletic-woman":      "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Athleisure.jpg?15479008772991519334",
	"bohemian-woman":      "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Bohemian.jpg?15479008772991519334",
	"bridal-woman":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Bridal.jpg?15479008772991519334",
	"business-woman":      "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Business_ae40b13f-f525-42d3-beb8-72683f551e17.jpg?15479008772991519334",
	"business-man":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Business.jpg?15479008772991519334",
	"casual-man":          "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Casual.jpg?9798129902872748421",
	"casual-woman":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Casual_c63c9c64-9c58-4fe9-8fee-715dc1c2213a.jpg?15479008772991519334",
	"chic-woman":          "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Chic.jpg?15479008772991519334",
	"hip-hop-woman":       "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Hip_Hop.jpg?15479008772991519334",
	"hip-hop-man":         "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Hip-Hop.jpg?15479008772991519334",
	"rave-man":            "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Rave-2.jpg?15479008772991519334",
	"rave-woman":          "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Rave.jpg?15479008772991519334",
	"rocker-man":          "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Rocker-2.jpg?15479008772991519334",
	"rocker-woman":        "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Rocker.jpg?15479008772991519334",
	"sophisticated-man":   "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Sophisticated-3.jpg?15479008772991519334",
	"sophisticated-woman": "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Sophisticated.jpg?15479008772991519334",
	"vintage-woman":       "https://cdn.shopify.com/s/files/1/0052/8731/3526/files/Vintage.jpg?15479008772991519334",
}

func ListStyles(w http.ResponseWriter, r *http.Request) {
	var payload stylesRequest
	if err := render.Bind(r, &payload); err != nil {
		render.Respond(w, r, err)
		return
	}

	// select from place meta the styles that match gender and pricing
	selectCol := "style_female"
	// TODO cast this to product gender
	if payload.Gender[0] == "man" {
		selectCol = "style_male"
	}
	rows, _ := data.DB.Select(selectCol).
		From("place_meta").
		Where(db.Cond{
			"pricing": payload.Pricing,
			selectCol: db.IsNotNull(),
		}).
		GroupBy(selectCol).
		Query()

	var styles []Style
	for {
		if !rows.Next() {
			break
		}
		var s string
		if err := rows.Scan(&s); err != nil {
			lg.Warn(err)
			continue
		}

		st := Style{
			Label:    s,
			Value:    s,
			ID:       s,
			ImageURL: styleImage[fmt.Sprintf("%s-%s", s, payload.Gender[0])],
		}
		styles = append(styles, st)
	}

	render.Respond(w, r, styles)
}
