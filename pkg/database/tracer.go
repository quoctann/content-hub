package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/quoctann/content-hub/pkg/logger"
)

// LoggerAdapter adapts our logger.ILogger to pgx tracelog.Logger
type LoggerAdapter struct {
	logger          logger.ILogger
	redactSensitive bool
}

// NewLoggerAdapter creates a new LoggerAdapter
func NewLoggerAdapter(l logger.ILogger, redactSensitive bool) *LoggerAdapter {
	return &LoggerAdapter{logger: l, redactSensitive: redactSensitive}
}

// Log implements the tracelog.Logger interface
func (l *LoggerAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	fields := traceLogFields(data, l.redactSensitive)
	if !l.redactSensitive {
		fields = append(fields, logger.String("security_warning", "SQL DEBUG LOGS CAN EXPOSE SENSITIVE DATA"))
	}

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

func traceLogFields(data map[string]interface{}, redactSensitive bool) []logger.Field {
	fields := make([]logger.Field, 0, len(data))
	for k, v := range data {
		if redactSensitive {
			switch k {
			case "args", "sql":
				continue
			case "err":
				fields = append(fields, logger.String("db_error", "database operation failed"))
				if err, ok := v.(error); ok {
					var pgErr *pgconn.PgError
					if errors.As(err, &pgErr) {
						fields = append(fields, logger.String("sqlstate", pgErr.Code))
					}
				}
				continue
			}
		}
		fields = append(fields, logger.Any(k, v))
	}
	return fields
}
