package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   Server   `mapstructure:"server"`
	Logger   Logger   `mapstructure:"logger"`
	Database Database `mapstructure:"database"`
	Security Security `mapstructure:"security"`
}

type Server struct {
	Host   string `mapstructure:"host"`
	Port   string `mapstructure:"port"`
	AppEnv string `mapstructure:"app_env"`
}

type Logger struct {
	Level string `mapstructure:"level"`
}

type Database struct {
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Name           string `mapstructure:"name"`
	SSLMode        string `mapstructure:"ssl_mode"`
	Schema         string `mapstructure:"schema"`
	AutoMigrate    bool   `mapstructure:"auto_migrate"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

type Security struct {
	APIKey         string `mapstructure:"api_key"`
	AllowedOrigins string `mapstructure:"allow_origins"`
	JWTSecret      string `mapstructure:"jwt_secret"`
	JWTExpiry      string `mapstructure:"jwt_expiry"`
	CSPDefaultSrc  string `mapstructure:"csp_default_src"`
	CSPScriptSrc   string `mapstructure:"csp_script_src"`
	CSPStyleSrc    string `mapstructure:"csp_style_src"`
	CSPImgSrc      string `mapstructure:"csp_img_src"`
	CSPFontSrc     string `mapstructure:"csp_font_src"`
	CSPConnectSrc  string `mapstructure:"csp_connect_src"`
	CSPFrameSrc    string `mapstructure:"csp_frame_src"`
	CSPMediaSrc    string `mapstructure:"csp_media_src"`
	CSPObjectSrc   string `mapstructure:"csp_object_src"`
}

// isFileNotFoundError checks if the error indicates a file not found.
func isFileNotFoundError(err error) bool {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return os.IsNotExist(pathErr)
	}
	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}

// GetAllowedOrigins returns a slice of allowed origins from comma-separated string
func (s *Security) GetAllowedOrigins() []string {
	if s.AllowedOrigins == "" {
		return []string{}
	}
	var origins []string
	for _, origin := range strings.Split(s.AllowedOrigins, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func LoadConfigWithEnv(env string) (*Config, error) {
	v := viper.New()

	// 1. Set default config type to yaml
	v.SetConfigType("yaml")

	// 2. Setup environment variables override
	// Example: SERVER_PORT, DATABASE_PASSWORD (instead of DATABASE.PASSWORD)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))


	if err := bindEnv(v); err != nil {
		return nil, fmt.Errorf("failed to bind env vars: %w", err)
	}

	// 3. Load configurations in order (cascading)

	// First: General config file (base defaults)
	v.SetConfigFile("config.yaml")
	if err := v.MergeInConfig(); err != nil {
		if !isFileNotFoundError(err) {
			return nil, fmt.Errorf("error reading general config file: %w", err)
		}
		// If config.yaml is missing, we might still proceed if env-specific file exists
	}

	// Second: Environment-specific config (e.g., config-dev.yaml)
	if env != "" {
		envConfigFile := fmt.Sprintf("config-%s.yaml", env)
		v.SetConfigFile(envConfigFile)
		if err := v.MergeInConfig(); err != nil {
			if !isFileNotFoundError(err) {
				return nil, fmt.Errorf("error reading %s config file: %w", envConfigFile, err)
			}
			// If env-specific file is missing, it's okay as long as base config or env vars exist
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate performs security checks on the loaded configuration.
func (c *Config) Validate() error {
	if c.Security.JWTSecret == "" {
		return errors.New("security: SECURITY_JWT_SECRET is not configured or uses the default placeholder; refusing to start")
	}

	return nil
}

func bindEnv(v *viper.Viper) error {
		envs := map[string]string{
			"server.port":              "SERVER_PORT",
			"app_env":                  "APP_ENV",
			"logger.level":             "LOGGER_LEVEL",
			"database.host":            "DATABASE_HOST",
			"database.port":            "DATABASE_PORT",
			"database.user":            "DATABASE_USER",
			"database.password":        "DATABASE_PASSWORD",
			"database.name":            "DATABASE_NAME",
			"database.ssl_mode":        "DATABASE_SSL_MODE",
			"database.schema":          "DATABASE_SCHEMA",
			"database.auto_migrate":    "DATABASE_AUTO_MIGRATE",
			"database.migrations_path": "DATABASE_MIGRATIONS_PATH",
			"security.allow_origins":   "SECURITY_ALLOW_ORIGINS",
			"security.jwt_secret":      "SECURITY_JWT_SECRET",
			"security.jwt_expiry":      "SECURITY_JWT_EXPIRY",
			"security.csp_default_src": "SECURITY_CSP_DEFAULT_SRC",
			"security.csp_script_src":  "SECURITY_CSP_SCRIPT_SRC",
			"security.csp_style_src":   "SECURITY_CSP_STYLE_SRC",
			"security.csp_img_src":     "SECURITY_CSP_IMG_SRC",
			"security.csp_font_src":    "SECURITY_CSP_FONT_SRC",
			"security.csp_connect_src": "SECURITY_CSP_CONNECT_SRC",
			"security.csp_frame_src":   "SECURITY_CSP_FRAME_SRC",
			"security.csp_media_src":   "SECURITY_CSP_MEDIA_SRC",
			"security.csp_object_src":  "SECURITY_CSP_OBJECT_SRC",
		}
	
		for key, env := range envs {
			if err := v.BindEnv(key, env); err != nil {
				return err
			}
		}
	
		return nil
	}