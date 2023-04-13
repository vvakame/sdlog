package gcpslog

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"golang.org/x/exp/slog"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ho := &HandlerOptions{
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
	logger.InfoCtx(ctx, "info message")
	logger.ErrorCtx(ctx, "error message", "error", errors.New("error"))
	logger.LogAttrs(ctx, slog.LevelDebug, "log attrs", slog.String("key", "value"))

	t.Log(buf.String())
}
