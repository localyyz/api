package user

import (
	"context"
	"net/http"
	"strconv"

	db "upper.io/db.v3"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/data/presenter"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/pressly/chi"
	"github.com/pressly/chi/render"
)

func PaymentMethodCtx(next http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		paymentID, err := strconv.ParseInt(chi.URLParam(r, "paymentID"), 10, 64)
		if err != nil {
			render.Render(w, r, api.ErrBadID)
			return
		}
		ctx := r.Context()
		user := ctx.Value("session.user").(*data.User)

		payment, err := data.DB.PaymentMethod.FindOne(
			db.Cond{
				"id":      paymentID,
				"user_id": user.ID,
			},
		)
		if err != nil {
			render.Respond(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, "payment", payment)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(handler)
}

func ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("session.user").(*data.User)

	payments, err := data.DB.PaymentMethod.FindAllByUserID(user.ID)
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	presented := presenter.NewPaymentMethodList(ctx, payments)
	if err := render.RenderList(w, r, presented); err != nil {
		render.Respond(w, r, err)
	}
}

func GetPaymentMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payment := ctx.Value("payment").(*data.PaymentMethod)
	render.Render(w, r, presenter.NewPaymentMethod(ctx, payment))
}

func CreatePaymentMethod(w http.ResponseWriter, r *http.Request) {
}

func UpdatePaymentMethod(w http.ResponseWriter, r *http.Request) {
}

func RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
}
