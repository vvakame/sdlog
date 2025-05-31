package gcpslog

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/otel/trace"
)

// spec. https://cloud.google.com/logging/docs/agent/logging/configuration#special-fields

// HandlerOptions are options for a Cloud Logging compatible handler.
type HandlerOptions struct {
	Level     slog.Leveler
	ProjectID string
	TraceInfo func(ctx context.Context) (traceID string, spanID string)
}

// NewHandler creates a Cloud Logging compatible handler with the given options that writes to w.
func (ho HandlerOptions) NewHandler(w io.Writer) slog.Handler {
	if ho.ProjectID == "" {
		ho.ProjectID = gcpProjectID()
	}
	if ho.TraceInfo == nil {
		ho.TraceInfo = openCensusTraceInfo
	}

	h := &handler{
		base: slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   false,
			Level:       ho.Level,
			ReplaceAttr: replaceAttrs,
		}),
		projectID: ho.ProjectID,
		traceInfo: ho.TraceInfo,
	}

	return h
}

type handler struct {
	base      slog.Handler
	projectID string
	traceInfo func(ctx context.Context) (string, string)
}

func (h *handler) clone() *handler {
	return &handler{
		base:      h.base,
		projectID: h.projectID,
		traceInfo: h.traceInfo,
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	if record.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		f, _ := fs.Next()

		// spec: https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#logentrysourcelocation
		record.AddAttrs(
			slog.Group(
				"logging.googleapis.com/sourceLocation",
				slog.String("file", f.File),
				slog.String("line", strconv.Itoa(f.Line)),
				slog.String("function", f.Function),
			),
		)
	}

	traceID, spanID := h.traceInfo(ctx)
	if traceID != "" && !strings.Contains(traceID, "/") {
		traceID = fmt.Sprintf("projects/%s/traces/%s", h.projectID, traceID)
	}
	if traceID != "" {
		record.AddAttrs(slog.String("logging.googleapis.com/trace", traceID))
	}
	if spanID != "" {
		record.AddAttrs(slog.String("logging.googleapis.com/spanId", spanID))
	}

	return h.base.Handle(ctx, record)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h = h.clone()
	h.base = h.base.WithAttrs(attrs)

	return h
}

func (h *handler) WithGroup(name string) slog.Handler {
	h = h.clone()
	h.base = h.base.WithGroup(name)

	return h
}

func gcpProjectID() string {
	if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	}
	if v, _ := metadata.ProjectID(); v != "" {
		return v
	}

	return ""
}

func openCensusTraceInfo(ctx context.Context) (string, string) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return "", ""
	}

	return span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String()
}

func replaceAttrs(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		a.Key = "time"
		a.Value = slog.StringValue(a.Value.Time().Format(time.RFC3339Nano))
	case slog.LevelKey:
		a.Key = "severity"
		level, ok := a.Value.Any().(slog.Level)
		if !ok {
			level = slog.LevelError
		}
		switch level {
		case slog.LevelDebug:
			a.Value = slog.StringValue("DEBUG")
		case slog.LevelInfo:
			a.Value = slog.StringValue("INFO")
		case slog.LevelWarn:
			a.Value = slog.StringValue("WARNING")
		case slog.LevelError:
			a.Value = slog.StringValue("ERROR")
		default:
			a.Value = slog.StringValue("ERROR")
		}
	case slog.MessageKey:
		a.Key = "message"
	case slog.SourceKey:
		// nothing to do
	default:
		// ok
	}

	return a
}
