package test

import (
	"testing"
)

func TestGetLivestream(t *testing.T) {
	// cl := livestreamAPI.NewLivestreamClient(log)

	// resp, err := cl.Get("/this_user_should_not_exist")
	// if err != nil {
	// 	t.Fatalf("unable to send request %v", err)
	// }

	// assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	// response, err := cl.Get("/user1")
	// if err != nil {
	// 	t.Fatalf("unable to send request %v", err)
	// }
	// defer response.Body.Close() // nolint

	// assert.Equal(t, http.StatusOK, response.StatusCode)

	// var res livestreamAPI.GetResponse
	// if err = json.NewDecoder(response.Body).Decode(&res); err != nil {
	// 	t.Fatalf("unable to parse response %v", err)
	// }

	// assert.Equal(t, "user1", res.Username)
}
