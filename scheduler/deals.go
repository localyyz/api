package scheduler

import (
	"time"

	"github.com/pressly/lg"
)

func (h *Handler) ScheduleDeals() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_schedule_deals running...")
	defer func() {
		lg.Infof("job_schedule_deals finished in %s", time.Since(s))
	}()

	// expire collections
	h.DB.Exec(`UPDATE collections SET status = 3 WHERE lightning = true AND NOW() at time zone 'utc' > end_at and status = 2`)
	// activate collections
	h.DB.Exec(`UPDATE collections SET status = 2 WHERE lightning = true AND NOW() at time zone 'utc' > start_at and status = 1`)
	// expire user deals
	h.DB.Exec(`UPDATE user_deals SET status = 3 WHERE NOW() at time zone 'utc' > end_at and status = 2`)
}
