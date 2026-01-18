package mock

import (
	"main/internal/external/streamserver"
	baseclient "main/pkg/api/client"
	"net/http"
)

type StreamServerClient struct {
	base    *baseclient.Client
	BaseURL string
}

func NewStreamServerClient(baseUrl string) *StreamServerClient {
	return &StreamServerClient{
		base:    baseclient.NewClient(),
		BaseURL: baseUrl}
}

func (c *StreamServerClient) Start(username string) (*http.Response, error) {
	req, err := c.base.Post(c.BaseURL+"/streams", streamserver.PostRequest{Channel: username})
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *StreamServerClient) Stop(username string) (*http.Response, error) {
	req, err := c.base.Delete(c.BaseURL+username, nil)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}
