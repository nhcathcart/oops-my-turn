//go:build integration

package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nhcathcart/oops-my-turn/backend/internal/repositories"
	"github.com/nhcathcart/oops-my-turn/backend/internal/server"
	"github.com/nhcathcart/oops-my-turn/backend/schemata"
	"github.com/nhcathcart/oops-my-turn/backend/sdk"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testJWTSecret = "test-secret"

type IntegrationTestSuite struct {
	suite.Suite
	db     *sql.DB
	ts     *httptest.Server
	client *sdk.ClientWithResponses
	repos  repositories.Repositories
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("oops_my_turn"),
		tcpostgres.WithUsername("oops_my_turn"),
		tcpostgres.WithPassword("12345"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(s.T(), err)
	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)

	dsn := fmt.Sprintf("postgres://oops_my_turn:12345@%s:%s/oops_my_turn?sslmode=disable", host, mappedPort.Port())
	dbConfig, err := pgx.ParseConfig(dsn)
	require.NoError(s.T(), err)
	s.db = stdlib.OpenDB(*dbConfig)
	s.T().Cleanup(func() { _ = s.db.Close() })

	s.runMigrations()

	repos := server.NewRepositories(s.db)
	s.repos = repos
	svcs := server.NewServices(repos)

	sv := server.NewServer(server.Options{
		Version:   "test",
		JWTSecret: testJWTSecret,
	}, s.db, slog.Default(), repos, svcs)
	handler, _, err := sv.CreateRoutes()
	require.NoError(s.T(), err)

	s.ts = httptest.NewServer(handler)
	s.T().Cleanup(s.ts.Close)

	sessionToken, err := server.SignJWT(testJWTSecret, server.JWTClaims{
		Sub:       "test-user",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	})
	require.NoError(s.T(), err)

	s.client, err = sdk.NewClientWithResponses(s.ts.URL, sdk.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
			return nil
		},
	))
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) SetupTest() {
	_, err := s.db.Exec(`TRUNCATE TABLE users RESTART IDENTITY CASCADE`)
	require.NoError(s.T(), err)
	s.seedTestUser()
}

func (s *IntegrationTestSuite) seedTestUser() {
	_, err := s.db.Exec(
		`INSERT INTO users (id, google_id, email, first_name, last_name)
		 VALUES ('test-user', 'google-test', 'test@example.com', 'Test', 'User')`,
	)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) runMigrations() {
	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: schemata.MigrationsFolder,
		Root:       schemata.RootFolder,
	}
	migrate.SetTable("_migration_status")
	n, err := migrate.Exec(s.db, "postgres", migrations, migrate.Up)
	require.NoError(s.T(), err)
	s.T().Logf("applied %d migration(s)", n)
}
