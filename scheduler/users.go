package scheduler

func (h *Handler) ScheduleWelcomeEmail() {
	h.wg.Add(1)
	defer h.wg.Done()

	// fetch all users who was recently signed up
}
