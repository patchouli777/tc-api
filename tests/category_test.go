package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"twitchy-api/internal/lib/setup"
	api "twitchy-api/pkg/api/category"

	"github.com/stretchr/testify/suite"
)

type CategoryHandlerTestSuite struct {
	suite.Suite
	cl *api.Client
}

func (s *CategoryHandlerTestSuite) SetupSuite() {
	ctx := context.Background()
	s.cl = api.NewClient(ts.URL + "/api/categories/")

	setup.AddTags(ctx, pgpool)
	setup.AddCategories(ctx, app.CategoryRepo)
}

func (s *CategoryHandlerTestSuite) TearDownSuite() {}

func TestCategoryHandlerSuite(t *testing.T) {
	suite.Run(t, new(CategoryHandlerTestSuite))
}

func (s *CategoryHandlerTestSuite) TestCategoryNotFound() {
	ctx := context.Background()

	// get by link
	nonExistent := "path-of-not-existing-category"
	resp, err := s.cl.Get(nonExistent)
	if err != nil {
		s.Fail("unable to send request %v", err)
	}

	s.Equal(http.StatusNotFound, resp.StatusCode)
	c, err := app.CategoryRepo.GetByLink(ctx, nonExistent)
	s.Error(err)
	s.Nil(c)

	// get by id
	nonExistentId := 999999999
	resp, err = s.cl.Get(fmt.Sprintf("%d", nonExistentId))
	if err != nil {
		s.Fail("unable to send request %v", err)
	}

	s.Equal(http.StatusNotFound, resp.StatusCode)
	c, err = app.CategoryRepo.Get(ctx, nonExistentId)
	s.Error(err)
	s.Nil(c)
}

func (s *CategoryHandlerTestSuite) TestCategoryFound() {
	ctx := context.Background()

	res, err := s.cl.GetById("1")
	s.NoError(err)

	categoryById := res.Category
	s.Equal(1, categoryById.Id)

	sameCategoryFromRepo, err := app.CategoryRepo.Get(ctx, 1)
	s.NoError(err)
	s.Equal(categoryById.Link, sameCategoryFromRepo.Link)
	s.Equal(categoryById.Id, int(sameCategoryFromRepo.Id))

	res, err = s.cl.GetByLink("apex")
	s.NoError(err)

	categoryByLink := res.Category
	s.Equal("apex", categoryByLink.Link)

	sameCategoryFromRepo, err = app.CategoryRepo.GetByLink(ctx, "apex")
	s.NoError(err)
	s.Equal(categoryByLink.Id, int(sameCategoryFromRepo.Id))
	s.Equal(categoryByLink.Link, sameCategoryFromRepo.Link)
}
