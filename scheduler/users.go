package scheduler

func (h *Handler) ScheduleWelcomeEmail() {
	h.wg.Add(1)
	defer h.wg.Done()

	if h.Environment != "production" {
		// do not send if not production
		return
	}
	// fetch all users who was recently signed up
}
