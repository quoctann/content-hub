package shared

import (
	"fmt"

	"github.com/quoctann/content-hub/pkg/logger"
)

// InitLogger initializes the application logger.
func InitLogger(env string) (logger.ILogger, error) {
	l, err := logger.NewZapLogger(env)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return l, nil
}
