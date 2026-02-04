package database

import (
	"context"

	"github.com/jackc/pgx/v5/tracelog"
	"github.com/quoctann/content-hub/pkg/logger"
)

// LoggerAdapter adapts our logger.ILogger to pgx tracelog.Logger
type LoggerAdapter struct {
	logger logger.ILogger
}

// NewLoggerAdapter creates a new LoggerAdapter
func NewLoggerAdapter(l logger.ILogger) *LoggerAdapter {
	return &LoggerAdapter{logger: l}
}

// Log implements the tracelog.Logger interface
func (l *LoggerAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := make([]logger.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, logger.Any(k, v))
	}

	// Add warning about sensitive data
	fields = append(fields, logger.String("security_warning", "LOG DEBUG SQL CAN EXPOSE SENSITIVE DATA"))

	switch level {
	case tracelog.LogLevelTrace:
		l.logger.Debug(ctx, msg, fields...)
	case tracelog.LogLevelDebug:
		l.logger.Debug(ctx, msg, fields...)
	case tracelog.LogLevelInfo:
		l.logger.Info(ctx, msg, fields...)
	case tracelog.LogLevelWarn:
		l.logger.Warn(ctx, msg, fields...)
	case tracelog.LogLevelError:
		l.logger.Error(ctx, msg, fields...)
	default:
		l.logger.Info(ctx, msg, fields...)
	}
}
