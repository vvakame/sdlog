package gcpslog_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/vvakame/sdlog/v2/gcpslog"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ho := &gcpslog.HandlerOptions{
		Level:     slog.LevelDebug,
		ProjectID: "sdlog-test-project",
		TraceInfo: func(ctx context.Context) (string, string) {
			return "trace-id-a", "span-id-b"
		},
	}

	var buf bytes.Buffer
	h := ho.NewHandler(&buf)

	logger := slog.New(h)
	logger.Enabled(ctx, slog.LevelDebug)
	logger.InfoContext(ctx, "info message")
	logger.ErrorContext(ctx, "error message", "error", errors.New("error"))
	logger.WarnContext(ctx, "warn message", slog.String("time", "2025-06-17T22:00:00Z"))
	logger.LogAttrs(ctx, slog.LevelDebug, "log attrs", slog.String("key", "value"))

	t.Log(buf.String())
}

func Test_example(t *testing.T) {
	defaultLogger := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(defaultLogger)
	})

	var buf bytes.Buffer
	slog.SetDefault(slog.New(gcpslog.HandlerOptions{}.NewHandler(&buf)))

	ctx := context.Background()

	slog.InfoContext(ctx, "info message")
	slog.ErrorContext(ctx, "error message", "error", errors.New("error"))
	slog.LogAttrs(ctx, slog.LevelDebug, "log attrs", slog.String("key", "value"))

	t.Log(buf.String())
}

func TestHandler_ReplaceAttr(t *testing.T) {
	t.Parallel()

	type account struct {
		Name     string
		Password string
	}

	var gotKeys []string
	ho := &gcpslog.HandlerOptions{
		ProjectID: "sdlog-test-project",
		TraceInfo: func(ctx context.Context) (string, string) {
			return "", ""
		},
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 {
				gotKeys = append(gotKeys, a.Key)
			}
			switch {
			case a.Key == "drop":
				return slog.Attr{}
			case a.Value.Kind() == slog.KindAny:
				if v, ok := a.Value.Any().(account); ok {
					v.Password = "[REDACTED]"
					return slog.Any(a.Key, v)
				}
			}
			return a
		},
	}

	var buf bytes.Buffer
	logger := slog.New(ho.NewHandler(&buf))
	logger.With(slog.Any("with", account{Name: "a", Password: "with-secret"})).
		WithGroup("g").
		Info("message", slog.Any("account", account{Name: "b", Password: "group-secret"}), slog.String("drop", "dropped"))

	out := buf.String()
	for _, leaked := range []string{"with-secret", "group-secret", "dropped"} {
		if strings.Contains(out, leaked) {
			t.Errorf("unexpected %q in output: %s", leaked, out)
		}
	}
	for _, want := range []string{`"severity":"INFO"`, `"message":"message"`, `"time":`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in output: %s", want, out)
		}
	}
	// built-in attributes are passed with the standard slog keys
	for _, key := range []string{slog.TimeKey, slog.LevelKey, slog.MessageKey} {
		if !slices.Contains(gotKeys, key) {
			t.Errorf("ReplaceAttr was not called with %q: %v", key, gotKeys)
		}
	}
}
