package client

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type BaseClient struct {
	Client *http.Client
}

func (c *BaseClient) Get(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *BaseClient) Post(url string, data any) (*http.Request, error) {
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

func (c *BaseClient) Patch(url string, data any) (*http.Request, error) {
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

func (c *BaseClient) Delete(url string, data any) (*http.Request, error) {
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
