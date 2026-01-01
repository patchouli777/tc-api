package main

import (
	"encoding/json"
	"fmt"
	"main/pkg/api/streamserver"
	"net/http"
)

func Get(srv *StreamServerMock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		data, err := srv.Get(r.Context(), id)
		if err != nil {
			json.NewEncoder(w).Encode(struct{ Error string }{Error: err.Error()})
			return
		}

		json.NewEncoder(w).Encode(streamserver.GetResponse{
			Stream: streamserver.StreamData{
				ID:      data.channel,
				Clients: data.viewers,
				Name:    data.channel}})
	}
}

func Post(srv *StreamServerMock) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req streamserver.PostRequest

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

		streams := make([]streamserver.StreamData, len(data))
		for i := range len(data) {
			streams[i] = streamserver.StreamData{
				ID:      data[i].channel,
				Clients: data[i].viewers,
				Name:    data[i].channel}
		}

		json.NewEncoder(w).Encode(streamserver.ListResponse{Streams: streams})
	}
}
