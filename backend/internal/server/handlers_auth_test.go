package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/nhcathcart/oops-my-turn/backend/internal/repositories"
	models "github.com/nhcathcart/oops-my-turn/backend/models/generated"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestHandleGoogleCallback(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      *GoogleCallbackInput
		server     *Server
		status     int
		assert     func(t *testing.T, output *GoogleCallbackOutput)
		assertRepo func(t *testing.T, repo *stubUserRepository)
	}{
		{
			name: "invalid state",
			input: &GoogleCallbackInput{
				State:      "request-state",
				Code:       "code-123",
				OAuthState: "cookie-state",
			},
			server: newTestServer(
				Options{JWTSecret: "secret", FrontendURL: "https://frontend.example", BackendURL: "https://backend.example"},
				repositories.Repositories{User: &stubUserRepository{}},
				&stubGoogleOAuthClient{},
			),
			status: http.StatusBadRequest,
			assertRepo: func(t *testing.T, repo *stubUserRepository) {
				t.Helper()
				require.Zero(t, repo.upsertCalls)
			},
		},
		{
			name: "token exchange failure",
			input: &GoogleCallbackInput{
				State:      "state-123",
				Code:       "code-123",
				OAuthState: "state-123",
			},
			server: newTestServer(
				Options{JWTSecret: "secret", FrontendURL: "https://frontend.example", BackendURL: "https://backend.example"},
				repositories.Repositories{User: &stubUserRepository{}},
				&stubGoogleOAuthClient{
					exchangeErr: errors.New("exchange failed"),
				},
			),
			status: http.StatusBadRequest,
			assertRepo: func(t *testing.T, repo *stubUserRepository) {
				t.Helper()
				require.Zero(t, repo.upsertCalls)
			},
		},
		{
			name: "userinfo failure",
			input: &GoogleCallbackInput{
				State:      "state-123",
				Code:       "code-123",
				OAuthState: "state-123",
			},
			server: newTestServer(
				Options{JWTSecret: "secret", FrontendURL: "https://frontend.example", BackendURL: "https://backend.example"},
				repositories.Repositories{User: &stubUserRepository{}},
				&stubGoogleOAuthClient{
					token:        &oauth2.Token{AccessToken: "access-token"},
					fetchUserErr: errors.New("userinfo failed"),
				},
			),
			status: http.StatusInternalServerError,
			assertRepo: func(t *testing.T, repo *stubUserRepository) {
				t.Helper()
				require.Zero(t, repo.upsertCalls)
			},
		},
		{
			name: "upsert failure",
			input: &GoogleCallbackInput{
				State:      "state-123",
				Code:       "code-123",
				OAuthState: "state-123",
			},
			server: newTestServer(
				Options{JWTSecret: "secret", FrontendURL: "https://frontend.example", BackendURL: "https://backend.example"},
				repositories.Repositories{User: &stubUserRepository{upsertErr: errors.New("db failed")}},
				&stubGoogleOAuthClient{
					token: &oauth2.Token{AccessToken: "access-token"},
					userInfo: &googleUserInfo{
						Sub:        "google-123",
						Email:      "user@example.com",
						GivenName:  "Jane",
						FamilyName: "Doe",
					},
				},
			),
			status: http.StatusInternalServerError,
			assertRepo: func(t *testing.T, repo *stubUserRepository) {
				t.Helper()
				require.Equal(t, 1, repo.upsertCalls)
			},
		},
		{
			name: "success",
			input: &GoogleCallbackInput{
				State:      "state-123",
				Code:       "code-123",
				OAuthState: "state-123",
			},
			server: newTestServer(
				Options{JWTSecret: "secret", FrontendURL: "https://frontend.example/app", BackendURL: "https://backend.example"},
				repositories.Repositories{User: &stubUserRepository{
					user: &models.User{
						ID:        "usr_123",
						Email:     "user@example.com",
						FirstName: "Jane",
						LastName:  "Doe",
					},
				}},
				&stubGoogleOAuthClient{
					token: &oauth2.Token{AccessToken: "access-token"},
					userInfo: &googleUserInfo{
						Sub:        "google-123",
						Email:      "user@example.com",
						GivenName:  "Jane",
						FamilyName: "Doe",
					},
				},
			),
			status: http.StatusFound,
			assert: func(t *testing.T, output *GoogleCallbackOutput) {
				t.Helper()
				require.NotNil(t, output)
				require.Equal(t, "https://frontend.example/app", output.Location)
				require.Len(t, output.SetCookie, 2)

				sessionCookie := output.SetCookie[0]
				require.Equal(t, sessionCookieName, sessionCookie.Name)
				require.NotEmpty(t, sessionCookie.Value)
				require.True(t, sessionCookie.HttpOnly)
				require.True(t, sessionCookie.Secure)
				require.Equal(t, http.SameSiteNoneMode, sessionCookie.SameSite)

				claims, err := ParseJWT("secret", sessionCookie.Value)
				require.NoError(t, err)
				require.Equal(t, "usr_123", claims.Sub)
				require.Equal(t, "user@example.com", claims.Email)
				require.Equal(t, "Jane", claims.FirstName)
				require.Equal(t, "Doe", claims.LastName)

				oauthCookie := output.SetCookie[1]
				require.Equal(t, oauthStateCookie, oauthCookie.Name)
				require.Equal(t, "", oauthCookie.Value)
				require.Equal(t, -1, oauthCookie.MaxAge)
			},
			assertRepo: func(t *testing.T, repo *stubUserRepository) {
				t.Helper()
				require.Equal(t, 1, repo.upsertCalls)
				require.Equal(t, "google-123", repo.googleID)
				require.Equal(t, "user@example.com", repo.email)
				require.Equal(t, "Jane", repo.firstName)
				require.Equal(t, "Doe", repo.lastName)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			output, err := tc.server.handleGoogleCallback(context.Background(), tc.input)

			if tc.status >= http.StatusBadRequest {
				require.Nil(t, output)
				requireHumaStatus(t, err, tc.status)
			} else {
				require.NoError(t, err)
				require.NotNil(t, output)
				tc.assert(t, output)
			}

			repo := tc.server.repos.User.(*stubUserRepository)
			tc.assertRepo(t, repo)
		})
	}
}

func TestHandleGoogleLogin(t *testing.T) {
	t.Parallel()

	server := newTestServer(
		Options{},
		repositories.Repositories{},
		&stubGoogleOAuthClient{
			authCodeURL: "https://accounts.example/auth?state=test-state",
		},
	)

	output, err := server.handleGoogleLogin(context.Background(), &struct{}{})
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, http.StatusFound, output.Status)
	require.Equal(t, "https://accounts.example/auth?state=test-state", output.Location)
	require.Len(t, output.SetCookie, 1)

	stateCookie := output.SetCookie[0]
	require.Equal(t, oauthStateCookie, stateCookie.Name)
	require.NotEmpty(t, stateCookie.Value)
	require.Equal(t, "/", stateCookie.Path)
	require.Equal(t, 600, stateCookie.MaxAge)
	require.True(t, stateCookie.HttpOnly)
	require.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
}

func TestHandleLogout(t *testing.T) {
	t.Parallel()

	server := newTestServer(Options{}, repositories.Repositories{}, nil)

	output, err := server.handleLogout(context.Background(), &struct{}{})
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, sessionCookieName, output.SetCookie.Name)
	require.Equal(t, "", output.SetCookie.Value)
	require.Equal(t, -1, output.SetCookie.MaxAge)
}

func TestHandleMe(t *testing.T) {
	t.Parallel()

	t.Run("unauthorized", func(t *testing.T) {
		t.Parallel()

		server := newTestServer(Options{}, repositories.Repositories{}, nil)

		output, err := server.handleMe(context.Background(), &struct{}{})
		require.Nil(t, output)
		requireHumaStatus(t, err, http.StatusUnauthorized)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		server := newTestServer(Options{}, repositories.Repositories{}, nil)
		ctx := context.WithValue(context.Background(), jwtClaimsKey{}, &JWTClaims{
			Sub:       "usr_123",
			Email:     "user@example.com",
			FirstName: "Jane",
			LastName:  "Doe",
		})

		output, err := server.handleMe(ctx, &struct{}{})
		require.NoError(t, err)
		require.NotNil(t, output)
		require.Equal(t, meResponseBody{
			ID:        "usr_123",
			Email:     "user@example.com",
			FirstName: "Jane",
			LastName:  "Doe",
		}, output.Body)
	})
}

func newTestServer(options Options, repos repositories.Repositories, oauthClient googleOAuthClient) *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	serverOptions := []ServerOption{}
	if oauthClient != nil {
		serverOptions = append(serverOptions, WithGoogleOAuthClient(oauthClient))
	}
	return NewServer(options, nil, logger, repos, Services{}, serverOptions...)
}

func requireHumaStatus(t *testing.T, err error, status int) {
	t.Helper()

	require.Error(t, err)
	statusErr, ok := err.(huma.StatusError)
	require.True(t, ok)
	require.Equal(t, status, statusErr.GetStatus())
}

type stubGoogleOAuthClient struct {
	authCodeURL  string
	token        *oauth2.Token
	userInfo     *googleUserInfo
	exchangeErr  error
	fetchUserErr error
}

func (c *stubGoogleOAuthClient) AuthCodeURL(state string) string {
	if c.authCodeURL != "" {
		return c.authCodeURL
	}
	return "https://accounts.example/auth?state=" + state
}

func (c *stubGoogleOAuthClient) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	if c.exchangeErr != nil {
		return nil, c.exchangeErr
	}
	return c.token, nil
}

func (c *stubGoogleOAuthClient) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	if c.fetchUserErr != nil {
		return nil, c.fetchUserErr
	}
	return c.userInfo, nil
}

type stubUserRepository struct {
	user        *models.User
	upsertErr   error
	upsertCalls int
	googleID    string
	email       string
	firstName   string
	lastName    string
}

func (r *stubUserRepository) Upsert(ctx context.Context, googleID, email, firstName, lastName string) (*models.User, error) {
	r.upsertCalls++
	r.googleID = googleID
	r.email = email
	r.firstName = firstName
	r.lastName = lastName

	if r.upsertErr != nil {
		return nil, r.upsertErr
	}

	if r.user != nil {
		return r.user, nil
	}

	return &models.User{
		ID:        "usr_default",
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}
