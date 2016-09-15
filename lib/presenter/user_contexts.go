package presenter

type UserContext struct {
	Commented bool `json:"commented"`
	Liked     bool `json:"liked"`
	Promoted  bool `json:"promoted"`
	PeekedAt  bool `json:"peekedAt"`
}
