package presenter

import (
	"context"
	"net/http"

	"github.com/pressly/chi/render"

	"bitbucket.org/moodie-app/moodie-api/data"
)

type Locale struct {
	*data.Locale
	ctx context.Context
}

func NewLocale(ctx context.Context, locale *data.Locale) *Locale {
	return &Locale{
		ctx:    ctx,
		Locale: locale,
	}
}

func NewLocaleList(ctx context.Context, locales []*data.Locale) []render.Renderer {
	var list []render.Renderer
	for _, locale := range locales {
		list = append(list, NewLocale(ctx, locale))
	}
	return list
}

func (l *Locale) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
