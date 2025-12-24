package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ListStreamServerResponse struct {
	Code    int          `json:"code"`
	Server  string       `json:"server"`
	Service string       `json:"service"`
	Pid     string       `json:"pid"`
	Streams []StreamData `json:"streams"`
}

type GetStreamServerResponse struct {
	Code    int        `json:"code"`
	Server  string     `json:"server"`
	Service string     `json:"service"`
	Pid     string     `json:"pid"`
	Stream  StreamData `json:"stream"`
}

type PostStreamServerRequest struct {
	Channel string `json:"channel"`
}

func Get(srv *StreamServerMock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		data, err := srv.Get(r.Context(), id)
		if err != nil {
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(GetStreamServerResponse{
			Stream: StreamData{
				ID:      data.channel,
				Clients: data.viewers,
				Name:    data.channel}})
	}
}

func Post(srv *StreamServerMock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PostStreamServerRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = srv.Start(r.Context(), req.Channel)
		if err != nil {
			fmt.Println(err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	// id := r.PathValue("id")
	fmt.Println("delete not implemented")
}

func List(srv *StreamServerMock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := srv.List(r.Context())
		if err != nil {
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		streams := make([]StreamData, len(data))
		for i := range len(data) {
			streams[i] = StreamData{
				ID:      data[i].channel,
				Clients: data[i].viewers,
				Name:    data[i].channel}
		}

		json.NewEncoder(w).Encode(ListStreamServerResponse{Streams: streams})
	}
}
