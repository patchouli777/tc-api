package category

import (
	"encoding/json"
	"net/http"
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

func (c *Client) Get(categoryLink string) (*http.Response, error) {
	r := GetRequest{CategoryLink: categoryLink}

	req, err := c.base.Get(c.BaseURL + r.CategoryLink)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close() // nolint

	return c.base.Client.Do(req)
}

func (c *Client) GetByLink(categoryLink string) (GetResponse, error) {
	r := GetRequest{CategoryLink: categoryLink}

	req, err := c.base.Get(c.BaseURL + r.CategoryLink)
	if err != nil {
		return GetResponse{}, err
	}
	defer req.Body.Close() // nolint

	resp, err := c.base.Client.Do(req)
	if err != nil {
		return GetResponse{}, err
	}

	var res GetResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return GetResponse{}, err
	}

	return res, nil
}

func (c *Client) GetById(id string) (GetResponse, error) {
	req, err := c.base.Get(c.BaseURL + id)
	if err != nil {
		return GetResponse{}, err
	}
	defer req.Body.Close() // nolint

	resp, err := c.base.Client.Do(req)
	if err != nil {
		return GetResponse{}, err
	}

	var res GetResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return GetResponse{}, err
	}

	return res, nil
}
