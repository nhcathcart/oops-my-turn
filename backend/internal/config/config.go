package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	Debug     bool   `env:"DEBUG" envDefault:"false"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"text"` // "text" or "json"
	Host      string `env:"HOST" envDefault:"0.0.0.0"`
	Port      int    `env:"PORT" envDefault:"9000"`

	DbHost      string `env:"DB_HOST" envDefault:"localhost"`
	DbPort      int    `env:"DB_PORT" envDefault:"5432"`
	DbUser      string `env:"DB_USER" envDefault:"oops_my_turn"`
	DbPassword  string `env:"DB_PASSWORD" envDefault:"12345"`
	DbName      string `env:"DB_NAME" envDefault:"oops_my_turn"`
	DbSSLMode   string `env:"DB_SSLMODE" envDefault:"disable"`
	DBSecretARN string `env:"DB_SECRET_ARN"`

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	JWTSecret          string `env:"JWT_SECRET"`
	FrontendURL        string `env:"FRONTEND_URL" envDefault:"http://localhost:5173"`
	BackendURL         string `env:"BACKEND_URL" envDefault:"http://localhost:9000"`

	AppSecretsARN string `env:"APP_SECRETS_ARN"`
	AWSRegion     string `env:"AWS_REGION" envDefault:"us-east-1"`
}

type dbSecretPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type appSecretsPayload struct {
	GoogleClientID     string `json:"google_client_id"`
	GoogleClientSecret string `json:"google_client_secret"`
	JWTSecret          string `json:"jwt_secret"`
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

func ParseConfigFromEnv(environment map[string]string) (*Config, error) {
	var cfg Config
	if err := env.ParseWithOptions(&cfg, env.Options{Environment: environment}); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) ResolveSecrets(ctx context.Context) error {
	if c.DBSecretARN == "" && c.AppSecretsARN == "" {
		return nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(c.AWSRegion))
	if err != nil {
		return fmt.Errorf("loading AWS config for secrets: %w", err)
	}
	client := secretsmanager.NewFromConfig(awsCfg)

	if c.DBSecretARN != "" {
		var dbSecret dbSecretPayload
		if err := loadSecretJSON(ctx, client, c.DBSecretARN, &dbSecret); err != nil {
			return fmt.Errorf("loading DB secret: %w", err)
		}
		if dbSecret.Username != "" {
			c.DbUser = dbSecret.Username
		}
		if dbSecret.Password != "" {
			c.DbPassword = dbSecret.Password
		}
	}

	if c.AppSecretsARN != "" {
		var appSecret appSecretsPayload
		if err := loadSecretJSON(ctx, client, c.AppSecretsARN, &appSecret); err != nil {
			return fmt.Errorf("loading app secret: %w", err)
		}
		if appSecret.GoogleClientID != "" {
			c.GoogleClientID = appSecret.GoogleClientID
		}
		if appSecret.GoogleClientSecret != "" {
			c.GoogleClientSecret = appSecret.GoogleClientSecret
		}
		if appSecret.JWTSecret != "" {
			c.JWTSecret = appSecret.JWTSecret
		}
	}

	return nil
}

func loadSecretJSON(ctx context.Context, client *secretsmanager.Client, secretARN string, dest any) error {
	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return err
	}
	if out.SecretString == nil {
		return fmt.Errorf("secret %s does not contain a secret string", secretARN)
	}
	if err := json.Unmarshal([]byte(*out.SecretString), dest); err != nil {
		return fmt.Errorf("unmarshal secret %s: %w", secretARN, err)
	}
	return nil
}

// ValidateAuth checks that required auth environment variables are set.
func (c *Config) ValidateAuth() error {
	for _, required := range []struct{ name, value string }{
		{"JWT_SECRET", c.JWTSecret},
		{"GOOGLE_CLIENT_ID", c.GoogleClientID},
		{"GOOGLE_CLIENT_SECRET", c.GoogleClientSecret},
	} {
		if required.value == "" {
			return fmt.Errorf("%s is required", required.name)
		}
	}
	return nil
}

func EnvToMap(environment []string) map[string]string {
	r := map[string]string{}
	for _, e := range environment {
		p := strings.SplitN(e, "=", 2)
		if len(p) == 2 {
			r[p[0]] = p[1]
		}
	}
	return r
}
