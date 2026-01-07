package livestream

import (
	"log/slog"
	"main/internal/lib/sl"
	"main/pkg/client"
	"net/http"
)

type LivestreamClient struct {
	base    *client.BaseClient
	BaseURL string
}

func NewLivestreamClient(log *slog.Logger) *LivestreamClient {
	return &LivestreamClient{
		base: &client.BaseClient{
			Log:    log,
			Client: &http.Client{},
		},
		BaseURL: "http://localhost:8090/api/livestreams"}
}

func (c *LivestreamClient) Get(channel string) (*http.Response, error) {
	r := GetRequest{Channel: channel}

	req, err := c.base.Get(c.BaseURL + r.Channel)
	if err != nil {
		c.base.Log.Error("unable to create presence request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) List() (*http.Response, error) {
	req, err := c.base.Get(c.BaseURL)
	if err != nil {
		c.base.Log.Error("unable to create list request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) Patch(username, title, link string) (*http.Response, error) {
	data := PatchRequest{Title: &title, CategoryLink: &link}

	req, err := c.base.Patch(c.BaseURL+username, data)
	if err != nil {
		c.base.Log.Error("unable to create list request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}
