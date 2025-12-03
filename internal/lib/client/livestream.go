package client

import (
	"main/internal/endpoint/livestream"
	"main/internal/lib/sl"
	"net/http"
)

func (c *CustomClient) LivestreamsGet(channel string) (*http.Response, error) {
	r := livestream.GetRequest{Channel: channel}

	req, err := c.Get(c.BaseURL + r.Channel)
	if err != nil {
		c.Log.Error("unable to create presence request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.Client.Do(req)
}

func (c *CustomClient) LivestreamsList() (*http.Response, error) {
	req, err := c.Get(c.BaseURL)
	if err != nil {
		c.Log.Error("unable to create list request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.Client.Do(req)
}

func (c *CustomClient) LivestreamsStart(username string, title string, category string) (*http.Response, error) {
	r := livestream.PostRequest{
		Title:        title,
		CategoryLink: category,
	}

	req, err := c.Post(c.BaseURL+username, r)
	if err != nil {
		c.Log.Error("unable to create post request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.Client.Do(req)
}

func (c *CustomClient) LivestreamsStop(username string) (*http.Response, error) {
	req, err := c.Delete(c.BaseURL+username, nil)
	if err != nil {
		c.Log.Error("unable to create delete request", sl.Err(err))
		return nil, err
	}
	defer req.Body.Close()

	return c.Client.Do(req)
}
