package worker

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	db "upper.io/db.v2"
)

// PromoWorker expires promotions and claims
func PromoWorker() {
	lg.Info("starting promo worker")

	now := data.GetTimeUTCPointer()
	query := data.DB.
		Select(db.Raw("cl.*")).
		From("claims cl").
		LeftJoin("promos p").
		On("cl.promo_id = p.id").
		Where(
			db.Cond{
				"cl.status IN": []data.ClaimStatus{
					data.ClaimStatusActive,
					data.ClaimStatusSaved,
					data.ClaimStatusPeeked,
				},
				"p.end_at <": now,
			})

	var expiredClaims []*data.Claim
	if err := query.All(&expiredClaims); err != nil {
		lg.Warn("promo worker errored: %v", err)
		return
	}

	for _, exp := range expiredClaims {
		exp.Status = data.ClaimStatusExpired
		exp.UpdatedAt = data.GetTimeUTCPointer()

		if err := data.DB.Claim.Save(exp); err != nil {
			lg.Warn("failed to expire claim: %v", err)
		}
	}

}

// Find expired promotions. Recreate them.
func RefreshPromoWorker() {
	lg.Info("starting refresher worker")

	var promos []*data.Promo
	err := data.DB.Select(
		db.Raw("distinct on (place_id) *"),
	).From("promos").Where(
		db.Cond{"end_at between": db.Raw("now()::date - interval '1 day' and now()::date")},
	).All(&promos)
	if err != nil {
		lg.Warnf("promo refresh worker: %+v", err)
		return
	}

	for _, p := range promos {
		p.ID = 0 // new promotion

		dayDiff := (*p.EndAt).Sub(*p.StartAt)
		p.StartAt = data.GetTimeUTCPointer()
		endAt := (*p.StartAt).Add(dayDiff)
		p.EndAt = &endAt

		if err := data.DB.Promo.Save(p); err != nil {
			lg.Warnf("promo refresh worker: %+v", err)
		}
	}
}
