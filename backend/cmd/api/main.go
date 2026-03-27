package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nhcathcart/oops-my-turn/backend/internal/config"
	"github.com/nhcathcart/oops-my-turn/backend/internal/server"
)

var version = "dev"

func run(ctx context.Context, env map[string]string) error {
	printSpec := flag.Bool("print-spec", false, "Print the OpenAPI specification and exit")
	flag.Parse()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	cfg, err := config.ParseConfigFromEnv(env)
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	if err := cfg.ResolveSecrets(ctx); err != nil {
		return fmt.Errorf("resolving secrets: %w", err)
	}

	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}

	var handler slog.Handler
	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	}
	logger := slog.New(handler)

	dbConfig, err := pgx.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		return fmt.Errorf("parsing database config: %w", err)
	}
	db := stdlib.OpenDB(*dbConfig)
	defer func() { _ = db.Close() }()

	repos := server.NewRepositories(db)
	svcs := server.NewServices(repos)

	options := server.Options{
		Version: version,

		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		JWTSecret:          cfg.JWTSecret,
		FrontendURL:        cfg.FrontendURL,
		BackendURL:         cfg.BackendURL,
	}

	if *printSpec {
		s := server.NewServer(options, db, logger, repos, svcs)
		_, api, err := s.CreateRoutes()
		if err != nil {
			return fmt.Errorf("creating routes: %w", err)
		}
		if err := server.PrintOpenAPISpec(api); err != nil {
			return fmt.Errorf("printing OpenAPI spec: %w", err)
		}
		return nil
	}

	s := server.NewServer(options, db, logger, repos, svcs)
	httpHandler, _, err := s.CreateRoutes()
	if err != nil {
		return fmt.Errorf("creating routes: %w", err)
	}

	if err := cfg.ValidateAuth(); err != nil {
		return fmt.Errorf("invalid auth config: %w", err)
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           httpHandler,
		ReadHeaderTimeout: 30 * time.Second,
	}

	serverErrChan := make(chan error, 1)
	go func() {
		logger.Info("Starting HTTP server", "addr", srv.Addr)
		serverErrChan <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down HTTP server")
	case err := <-serverErrChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server failed: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("HTTP server shutdown complete")
	return nil
}

func main() {
	if err := run(context.Background(), config.EnvToMap(os.Environ())); err != nil {
		log.Fatal(err)
	}
}
