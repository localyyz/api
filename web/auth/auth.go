package auth

import (
	"net/http"
	"strings"

	"upper.io/db"

	"bitbucket.org/pxue/api/data"
	"bitbucket.org/pxue/api/lib/connect"
	"bitbucket.org/pxue/api/lib/ws"

	"github.com/pressly/chi"
	"golang.org/x/net/context"
)

func AuthCtx(next chi.Handler) chi.Handler {
	return chi.HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		connectID := strings.ToLower(chi.URLParam(ctx, "network"))

		auth, err := connect.NewSocialAuth(connectID)
		if err != nil {
			ws.Respond(w, http.StatusNotFound, err)
			return
		}

		ctx = context.WithValue(ctx, "network", connectID)
		ctx = context.WithValue(ctx, "auth", auth)
		next.ServeHTTPC(ctx, w, r)
	})
}

// Login handles both first time login (signup) and repeated login
func Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	network := ctx.Value("network").(string)
	auth := ctx.Value("auth").(connect.Connect)

	var payload struct {
		ShortLivedToken string `json:"short_lived_token"`
		Username        string `json:"username"`
		UID             string `json:"uid"`
	}
	if err := ws.Bind(r.Body, &payload); err != nil {
		ws.Respond(w, http.StatusBadRequest, err)
		return
	}

	dbUser, err := data.DB.Account.FindByUsername(payload.UID)
	if err != nil {
		if err != db.ErrNoMoreRows {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
	}

	if payload.ShortLivedToken != "" && dbUser == nil {
		usrToken, err := auth.ExchangeToken(payload.ShortLivedToken)
		if err != nil {
			ws.Respond(w, http.StatusServiceUnavailable, err)
			return
		}

		dbUser = &data.Account{
			AccessToken: usrToken,
			Network:     network,
		}
		if err := auth.GetUser(dbUser); err != nil {
			ws.Respond(w, http.StatusServiceUnavailable, err)
			return
		}

		if err := data.DB.Account.Save(dbUser); err != nil {
			ws.Respond(w, http.StatusInternalServerError, err)
			return
		}
	}

	ws.Respond(w, http.StatusOK, dbUser)
}

func Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
}
