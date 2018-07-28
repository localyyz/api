package scheduler

import (
	"time"

	"github.com/pressly/lg"
)

func (h *Handler) ScheduleWelcomeEmail() {
	h.wg.Add(1)
	defer h.wg.Done()

	s := time.Now()
	lg.Info("job_welcome_email running...")
	defer func() {
		lg.Infof("job_welcome_email finished in %s", time.Since(s))
	}()

	if h.Environment != "production" {
		// do not send if not production
		return
	}
	// fetch all users who was recently signed up
}
