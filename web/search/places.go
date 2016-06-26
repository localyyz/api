package search

import (
	"net/http"

	"golang.org/x/net/context"
)

func AutocompletePlaces(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var payload struct {
	}
}
