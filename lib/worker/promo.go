package worker

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	db "upper.io/db.v2"
)

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

		if err := data.DB.Claim.Save(exp); err != nil {
			lg.Warn("failed to expire claim: %v", err)
		}
	}

}
