package client

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"main/internal/lib/sl"
	"net/http"
)

type CustomClient struct {
	Log     *slog.Logger
	Client  *http.Client
	BaseURL string
}

func (c *CustomClient) Get(url string) (*http.Request, error) {
	const op = "lib.client.CustomClient.Get"

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		c.Log.Error("unable to create request", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	return req, nil
}

func (c *CustomClient) Post(url string, data any) (*http.Request, error) {
	const op = "lib.client.CustomClient.Post"

	bs, err := json.Marshal(data)
	if err != nil {
		c.Log.Error("unable marshal request data", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bs))
	if err != nil {
		c.Log.Error("unable to create request", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}

func (c *CustomClient) Patch(url string, data any) (*http.Request, error) {
	const op = "lib.client.CustomClient.Patch"

	bs, err := json.Marshal(data)
	if err != nil {
		c.Log.Error("unable marshal request data", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(bs))
	if err != nil {
		c.Log.Error("unable to create request", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}

func (c *CustomClient) Delete(url string, data any) (*http.Request, error) {
	const op = "lib.client.CustomClient.Delete"

	bs, err := json.Marshal(data)
	if err != nil {
		c.Log.Error("unable marshal request data", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(bs))
	if err != nil {
		c.Log.Error("unable to create request", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, err
}
