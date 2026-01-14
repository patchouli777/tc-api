package test

import (
	"encoding/json"
	api "main/pkg/api/category"
	"net/http"

	"github.com/stretchr/testify/suite"
)

type CategoryHandlerTestSuite struct {
	suite.Suite
	cl *api.Client
}

func (s *CategoryHandlerTestSuite) SetupSuite() {
	s.cl = api.NewClient(log, ts.URL+"/api/categories/")
}

func (s *CategoryHandlerTestSuite) TestGetCategory() {
	resp, err := s.cl.Get("path-of-not-existing-category")
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	s.Equal(http.StatusNotFound, resp.StatusCode)

	resp, err = s.cl.Get("999999999")
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *CategoryHandlerTestSuite) TestGetCategory_ID() {
	resp, err := s.cl.Get("1")
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	s.Equal(http.StatusAccepted, resp.StatusCode)

	var res api.GetResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		s.Fail("unable to parse response %v", err)
	}

	category := res.Category
	link := category.Link
	s.Equal(int32(1), category.Id)
	s.Len(category.Tags, 2)

	resp, err = s.cl.Get(link)
	if err != nil {
		s.Fail("unable to send request %v", err)
	}
	s.Equal(http.StatusAccepted, resp.StatusCode)

	var sameRes api.GetResponse
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		s.Fail("unable to parse response %v", err)
	}

	sameCategory := sameRes.Category
	s.Equal(category, sameCategory)
}

// func TestGetCategory_Link(t *testing.T) {
// 	client := api.NewClient(testLog, server.URL+"/api/categories")
// 	ctx := context.Background()

// 	// Test link lookup
// 	cat, err := client.GetByLink(ctx, "path-of-exile")
// 	require.NoError(t, err)
// 	assert.Equal(t, int32(123), cat.Category.Id)
// 	assert.Equal(t, "Path of Exile", cat.Category.Name)

// 	// Test invalid link (not found)
// 	_, err = client.GetByLink(ctx, "nonexistent")
// 	require.Error(t, err)
// 	assert.Equal(t, http.StatusNotFound, client.lastStatus)
// }

// func TestGetCategory_TagConversion(t *testing.T) {
// 	server := setupTestServer(t)
// 	defer server.Close()

// 	// // Setup repo mock with CategoryTags
// 	// mockRepo.EXPECT().GetByLink(...).Return(&repo.Category{
// 	//     Id:        1,
// 	//     Tags:      repo.CategoryTags{{Id: 10, Name: "action"}, {Id: 20, Name: "rpg"}},
// 	//     // ... other fields
// 	// }, nil)

// 	client := api.NewClient(testLog, server.URL+"/api/categories")
// 	cat, err := client.GetByLink(context.Background(), "test")
// 	require.NoError(t, err)

// 	// Verify internal CategoryTags → c.CategoryTag conversion
// 	assert.Len(t, cat.Category.Tags, 2)
// 	assert.Equal(t, c.CategoryTag{Id: 10, Name: "action"}, cat.Category.Tags[0])
// }

// func TestGetCategory_InvalidIdentifier(t *testing.T) {
// 	server := setupTestServer(t)
// 	defer server.Close()

// 	resp, err := http.Get(server.URL + "/api/categories/invalid")
// 	require.NoError(t, err)
// 	assert.Equal(t, http.StatusOK, resp.StatusCode) // Tries atoi first, then link

// 	// If both fail → 404 (repo returns errNotFound)
// 	// Mock repo to return errNotFound for "invalid"
// }
