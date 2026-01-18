package mock

import (
	"encoding/json"
	"main/internal/external/streamserver"
	"net/http"
	"slices"
)

func (s *handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var p streamserver.SubscribeRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		json.NewEncoder(w).Encode(mockError{Error: err.Error(), Op: "Subscribe"})
		return
	}

	s.state.subs = append(s.state.subs, p.CallbackURL)

	json.NewEncoder(w).Encode(streamserver.SubscribeResponse{Success: true})
}

func (s *handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var p streamserver.UnsubscribeRequest
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		json.NewEncoder(w).Encode(mockError{Error: err.Error(), Op: "Unsubscribe"})
		return
	}

	s.state.subs = slices.DeleteFunc(s.state.subs, func(e string) bool {
		return e == p.CallbackURL
	})

	json.NewEncoder(w).Encode(streamserver.SubscribeResponse{Success: true})
}
