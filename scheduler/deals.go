package scheduler

func (h *Handler) ScheduleDeals() {
	h.wg.Add(1)
	defer h.wg.Done()

	// expire collections
	h.DB.Exec(`UPDATE collections SET status = 3 WHERE lightning = true AND NOW() at time zone 'utc' > end_at and status = 2`)
	// activate collections
	h.DB.Exec(`UPDATE collections SET status = 2 WHERE lightning = true AND NOW() at time zone 'utc' > start_at and status = 1`)
	// expire user deals
	h.DB.Exec(`UPDATE user_deals SET status = 3 WHERE NOW() at time zone 'utc' > end_at and status = 2`)
}
