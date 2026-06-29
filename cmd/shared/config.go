package shared

import (
	"fmt"
	"os"

	"github.com/quoctann/content-hub/pkg/config"
)

// LoadConfig loads the application configuration based on the environment.
// It prioritizes the env argument, fallback to APP_ENV environment variable,
// then defaults to "local".
func LoadConfig(env string) (*config.Config, error) {
	if env == "" {
		env = os.Getenv("APP_ENV")
	}

	cfg, err := config.LoadConfigWithEnv(env)
	if err != nil {
		return nil, fmt.Errorf("failed to load config for env %s: %w", env, err)
	}
	return cfg, nil
}
