package livestream

import (
	"errors"
	"log/slog"
	"main/internal/lib/sl"
	"main/pkg/client"
	"net/http"
	"os"
)

type StreamServerClient struct {
	base    *client.BaseClient
	BaseURL string
}

func NewStreamServerClient() *StreamServerClient {
	return &StreamServerClient{
		base: &client.BaseClient{
			Log:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			Client: &http.Client{},
		},
		BaseURL: "http://localhost:1985/api/v1/streams"}
}

func (c *StreamServerClient) Start(username string) (*http.Response, error) {
	type Data struct {
		Channel string `json:"channel"`
	}

	req, err := c.base.Post(c.BaseURL, Data{Channel: username})
	if err != nil {
		c.base.Log.Error("unable to create post request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.base.Client.Do(req)
}

func (c *StreamServerClient) Stop(username string) (*http.Response, error) {
	req, err := c.base.Delete(c.BaseURL+username, nil)
	if err != nil {
		c.base.Log.Error("unable to create delete request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.base.Client.Do(req)
}

func (c *StreamServerClient) List(username string) (*http.Response, error) {
	return nil, errors.New("not implemented")
}

func (c *StreamServerClient) Get(username string) (*http.Response, error) {
	return nil, errors.New("not implemented")
}
