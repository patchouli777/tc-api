package livestream

import (
	"log/slog"
	"main/internal/lib/sl"
	"main/pkg/client"
	"net/http"
	"os"
)

type LivestreamClient struct {
	base    *client.BaseClient
	BaseURL string
}

func NewClient() *LivestreamClient {
	return &LivestreamClient{
		base: &client.BaseClient{
			Log:    slog.New(slog.NewTextHandler(os.Stdout, nil)),
			Client: &http.Client{},
		},
		BaseURL: "localhost:8090/livestreams/"}
}

// func (c *LivestreamClient) Delete(channel string) (*http.Response, error) {return nil, errors.New("not implemented")}

// func (c *LivestreamClient) Post(channel string) (*http.Response, error) {return nil, errors.New("not implemented")}

func (c *LivestreamClient) Get(channel string) (*http.Response, error) {
	r := GetRequest{Channel: channel}

	req, err := c.base.Get(c.BaseURL + r.Channel)
	if err != nil {
		c.base.Log.Error("unable to create presence request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) List() (*http.Response, error) {
	req, err := c.base.Get(c.BaseURL)
	if err != nil {
		c.base.Log.Error("unable to create list request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.base.Client.Do(req)
}

func (c *LivestreamClient) Patch(username, title, link string) (*http.Response, error) {
	data := PatchRequest{Title: &title, CategoryLink: &link}

	req, err := c.base.Patch(c.BaseURL+username, data)
	if err != nil {
		c.base.Log.Error("unable to create list request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.base.Client.Do(req)
}
