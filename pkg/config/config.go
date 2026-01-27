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
	AutoMigrate    bool   `mapstructure:"auto_migrate"`
	MigrationsPath string `mapstructure:"migrations_path"`
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

func LoadConfigWithEnv(env string) (*Config, error) {
	v := viper.New()

	// 1. Set default config type to yaml
	v.SetConfigType("yaml")

	// 2. Setup environment variables override
	// Example: SERVER_PORT, DATABASE_PASSWORD (instead of DATABASE.PASSWORD)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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

	return &cfg, nil
}
