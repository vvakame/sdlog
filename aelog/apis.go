package aelog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vvakame/sdlog/buildlog"
)

// LogWriter use write log entry to somewhere. default is stdout.
var LogWriter = func(ctx context.Context, logEntry *buildlog.LogEntry) {
	b, err := json.Marshal(logEntry)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

type contextHTTPRequestKey struct{}

// WithHTTPRequest is required when you don't use OpenCensus.
func WithHTTPRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, contextHTTPRequestKey{}, r)
}

type contextSourceLocationSkip struct{}

// WithSourceLocationSkip is add skip source location number to context.
func WithSourceLocationSkip(ctx context.Context, skip int) context.Context {
	return context.WithValue(ctx, contextSourceLocationSkip{}, skip)
}

// Criticalf is like Debugf, but at Critical level.
func Criticalf(ctx context.Context, format string, args ...interface{}) {
	emitLog(ctx, buildlog.SeverityCritical, format, args...)
}

// Debugf formats its arguments according to the format, analogous to fmt.Printf, and records the text as a log message at Debug level.
// The message will be associated with the request linked with the provided context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	emitLog(ctx, buildlog.SeverityDebug, format, args...)
}

// Errorf is like Debugf, but at Error level.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	emitLog(ctx, buildlog.SeverityError, format, args...)
}

// Infof is like Debugf, but at Info level.
func Infof(ctx context.Context, format string, args ...interface{}) {
	emitLog(ctx, buildlog.SeverityInfo, format, args...)
}

// Warningf is like Debugf, but at Warning level.
func Warningf(ctx context.Context, format string, args ...interface{}) {
	emitLog(ctx, buildlog.SeverityWarning, format, args...)
}

func emitLog(ctx context.Context, severity buildlog.Severity, format string, args ...interface{}) {
	ctx, err := buildlog.WithConfigurator(ctx, &AppEngineConfigurator{})
	if err != nil {
		panic(err)
	}

	skip, ok := ctx.Value(contextSourceLocationSkip{}).(int)
	if !ok {
		skip = 5
	}
	logEntry := buildlog.NewLogEntry(ctx, buildlog.WithSourceLocationSkip(skip))
	logEntry.Severity = severity
	logEntry.Message = fmt.Sprintf(format, args...)

	LogWriter(ctx, logEntry)
}
