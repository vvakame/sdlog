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

func newRequestLogEntry(ctx context.Context, r *http.Request) *LogEntry {
	u := *r.URL
	u.Fragment = ""

	logger, ok := ctx.Value(contextLoggerKey{}).(Configurator)
	if !ok {
		logger = DefaultConfigurator
	}

	traceID, spanID := logger.TraceInfo(ctx)

	if !strings.Contains(traceID, "/") {
		traceID = fmt.Sprintf("projects/%s/traces/%s", logger.ProjectID(), traceID)
	}

	remoteIP := ""
	if v := r.Header.Get("X-AppEngine-User-IP"); v != "" {
		remoteIP = v
	} else if v := r.Header.Get("X-Forwarded-For"); v != "" {
		remoteIP = v
	} else {
		remoteIP = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	endAt := time.Now()

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

	logEntry := &LogEntry{
		Time:        Time(endAt),
		HTTPRequest: httpRequestEntry,
		Trace:       traceID,
		SpanID:      spanID,
	}

	return logEntry
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
