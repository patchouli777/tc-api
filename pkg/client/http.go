package client

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"main/internal/lib/sl"
	"net/http"
)

type BaseClient struct {
	Log    *slog.Logger
	Client *http.Client
}

func (c *BaseClient) Get(url string) (*http.Request, error) {
	const op = "lib.client.BaseClient.Get"

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	if err != nil {
		c.Log.Error("unable to create request", slog.String("op", op), sl.Err(err))
		return nil, err
	}

	return req, nil
}

func (c *BaseClient) Post(url string, data any) (*http.Request, error) {
	const op = "lib.client.BaseClient.Post"

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

func (c *BaseClient) Patch(url string, data any) (*http.Request, error) {
	const op = "lib.client.BaseClient.Patch"

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

func (c *BaseClient) Delete(url string, data any) (*http.Request, error) {
	const op = "lib.client.BaseClient.Delete"

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
