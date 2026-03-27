package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthStateCookie = "oauth_state"

type googleOAuthClient interface {
	AuthCodeURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	FetchUserInfo(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error)
}

type realGoogleOAuthClient struct {
	config *oauth2.Config
}

func (c *realGoogleOAuthClient) AuthCodeURL(state string) string {
	return c.config.AuthCodeURL(state)
}

func (c *realGoogleOAuthClient) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return c.config.Exchange(ctx, code)
}

func (c *realGoogleOAuthClient) FetchUserInfo(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	return fetchGoogleUserInfo(ctx, c.config, token)
}

func (s *Server) googleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     s.options.GoogleClientID,
		ClientSecret: s.options.GoogleClientSecret,
		RedirectURL:  s.options.BackendURL + "/api/v1/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func (s *Server) oauthClient() googleOAuthClient {
	if s.googleOAuth != nil {
		return s.googleOAuth
	}

	return &realGoogleOAuthClient{config: s.googleOAuthConfig()}
}

// GET /api/v1/auth/google/login

type GoogleLoginOutput struct {
	Status    int           `json:"-"`
	Location  string        `header:"Location"`
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

func (s *Server) handleGoogleLogin(ctx context.Context, input *struct{}) (*GoogleLoginOutput, error) {
	state, err := generateOAuthState()
	if err != nil {
		s.logger.Error("Failed to generate OAuth state", "error", err)
		return nil, huma.Error500InternalServerError("internal error")
	}

	client := s.oauthClient()
	location := client.AuthCodeURL(state)
	stateCookie := newOAuthStateCookie(state)

	return &GoogleLoginOutput{
		Status:   http.StatusFound,
		Location: location,
		SetCookie: []http.Cookie{
			stateCookie,
		},
	}, nil
}

// GET /api/v1/auth/google/callback

type GoogleCallbackInput struct {
	State      string `query:"state"`
	Code       string `query:"code"`
	OAuthState string `cookie:"oauth_state"`
}

type GoogleCallbackOutput struct {
	Status    int           `json:"-"`
	Location  string        `header:"Location"`
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

func (s *Server) handleGoogleCallback(ctx context.Context, input *GoogleCallbackInput) (*GoogleCallbackOutput, error) {
	if input.OAuthState == "" || input.OAuthState != input.State {
		s.logger.Error("Invalid OAuth state")
		return nil, huma.Error400BadRequest("invalid state")
	}

	client := s.oauthClient()

	token, err := client.ExchangeCode(ctx, input.Code)
	if err != nil {
		s.logger.Error("OAuth token exchange failed", "error", err)
		return nil, huma.Error400BadRequest("token exchange failed")
	}

	userInfo, err := client.FetchUserInfo(ctx, token)
	if err != nil {
		s.logger.Error("Failed to fetch Google user info", "error", err)
		return nil, huma.Error500InternalServerError("failed to fetch user info")
	}

	user, err := s.repos.User.Upsert(ctx, userInfo.Sub, userInfo.Email, userInfo.GivenName, userInfo.FamilyName)
	if err != nil {
		s.logger.Error("Failed to upsert user", "error", err)
		return nil, huma.Error500InternalServerError("database error")
	}

	sessionClaims := JWTClaims{
		Sub:       user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	jwtToken, err := SignJWT(s.options.JWTSecret, sessionClaims)
	if err != nil {
		s.logger.Error("Failed to sign JWT", "error", err)
		return nil, huma.Error500InternalServerError("token signing failed")
	}

	secure := isSecureURL(s.options.BackendURL)
	sessionCookie := newSessionCookie(jwtToken, secure)
	clearStateCookie := expiredOAuthStateCookie()

	return &GoogleCallbackOutput{
		Status:   http.StatusFound,
		Location: s.options.FrontendURL,
		SetCookie: []http.Cookie{
			sessionCookie,
			clearStateCookie,
		},
	}, nil
}

type googleUserInfo struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

func fetchGoogleUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (*googleUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func newOAuthStateCookie(state string) http.Cookie {
	return http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func expiredOAuthStateCookie() http.Cookie {
	return http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func newSessionCookie(token string, secure bool) http.Cookie {
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}

	return http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	}
}

// POST /api/v1/auth/logout

type LogoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func (s *Server) handleLogout(ctx context.Context, input *struct{}) (*LogoutOutput, error) {
	return &LogoutOutput{
		SetCookie: logoutCookie(),
	}, nil
}

func logoutCookie() http.Cookie {
	return http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// GET /api/v1/me

type MeOutput struct {
	Body meResponseBody
}

type meResponseBody struct {
	ID        string `json:"id" doc:"User ID"`
	Email     string `json:"email" doc:"Email address"`
	FirstName string `json:"first_name" doc:"First name"`
	LastName  string `json:"last_name" doc:"Last name"`
}

func (s *Server) handleMe(ctx context.Context, input *struct{}) (*MeOutput, error) {
	claims, err := claimsFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	return &MeOutput{
		Body: meResponseBody{
			ID:        claims.Sub,
			Email:     claims.Email,
			FirstName: claims.FirstName,
			LastName:  claims.LastName,
		},
	}, nil
}
