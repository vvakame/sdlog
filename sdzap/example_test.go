package sdzap_test

import (
	"context"

	"github.com/vvakame/sdlog/sdzap"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Example() {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    sdzap.NewEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := cfg.Build(
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return &sdzap.Core{
				Core: core,
			}
		}),
		zap.AddStacktrace(zapcore.WarnLevel),
	)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx := context.Background()
	ctx = sdzap.WithLogger(ctx, logger)

	var span *trace.Span
	{
		ctx, span = trace.StartSpan(ctx, "test-span1")
		defer span.End()
		logger := sdzap.LoggerFromContext(ctx)
		logger.Info("test1", zap.String("foo", "bar"))
	}

	{
		ctx, span = trace.StartSpan(ctx, "test-span2")
		defer span.End()
		logger := sdzap.LoggerFromContext(ctx)
		logger.Warn("test2", zap.String("foo", "bar"))
	}

	// Output:

	// timestamp is varys for each execution.
}
