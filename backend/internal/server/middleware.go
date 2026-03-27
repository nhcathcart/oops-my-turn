package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func respondUnauthorized(ctx huma.Context) {
	ctx.SetStatus(http.StatusUnauthorized)
	ctx.SetHeader("Content-Type", "application/json")
	b, _ := json.Marshal(map[string]string{"error": "Unauthorized"})
	_, _ = ctx.BodyWriter().Write(b)
}

type jwtClaimsKey struct{}

func claimsFromCtx(ctx context.Context) (*JWTClaims, error) {
	claims, ok := ctx.Value(jwtClaimsKey{}).(*JWTClaims)
	if !ok || claims == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return claims, nil
}

// unauthenticatedPaths are routes that do not require a valid session.
var unauthenticatedPaths = map[string]bool{
	"/api/v1/hello":                true,
	"/api/v1/auth/logout":          true,
	"/api/v1/auth/google/login":    true,
	"/api/v1/auth/google/callback": true,
}

func (s *Server) authenticateJWT(ctx huma.Context, next func(huma.Context)) {
	path := strings.TrimSuffix(ctx.URL().Path, "/")
	if unauthenticatedPaths[path] {
		next(ctx)
		return
	}

	cookie, err := huma.ReadCookie(ctx, sessionCookieName)
	if err != nil {
		respondUnauthorized(ctx)
		return
	}

	claims, err := ParseJWT(s.options.JWTSecret, cookie.Value)
	if err != nil {
		respondUnauthorized(ctx)
		return
	}

	next(huma.WithValue(ctx, jwtClaimsKey{}, claims))
}
