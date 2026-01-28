package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	livestreamAPI "twitchy-api/pkg/api/livestream"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type LivestreamHandlerTestSuite struct {
	suite.Suite
	cl *livestreamAPI.Client
}

func (s *LivestreamHandlerTestSuite) SetupSuite() {
	s.cl = livestreamAPI.NewClient(ts.URL + "/api/livestreams/")
}
func TestLivestreamHandlerSuite(t *testing.T) {
	// suite.Run(t, new(LivestreamHandlerTestSuite))
}

func (s *LivestreamHandlerTestSuite) TestGetLivestream() {
	resp, err := s.cl.Get("/this_user_should_not_exist")
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	s.Equal(http.StatusBadRequest, resp.StatusCode)

	livestreamID := 2
	response, err := s.cl.Get(fmt.Sprintf("/%d", livestreamID))
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	defer response.Body.Close() // nolint

	require.Equal(s.T(), http.StatusOK, response.StatusCode)

	var res livestreamAPI.GetResponse
	if err = json.NewDecoder(response.Body).Decode(&res); err != nil {
		s.Fail("unable to parse response %v", err)
	}
	s.Equal(livestreamID, res.Id)
}
