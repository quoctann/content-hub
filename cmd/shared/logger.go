package shared

import (
	"fmt"

	"github.com/quoctann/content-hub/pkg/logger"
)

func InitLogger(appEnv, level string) (logger.ILogger, error) {
	l, err := logger.NewZapLogger(appEnv, level)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	return l, nil
}
