package category

import (
	"errors"
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

	var styles []string
	for {
		if !rows.Next() {
			break
		}
		var s string
		if err := rows.Scan(&s); err != nil {
			lg.Warn(err)
			continue
		}
		styles = append(styles, s)
	}

	render.Respond(w, r, styles)
}
