package streamserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"main/internal/lib/sl"
	"net/http"
)

type Adapter struct {
	log      *slog.Logger
	endpoint string
}

func NewAdapter(log *slog.Logger, endpoint string) *Adapter {
	return &Adapter{
		log:      log,
		endpoint: endpoint + "streams"}
}

func (u *Adapter) List(ctx context.Context, start, count int) (*ListResponse, error) {
	const op = "livestream.Adapter.List"

	cl := &http.Client{}
	response, err := cl.Get(fmt.Sprintf("%s?start=%d&count=%d", u.endpoint, start, count))
	if err != nil {
		u.log.Error("get livestreams", sl.Err(err), sl.Op(op))
		return nil, err
	}
	defer response.Body.Close() // nolint

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.log.Error("parse response", sl.Err(err), sl.Op(op))
		return nil, err
	}

	var resp ListResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		u.log.Error("unmarshal response", sl.Err(err), sl.Op(op))
		return nil, err
	}

	return &resp, nil
}

func (u *Adapter) Get(ctx context.Context, channel string) (*GetResponse, error) {
	const op = "livestream.Adapter.Get"

	cl := &http.Client{}
	response, err := cl.Get(u.endpoint + "/" + channel)
	if err != nil {
		u.log.Error("get livestream", sl.Err(err), sl.Op(op), slog.String("channel", channel))
		return nil, err
	}
	defer response.Body.Close() // nolint

	body, err := io.ReadAll(response.Body)
	if err != nil {
		u.log.Error("parse response", sl.Err(err), sl.Op(op))
		return nil, err
	}

	var resp GetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		u.log.Error("unmarshal response", sl.Err(err), sl.Op(op))
		return nil, err
	}

	return &resp, nil
}
