package mock

import (
	"encoding/json"
	"fmt"
	"main/internal/external/streamserver"
	baseclient "main/pkg/api/client"
	"net/http"
)

type serverState struct {
	streams *repository
	subs    []string
	cl      *baseclient.Client
}

type handler struct {
	state *serverState
}

func (s *handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	data, err := s.state.streams.Get(r.Context(), id)
	if err != nil {
		json.NewEncoder(w).Encode(mockError{Error: err.Error(), Op: "Get"})
		return
	}

	json.NewEncoder(w).Encode(streamserver.GetResponse{
		Stream: streamserver.StreamData{
			ID:      data.channel,
			Clients: data.viewers,
			Name:    data.channel}})
}

func (s *handler) Post(w http.ResponseWriter, r *http.Request) {
	var req streamserver.PostRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.state.streams.Start(r.Context(), req.Channel)
	if err != nil {
		fmt.Println(err)
	}

	for _, s := range s.state.subs {
		cl := baseclient.NewClient()
		cl.Post(s, streamserver.StreamEventPayload{
			Action: "on_publish",
		})
	}

	w.WriteHeader(http.StatusNoContent)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// id := r.PathValue("id")
	fmt.Println("delete not implemented")
}

func (s *handler) List(w http.ResponseWriter, r *http.Request) {
	data, err := s.state.streams.List(r.Context())
	if err != nil {
		json.NewEncoder(w).Encode(mockError{Error: err.Error(), Op: "List"})
		return
	}

	streams := make([]streamserver.StreamData, len(data))
	for i := range len(data) {
		streams[i] = streamserver.StreamData{
			ID:      data[i].channel,
			Clients: data[i].viewers,
			Name:    data[i].channel}
	}

	json.NewEncoder(w).Encode(streamserver.ListResponse{Streams: streams})
}
