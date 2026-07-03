package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type IContextLogger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
}

type IWithoutContextLogger interface {
	DebugWithoutCtx(msg string, fields ...Field)
	InfoWithoutCtx(msg string, fields ...Field)
	WarnWithoutCtx(msg string, fields ...Field)
	ErrorWithoutCtx(msg string, fields ...Field)
	FatalWithoutCtx(msg string, fields ...Field)
}

type ILogger interface {
	IContextLogger
	IWithoutContextLogger

	With(fields ...Field) ILogger
	Sync() error
}

// FieldType defines the type of the field value
type FieldType uint8

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeString
	FieldTypeInt64
	FieldTypeBool
	FieldTypeError
	FieldTypeAny
)

// Field represents a log field with a key-value pair.
type Field struct {
	Key       string
	Type      FieldType
	StringVal string
	Int64Val  int64
	Interface interface{}
}

type zapLogger struct {
	zap *zap.Logger
}

func NewZapLogger(appEnv, level string) (ILogger, error) {
	var cfg zap.Config
	if appEnv == "local" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if level != "" {
		var zapLevel zapcore.Level
		if err := zapLevel.UnmarshalText([]byte(level)); err == nil {
			cfg.Level = zap.NewAtomicLevelAt(zapLevel)
		}
	}

	l, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &zapLogger{zap: l}, nil
}

// Convert logger.Field to zap.Field
func (l *zapLogger) toZapFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch f.Type {
		case FieldTypeString:
			zapFields[i] = zap.String(f.Key, f.StringVal)
		case FieldTypeInt64:
			zapFields[i] = zap.Int64(f.Key, f.Int64Val)
		case FieldTypeBool:
			zapFields[i] = zap.Bool(f.Key, f.Int64Val == 1)
		case FieldTypeError:
			zapFields[i] = zap.Error(f.Interface.(error))
		case FieldTypeAny:
			zapFields[i] = zap.Any(f.Key, f.Interface)
		default:
			zapFields[i] = zap.Any(f.Key, f.Interface)
		}
	}
	return zapFields
}

// Context-aware methods - extract trace info from context
func (l *zapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.zap.Debug(msg, append(l.toZapFields(fields...), l.toZapFields(extractTraceInfo(ctx)...)...)...)
}

func (l *zapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.zap.Info(msg, append(l.toZapFields(fields...), l.toZapFields(extractTraceInfo(ctx)...)...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.zap.Warn(msg, append(l.toZapFields(fields...), l.toZapFields(extractTraceInfo(ctx)...)...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.zap.Error(msg, append(l.toZapFields(fields...), l.toZapFields(extractTraceInfo(ctx)...)...)...)
}

func (l *zapLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.zap.Fatal(msg, append(l.toZapFields(fields...), l.toZapFields(extractTraceInfo(ctx)...)...)...)
}

// Non-context methods - use only when context is unavailable
func (l *zapLogger) DebugWithoutCtx(msg string, fields ...Field) {
	l.zap.Debug(msg, l.toZapFields(fields...)...)
}

func (l *zapLogger) InfoWithoutCtx(msg string, fields ...Field) {
	l.zap.Info(msg, l.toZapFields(fields...)...)
}

func (l *zapLogger) WarnWithoutCtx(msg string, fields ...Field) {
	l.zap.Warn(msg, l.toZapFields(fields...)...)
}

func (l *zapLogger) ErrorWithoutCtx(msg string, fields ...Field) {
	l.zap.Error(msg, l.toZapFields(fields...)...)
}

func (l *zapLogger) FatalWithoutCtx(msg string, fields ...Field) {
	l.zap.Fatal(msg, l.toZapFields(fields...)...)
}

func (l *zapLogger) With(fields ...Field) ILogger {
	return &zapLogger{zap: l.zap.With(l.toZapFields(fields...)...)}
}

func (l *zapLogger) Sync() error {
	return l.zap.Sync()
}

// Trace Context Keys
const (
	TraceIDKey   = "trace_id"
	SpanIDKey    = "span_id"
	RequestIDKey = "request_id"
)

// WithTraceContext adds tracing information to the context
func WithTraceContext(ctx context.Context, reqID, traceID, spanID string) context.Context {
	if reqID != "" {
		ctx = context.WithValue(ctx, RequestIDKey, reqID)
	}
	if traceID != "" {
		ctx = context.WithValue(ctx, TraceIDKey, traceID)
	}
	if spanID != "" {
		ctx = context.WithValue(ctx, SpanIDKey, spanID)
	}
	return ctx
}

// extractTraceInfo extracts tracing information from the context
func extractTraceInfo(ctx context.Context) []Field {
	var fields []Field

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		fields = append(fields, String("request_id", reqID))
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
		fields = append(fields, String("span_id", spanID))
	}

	return fields
}

// Utility functions for creating fields
func String(key, val string) Field {
	return Field{Key: key, Type: FieldTypeString, StringVal: val}
}
func Int(key string, val int) Field {
	return Field{Key: key, Type: FieldTypeInt64, Int64Val: int64(val)}
}
func Int64(key string, val int64) Field {
	return Field{Key: key, Type: FieldTypeInt64, Int64Val: val}
}
func Error(err error) Field {
	return Field{Key: "error", Type: FieldTypeError, Interface: err}
}
func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: FieldTypeAny, Interface: val}
}
func Bool(key string, val bool) Field {
	var i int64
	if val {
		i = 1
	}
	return Field{Key: key, Type: FieldTypeBool, Int64Val: i}
}

/*
// For multiple loggers

type MultiLogger struct {
    loggers []ILogger
}

func NewMultiLogger(loggers ...ILogger) ILogger {
    return &MultiLogger{loggers: loggers}
}

func (m *MultiLogger) Info(ctx context.Context, msg string, fields ...Field) {
    for _, logger := range m.loggers {
        logger.Info(ctx, msg, fields...)
    }
}
// ... implement other ILogger methods

*/
