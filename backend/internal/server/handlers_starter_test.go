//go:build integration

package server_test

import (
	"context"
	"net/http"
)

func (s *IntegrationTestSuite) TestGetHello() {
	ctx := context.Background()

	resp, err := s.client.GetHelloWithResponse(ctx)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Require().NotNil(resp.JSON200)
	s.Equal("Hello from the starter template.", resp.JSON200.Message)
}

func (s *IntegrationTestSuite) TestGetMe() {
	ctx := context.Background()

	resp, err := s.client.GetMeWithResponse(ctx)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())
	s.Require().NotNil(resp.JSON200)
	s.Equal("test-user", resp.JSON200.Id)
	s.Equal("test@example.com", resp.JSON200.Email)
	s.Equal("Test", resp.JSON200.FirstName)
	s.Equal("User", resp.JSON200.LastName)
}

func (s *IntegrationTestSuite) TestLogout() {
	ctx := context.Background()

	resp, err := s.client.LogoutWithResponse(ctx)
	s.Require().NoError(err)
	s.Equal(http.StatusNoContent, resp.StatusCode())
}
