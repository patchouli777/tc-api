package livestream

import (
	"net/http"
	"twitchy-api/internal/lib/null"
	baseclient "twitchy-api/pkg/api/client"
)

type Client struct {
	base    *baseclient.Client
	BaseURL string
}

func NewClient(url string) *Client {
	return &Client{
		base:    baseclient.NewClient(),
		BaseURL: url}
}

func (c *Client) Get(channel string) (*http.Response, error) {
	r := GetRequest{Channel: channel}

	req, err := c.base.Get(c.BaseURL + r.Channel)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *Client) List() (*http.Response, error) {
	req, err := c.base.Get(c.BaseURL)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *Client) Patch(username, title string, categoryId int) (*http.Response, error) {
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
