package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

/*

Local development: place a .env file next to the binary (or in the working
directory). cmd/shared.LoadConfig calls godotenv.Load before parsing, so
the .env file takes effect automatically.

Production / Kubernetes: inject env vars directly (ConfigMap, Secret, etc.).
No YAML files are needed.

*/

type Config struct {
	Server   Server
	Logger   Logger
	Database Database
	Security Security
}

type Server struct {
	Host   string `env:"SERVER_HOST"   envDefault:"0.0.0.0"`
	Port   string `env:"SERVER_PORT"   envDefault:"8080"`
	AppEnv string `env:"APP_ENV"       envDefault:"local"`
}

type Logger struct {
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

type Database struct {
	Host           string `env:"DATABASE_HOST"            envDefault:"localhost"`
	Port           string `env:"DATABASE_PORT"            envDefault:"5432"`
	User           string `env:"DATABASE_USER"            envDefault:"postgres"`
	Password       string `env:"DATABASE_PASSWORD,required"`
	Name           string `env:"DATABASE_NAME"            envDefault:"postgres"`
	SSLMode        string `env:"DATABASE_SSL_MODE"        envDefault:"disable"`
	Schema         string `env:"DATABASE_SCHEMA"`
	AutoMigrate    bool   `env:"DATABASE_AUTO_MIGRATE"    envDefault:"false"`
	MigrationsPath string `env:"DATABASE_MIGRATIONS_PATH" envDefault:"migrations"`
}

type Security struct {
	APIKey string `env:"SECURITY_API_KEY"`

	// AllowedOrigins is a comma-separated list parsed into a slice automatically.
	// Example: SECURITY_ALLOW_ORIGINS=http://localhost:5173,http://localhost:3000
	AllowedOrigins []string `env:"SECURITY_ALLOW_ORIGINS" envSeparator:","`

	JWTSecret string        `env:"SECURITY_JWT_SECRET,required"`
	JWTExpiry time.Duration `env:"SECURITY_JWT_EXPIRY" envDefault:"24h"`

	// Content-Security-Policy directives. Leave empty to omit the header.
	CSPDefaultSrc string `env:"SECURITY_CSP_DEFAULT_SRC"`
	CSPScriptSrc  string `env:"SECURITY_CSP_SCRIPT_SRC"`
	CSPStyleSrc   string `env:"SECURITY_CSP_STYLE_SRC"`
	CSPImgSrc     string `env:"SECURITY_CSP_IMG_SRC"`
	CSPFontSrc    string `env:"SECURITY_CSP_FONT_SRC"`
	CSPConnectSrc string `env:"SECURITY_CSP_CONNECT_SRC"`
	CSPFrameSrc   string `env:"SECURITY_CSP_FRAME_SRC"`
	CSPMediaSrc   string `env:"SECURITY_CSP_MEDIA_SRC"`
	CSPObjectSrc  string `env:"SECURITY_CSP_OBJECT_SRC"`
}

// Load parses all Config fields from the current environment variables.
// For local development, call godotenv.Load(".env") before this function so
// that the .env file populates the process environment first.
func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return cfg, nil
}
