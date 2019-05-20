package buildlog

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type contextLoggerKey struct{}

// DefaultConfigurator will use logging without WithConfigurator.
var DefaultConfigurator Configurator = &appengineConfigurator{}

// Configurator provides some value from environment.
type Configurator interface {
	ProjectID() string
	TraceInfo(ctx context.Context) (traceID, spanID string)
	RemoteIP(r *http.Request) string
}

// WithConfigurator bundles Configurator to context.
func WithConfigurator(ctx context.Context, cfg Configurator) (context.Context, error) {
	return context.WithValue(ctx, contextLoggerKey{}, cfg), nil
}

// ConfiguratorFromContext extract Configurator from ctx.
func ConfiguratorFromContext(ctx context.Context) Configurator {
	cfg, ok := ctx.Value(contextLoggerKey{}).(Configurator)
	if !ok {
		return nil
	}
	return cfg
}

// NewLogEntry returns *LogEntry for current executing line.
func NewLogEntry(ctx context.Context, opts ...LogEntryOption) *LogEntry {
	opts = append([]LogEntryOption{WithSourceLocationSkip(3)}, opts...)
	return newApplicationLog(ctx, opts...)
}

// LogEntryOption provides some options.
type LogEntryOption func(*logEntryConfig)

type logEntryConfig struct {
	skip int
}

// WithSourceLocationSkip provides skip depth for runtime.Caller.
func WithSourceLocationSkip(skip int) LogEntryOption {
	return func(cfg *logEntryConfig) {
		cfg.skip = skip
	}
}

func newApplicationLog(ctx context.Context, opts ...LogEntryOption) *LogEntry {
	logger, ok := ctx.Value(contextLoggerKey{}).(Configurator)
	if !ok {
		logger = DefaultConfigurator
	}

	traceID, spanID := logger.TraceInfo(ctx)

	if !strings.Contains(traceID, "/") {
		traceID = fmt.Sprintf("projects/%s/traces/%s", logger.ProjectID(), traceID)
	}

	cfg := &logEntryConfig{}
	for _, ops := range opts {
		ops(cfg)
	}

	logEntry := &LogEntry{
		Time:           Time(time.Now()),
		Trace:          traceID,
		SpanID:         spanID,
		SourceLocation: newLogEntrySourceLocation(cfg),
	}

	return logEntry
}

func newHTTPRequestLogEntry(ctx context.Context, r *http.Request) *HTTPRequest {
	u := *r.URL
	u.Fragment = ""

	logger, ok := ctx.Value(contextLoggerKey{}).(Configurator)
	if !ok {
		logger = DefaultConfigurator
	}

	remoteIP := logger.RemoteIP(r)

	falseV := false
	httpRequestEntry := &HTTPRequest{
		RequestMethod:                  r.Method,
		RequestURL:                     u.RequestURI(),
		RequestSize:                    r.ContentLength,
		UserAgent:                      r.UserAgent(),
		RemoteIP:                       remoteIP,
		Referer:                        r.Referer(),
		CacheLookup:                    &falseV,
		CacheHit:                       &falseV,
		CacheValidatedWithOriginServer: &falseV,
		CacheFillBytes:                 nil,
		Protocol:                       r.Proto,
	}

	return httpRequestEntry
}

func newLogEntrySourceLocation(cfg *logEntryConfig) *LogEntrySourceLocation {
	skip := cfg.skip
	if skip == 0 {
		skip = 2
	}

	var sl *LogEntrySourceLocation
	if pc, file, line, ok := runtime.Caller(skip); ok {
		sl = &LogEntrySourceLocation{
			File: file,
			Line: int64(line),
		}
		if function := runtime.FuncForPC(pc); function != nil {
			sl.Function = function.Name()
		}
	}

	return sl
}
