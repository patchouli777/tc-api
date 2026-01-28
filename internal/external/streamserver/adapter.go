package streamserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Adapter struct {
	endpoint string
}

func NewAdapter(endpoint string) *Adapter {
	return &Adapter{endpoint: endpoint + "streams"}
}

func (u *Adapter) List(ctx context.Context, start, count int) (*ListResponse, error) {
	cl := &http.Client{}
	response, err := cl.Get(fmt.Sprintf("%s?start=%d&count=%d", u.endpoint, start, count))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp ListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (u *Adapter) Get(ctx context.Context, channel string) (*GetResponse, error) {
	cl := &http.Client{}
	response, err := cl.Get(u.endpoint + "/" + channel)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp GetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
