package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
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
	Server        Server
	Logger        Logger
	Database      Database
	Security      Security
	Observability Observability
	Upload        Upload
}

type Server struct {
	Host         string        `env:"SERVER_HOST"   envDefault:"0.0.0.0"`
	Port         string        `env:"SERVER_PORT"   envDefault:"8080"`
	AppEnv       string        `env:"APP_ENV"       envDefault:"local"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"60s"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"150s"`

	// TrustedProxies lists the proxy IPs/CIDRs whose X-Forwarded-For entries are
	// trusted when resolving the client IP (rate limiting). Empty means trust no
	// proxy and use the TCP peer address.
	// Example: SERVER_TRUSTED_PROXIES=10.42.0.0/16
	TrustedProxies []string `env:"SERVER_TRUSTED_PROXIES" envSeparator:","`
}

type Upload struct {
	Enabled            bool          `env:"UPLOAD_ENABLED" envDefault:"false"`
	MaxBatchFiles      int           `env:"UPLOAD_MAX_BATCH_FILES" envDefault:"5"`
	MaxFileBytes       int64         `env:"UPLOAD_MAX_FILE_BYTES" envDefault:"10485760"`
	Timeout            time.Duration `env:"UPLOAD_TIMEOUT" envDefault:"60s"`
	ImageKitPrivateKey string        `env:"IMAGEKIT_PRIVATE_KEY"`
	ImageKitFolder     string        `env:"IMAGEKIT_FOLDER" envDefault:"/content-hub"`
	ImageKitEndpoint   string        `env:"IMAGEKIT_UPLOAD_ENDPOINT" envDefault:"https://upload.imagekit.io/api/v1/files/upload"`
}

type Logger struct {
	Level string `env:"LOGGER_LEVEL" envDefault:"info"`
}

type Observability struct {
	OTelSDKDisabled          bool   `env:"OTEL_SDK_DISABLED" envDefault:"true"`
	OTelExporterOTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
	OTelExporterOTLPInsecure bool   `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"true"`
	OTelServiceName          string `env:"OTEL_SERVICE_NAME" envDefault:"content-hub-backend"`
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

	// notEmpty matters: `required` alone accepts SECURITY_JWT_SECRET="" and would
	// sign tokens with an empty HMAC key.
	JWTSecret string        `env:"SECURITY_JWT_SECRET,required,notEmpty"`
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

// minJWTSecretLength is the minimum HS256 key length (256 bits).
const minJWTSecretLength = 32

// Load parses all Config fields from the current environment variables.
// For local development, call godotenv.Load(".env") before this function so
// that the .env file populates the process environment first.
func Load() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if len(cfg.Security.JWTSecret) < minJWTSecretLength {
		return nil, fmt.Errorf("SECURITY_JWT_SECRET must be at least %d characters", minJWTSecretLength)
	}
	cfg.Server.TrustedProxies = compact(cfg.Server.TrustedProxies)
	for _, proxy := range cfg.Server.TrustedProxies {
		if _, _, err := net.ParseCIDR(proxy); err != nil && net.ParseIP(proxy) == nil {
			return nil, fmt.Errorf("SERVER_TRUSTED_PROXIES: %q is not an IP or CIDR", proxy)
		}
	}
	if cfg.Upload.Enabled {
		if cfg.Upload.MaxBatchFiles < 1 || cfg.Upload.MaxFileBytes < 1 || cfg.Upload.Timeout <= 0 {
			return nil, fmt.Errorf("upload configuration must use positive limits and timeout")
		}
		if cfg.Upload.ImageKitPrivateKey == "" {
			return nil, fmt.Errorf("IMAGEKIT_PRIVATE_KEY is required when uploads are enabled")
		}
		endpoint, err := url.Parse(cfg.Upload.ImageKitEndpoint)
		if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() != "upload.imagekit.io" {
			return nil, fmt.Errorf("IMAGEKIT_UPLOAD_ENDPOINT must be an HTTPS ImageKit upload endpoint")
		}
	}
	return cfg, nil
}

// compact trims entries and drops blanks, so an empty or trailing-comma list
// env var (e.g. SERVER_TRUSTED_PROXIES="") yields no entries.
func compact(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
