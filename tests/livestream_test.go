package test

import (
	"encoding/json"
	"log/slog"
	"main/internal/endpoint/livestream"
	"main/internal/lib/client"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLivestream(t *testing.T) {
	slogger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))

	cc := &client.CustomClient{BaseURL: ts.URL + "/livestreams",
		Client: ts.Client(),
		Log:    slogger}

	resp, err := cc.LivestreamsGet("/this_user_should_not_exist")
	if err != nil {
		t.Fatalf("unable to send request %v", err)
	}

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	response, err := cc.LivestreamsGet("/user1")
	if err != nil {
		t.Fatalf("unable to send request %v", err)
	}
	defer response.Body.Close()

	assert.Equal(t, http.StatusOK, response.StatusCode)

	var res livestream.GetResponse
	if err = json.NewDecoder(response.Body).Decode(&res); err != nil {
		t.Fatalf("unable to parse response %v", err)
	}

	assert.Equal(t, "user1", res.Username)
}
