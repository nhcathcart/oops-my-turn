package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	debugUtil "runtime/debug"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/rs/cors"
)

func (s *Server) CreateRoutes() (http.Handler, huma.API, error) {
	mux := http.NewServeMux()

	api := humago.New(mux, huma.DefaultConfig("Starter API", s.options.Version))

	api.UseMiddleware(s.recoverPanic, AddCommonHeaders, s.authenticateJWT)

	mux.HandleFunc("/healthz", HandleHealthCheck(s))

	s.registerCoreRoutes(api)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{s.options.FrontendURL},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})

	return c.Handler(mux), api, nil
}

func (s *Server) recoverPanic(ctx huma.Context, next func(huma.Context)) {
	defer func() {
		if err := recover(); err != nil {
			ctx.SetHeader("Connection", "close")
			ctx.SetStatus(http.StatusInternalServerError)

			s.logger.Error("Panic recovered",
				"error", err,
				"stack", string(debugUtil.Stack()),
			)

			errorResponse := map[string]string{
				"error":   "Internal Server Error",
				"message": "An unexpected error occurred while processing your request",
			}

			responseJSON, jsonErr := json.Marshal(errorResponse)
			if jsonErr == nil {
				ctx.SetHeader("Content-Type", "application/json")
				_, _ = ctx.BodyWriter().Write(responseJSON)
			}
		}
	}()
	next(ctx)
}

func AddCommonHeaders(ctx huma.Context, next func(huma.Context)) {
	ctx.SetHeader("Server", "Starter API")
	ctx.SetHeader("X-Content-Type-Options", "nosniff")
	ctx.SetHeader("X-Xss-Protection", "1; mode=block;")
	ctx.SetHeader("Referrer-Policy", "no-referrer")
	ctx.SetHeader("X-Frame-Options", "SAMEORIGIN")
	next(ctx)
}

func (s *Server) registerCoreRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "get-hello",
		Method:      http.MethodGet,
		Path:        "/api/v1/hello",
		Summary:     "Get a public starter response",
	}, s.handleHello)

	// start auth region
	huma.Register(api, huma.Operation{
		OperationID: "google-login",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/google/login",
		Summary:     "Initiate Google OAuth login",
		Hidden:      true,
	}, s.handleGoogleLogin)

	huma.Register(api, huma.Operation{
		OperationID: "google-callback",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/google/callback",
		Summary:     "Handle Google OAuth callback",
		Hidden:      true,
	}, s.handleGoogleCallback)

	huma.Register(api, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Log out the current user",
	}, s.handleLogout)

	huma.Register(api, huma.Operation{
		OperationID: "get-me",
		Method:      http.MethodGet,
		Path:        "/api/v1/me",
		Summary:     "Get the current authenticated user",
	}, s.handleMe)
	// end auth region
}

func PrintOpenAPISpec(api huma.API) error {
	spec, err := api.OpenAPI().DowngradeYAML()
	if err != nil {
		return err
	}
	fmt.Print(string(spec))
	return nil
}
