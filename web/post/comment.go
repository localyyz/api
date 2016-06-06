package post

import (
	"net/http"

	"github.com/pressly/chi"

	"golang.org/x/net/context"
)

func CommentCtx(next chi.Handler) chi.Handler {
	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {

		next.ServeHTTPC(ctx, w, r)
	}
	return chi.HandlerFunc(handler)
}

func ListPostComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//user := ctx.Value("session.user").(*data.User)
	//post := ctx.Value("post").(*data.Post)
}

func GetComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//user := ctx.Value("session.user").(*data.User)
	//post := ctx.Value("post").(*data.Post)
}

func AddComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//user := ctx.Value("session.user").(*data.User)
	//post := ctx.Value("post").(*data.Post)
}

func DeleteComment(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	//user := ctx.Value("session.user").(*data.User)
	//post := ctx.Value("post").(*data.Post)
}
