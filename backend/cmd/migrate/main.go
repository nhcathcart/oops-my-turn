package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nicholascathcart/nhc-starter/backend/schemata"
	migrate "github.com/rubenv/sql-migrate"
)

type Config struct {
	Enabled     bool   `env:"DB_MIGRATIONS_ENABLED" envDefault:"false"`
	DbHost      string `env:"DB_HOST" envDefault:"localhost"`
	DbPort      int    `env:"DB_PORT" envDefault:"5432"`
	DbUser      string `env:"DB_USER" envDefault:"starter"`
	DbPassword  string `env:"DB_PASSWORD" envDefault:"12345"`
	DbSecretARN string `env:"DB_SECRET_ARN"`
	DbName      string `env:"DB_NAME" envDefault:"starter"`
	DbSSLMode   string `env:"DB_SSLMODE" envDefault:"disable"`
	LogFormat   string `env:"LOG_FORMAT" envDefault:"text"`
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DbUser,
		url.QueryEscape(c.DbPassword),
		c.DbHost,
		c.DbPort,
		c.DbName,
		c.DbSSLMode,
	)
}

func main() {
	down := flag.Bool("down", false, "Roll back the last migration")
	flag.Parse()

	cfg, err := parseConfigFromEnv(envToMap(os.Environ()))
	if err != nil {
		log.Fatalf("could not parse environment variables: %v", err)
	}

	if err := cfg.resolveDBSecret(context.Background()); err != nil {
		log.Fatalf("could not resolve database secret: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("Starting migrate")

	dbConfig, err := pgx.ParseConfig(cfg.DatabaseURL())
	if err != nil {
		logger.Error("could not parse database config url", "error", err)
		os.Exit(1)
	}
	db := stdlib.OpenDB(*dbConfig)
	defer func() { _ = db.Close() }()

	direction := migrate.Up
	if *down {
		direction = migrate.Down
	}

	if err := runMigrations(cfg.Enabled, db, logger, direction); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func (c *Config) resolveDBSecret(ctx context.Context) error {
	if c.DbSecretARN == "" {
		return nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("loading aws config: %w", err)
	}

	client := secretsmanager.NewFromConfig(awsCfg)
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: &c.DbSecretARN})
	if err != nil {
		return fmt.Errorf("getting secret %q: %w", c.DbSecretARN, err)
	}
	if out.SecretString == nil {
		return fmt.Errorf("secret %q has no secret_string", c.DbSecretARN)
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(*out.SecretString), &payload); err != nil {
		return fmt.Errorf("decoding secret %q: %w", c.DbSecretARN, err)
	}
	if payload.Username != "" {
		c.DbUser = payload.Username
	}
	if payload.Password != "" {
		c.DbPassword = payload.Password
	}

	return nil
}

func runMigrations(enabled bool, db *sql.DB, logger *slog.Logger, direction migrate.MigrationDirection) error {
	const migrateTable = "_migration_status"

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: schemata.MigrationsFolder,
		Root:       schemata.RootFolder,
	}

	migrate.SetTable(migrateTable)

	if !enabled {
		logger.Info("Migrations disabled, skipping")
		return nil
	}

	if direction == migrate.Down {
		n, err := migrate.ExecMax(db, "postgres", migrations, migrate.Down, 1)
		if err != nil {
			return fmt.Errorf("rolling back migration: %w", err)
		}
		logger.Info("Migration rollback completed", "rolled_back", n, "migration_table", migrateTable)
		return nil
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	logger.Info("Migration run completed", "applied", n, "migration_table", migrateTable)
	return nil
}

func parseConfigFromEnv(environment map[string]string) (*Config, error) {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{Environment: environment}); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func envToMap(environment []string) map[string]string {
	r := map[string]string{}
	for _, e := range environment {
		p := strings.SplitN(e, "=", 2)
		if len(p) == 2 {
			r[p[0]] = p[1]
		}
	}
	return r
}
