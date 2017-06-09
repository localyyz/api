package workers

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/goware/lg"
	"upper.io/db.v3"
)

// PromoWorker expires promotions and claims
func PromoEndWorker() {
	lg.Info("starting promo ender")
	now := data.GetTimeUTCPointer()

	promos, err := data.DB.Promo.FindAll(
		db.Cond{
			"end_at <": now,
			"status":   data.PromoStatusActive,
		},
	)
	if err != nil {
		lg.Warnf("failed to query expired promos. Err: %+v", err)
		return
	}

	promoIDs := make([]int64, len(promos))
	for i, p := range promos {
		promoIDs[i] = p.ID
	}
	if len(promoIDs) == 0 {
		return
	}

	// update claims to expired
	q := data.DB.Update("claims").Set(
		"status", data.ClaimStatusExpired,
		"updated_at", now,
	).Where("promo_id IN ?", promoIDs)
	if _, err := q.Exec(); err != nil {
		lg.Warnf("failed to expire claim: %v", err)
	}

	for _, p := range promos {
		p.Status = data.PromoStatusCompleted
		if err := data.DB.Promo.Save(p); err != nil {
			lg.Warnf("failed to expire promotions: %v", err)
		}
	}
}

func PromoStartWorker() {
	lg.Info("starting promo starter")
	now := data.GetTimeUTCPointer()

	promos, err := data.DB.Promo.FindAll(
		db.Cond{
			"start_at <": now,
			"status":     data.PromoStatusScheduled,
		},
	)
	if err != nil {
		lg.Warnf("failed to query starting promos. Err: %+v", err)
		return
	}

	for _, p := range promos {
		p.Status = data.PromoStatusActive
		if err := data.DB.Promo.Save(p); err != nil {
			lg.Warnf("failed to start promotions: %v", err)
		}
	}

}

// Find expired promotions. Recreate them.
//func RefreshPromoWorker() {
//lg.Info("starting refresher worker")

//var promos []*data.Promo
//err := data.DB.Select(
//db.Raw("distinct on (place_id) *"),
//).From("promos").Where(
//db.Cond{"end_at between": db.Raw("now()::date - interval '1 day' and now()::date")},
//).All(&promos)
//if err != nil {
//lg.Warnf("promo refresh worker: %+v", err)
//return
//}

//for _, p := range promos {
//p.ID = 0 // new promotion

//dayDiff := (*p.EndAt).Sub(*p.StartAt)
//p.StartAt = data.GetTimeUTCPointer()
//endAt := (*p.StartAt).Add(dayDiff)
//p.EndAt = &endAt

//if err := data.DB.Promo.Save(p); err != nil {
//lg.Warnf("promo refresh worker: %+v", err)
//}
//}
//}
