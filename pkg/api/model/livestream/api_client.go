package livestream

import (
	"log/slog"
	"main/internal/lib/null"
	"main/pkg/api/client"
	"net/http"
)

type LivestreamClient struct {
	base    *client.BaseClient
	BaseURL string
}

func NewLivestreamClient(log *slog.Logger) *LivestreamClient {
	return &LivestreamClient{
		base: &client.BaseClient{
			Client: &http.Client{},
		},
		BaseURL: "http://localhost:8090/api/livestreams"}
}

func (c *LivestreamClient) Get(channel string) (*http.Response, error) {
	r := GetRequest{Channel: channel}

	req, err := c.base.Get(c.BaseURL + r.Channel)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) List() (*http.Response, error) {
	req, err := c.base.Get(c.BaseURL)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) Patch(username, title string, categoryId int) (*http.Response, error) {
	data := PatchRequest{
		Title: null.String{
			Value:    title,
			Explicit: true,
		},
		CategoryId: null.Int{
			Value:    categoryId,
			Explicit: true,
		}}

	req, err := c.base.Patch(c.BaseURL+username, data)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}
