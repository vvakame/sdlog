package sdzap

import (
	"context"
	"time"

	"github.com/vvakame/sdlog/aelog"

	"github.com/vvakame/sdlog/buildlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextLoggerKey struct{}

// WithLogger bind logger to specified context.
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	_, ok := ctx.Value(contextLoggerKey{}).(*zap.Logger)
	if ok {
		panic("don't pairing twice")
	}
	return context.WithValue(ctx, contextLoggerKey{}, logger)
}

// LoggerFromContext extracts *zap.Logger from context.
// It bind trace and spanId to logger.
func LoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(contextLoggerKey{}).(*zap.Logger)
	if !ok {
		panic("ctx doesn't have zap.Logger. try with WithLogger")
	}
	logger = logger.With(LogEntryTraceAndSpanID(ctx)...)
	return logger
}

// LevelEncoder encode value to Stackdriver format.
func LevelEncoder(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
	switch level {
	case zapcore.DebugLevel:
		encoder.AppendString("DEBUG")
	case zapcore.InfoLevel:
		encoder.AppendString("INFO")
	case zapcore.WarnLevel:
		encoder.AppendString("WARNING")
	case zapcore.ErrorLevel:
		encoder.AppendString("ERROR")
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		encoder.AppendString("CRITICAL")
	case zapcore.FatalLevel:
		encoder.AppendString("ALERT")
	}
}

// TimeEncoder encode value to Stackdriver format.
func TimeEncoder(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(t.Format(time.RFC3339Nano))
}

// DurationEncoder encode value to Stackdriver format.
func DurationEncoder(d time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
	encoder.AppendString(d.String())
}

// LogEntryTraceAndSpanID returns trace and spanId fields from context.
func LogEntryTraceAndSpanID(ctx context.Context) []zapcore.Field {
	cfg := buildlog.ConfiguratorFromContext(ctx)
	if cfg == nil {
		cfg = aelog.DefaultConfigurator
	}
	traceID, spanID := cfg.TraceInfo(ctx)

	return []zapcore.Field{
		zap.String("logging.googleapis.com/trace", traceID),
		zap.String("logging.googleapis.com/spanId", spanID),
	}
}

// LogEntrySourceLocation returns sourceLocation field.
func LogEntrySourceLocation(loc *buildlog.LogEntrySourceLocation) zapcore.Field {
	return zap.Object("logging.googleapis.com/sourceLocation", loc)
}
