package mock

type livestream struct {
	channel string
	viewers int
}

type mockError struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Op      string `json:"op"`
}
