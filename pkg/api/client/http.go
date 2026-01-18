package baseclient

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client struct {
	Client *http.Client
}

func NewClient() *Client {
	return &Client{Client: &http.Client{}}
}

func (c *Client) Get(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) Post(url string, data any) (*http.Request, error) {
	bs, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}

func (c *Client) Patch(url string, data any) (*http.Request, error) {
	bs, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}

func (c *Client) Delete(url string, data any) (*http.Request, error) {
	bs, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}
