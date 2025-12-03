package presence

type PresenceResponse struct {
	Watching bool `json:"watching"`
}

type PresenceRequest struct {
	Channel string `json:"channel"`
}
