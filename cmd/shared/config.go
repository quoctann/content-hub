package shared

import (
	"fmt"

	"github.com/quoctann/content-hub/pkg/config"
)

func LoadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}
