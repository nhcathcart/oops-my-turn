package server

import (
	"database/sql"
	"log/slog"

	"github.com/nhcathcart/oops-my-turn/backend/internal/repositories"
	"github.com/stephenafamo/bob"
)

type Options struct {
	Version string

	GoogleClientID     string
	GoogleClientSecret string
	JWTSecret          string
	FrontendURL        string
	BackendURL         string
}

type Server struct {
	options     Options
	db          *sql.DB
	bobDB       bob.DB
	logger      *slog.Logger
	repos       repositories.Repositories
	services    Services
	googleOAuth googleOAuthClient
}

func NewServer(options Options, db *sql.DB, logger *slog.Logger, repos repositories.Repositories, svcs Services, opts ...ServerOption) *Server {
	s := &Server{
		options:  options,
		db:       db,
		bobDB:    bob.NewDB(db),
		logger:   logger,
		repos:    repos,
		services: svcs,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type ServerOption func(*Server)

func WithGoogleOAuthClient(client googleOAuthClient) ServerOption {
	return func(s *Server) { s.googleOAuth = client }
}

func (s *Server) Version() string {
	return s.options.Version
}

type Services struct{}

func NewRepositories(db *sql.DB) repositories.Repositories {
	return repositories.NewRepositories(db)
}

func NewServices(_ repositories.Repositories) Services {
	return Services{}
}
