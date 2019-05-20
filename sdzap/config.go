package sdzap

import "go.uber.org/zap/zapcore"

// NewEncoderConfig returns https://cloud.google.com/logging/docs/agent/configuration#special-fields compatible settings.
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "severity",
		NameKey:       "logName",
		MessageKey:    "message",
		StacktraceKey: "", // cared in sdzap.Core. https://cloud.google.com/logging/docs/agent/configuration#special-fields

		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    LevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: DurationEncoder,

		CallerKey:    "",  // cared in sdzap.Core
		EncodeCaller: nil, // cared in sdzap.Core. We need ObjectEncoder.

		EncodeName: nil,
	}
}
