package sdzap

import (
	"runtime"

	"github.com/vvakame/sdlog/buildlog"
	"go.uber.org/zap/zapcore"
)

var _ zapcore.Core = (*Core)(nil)

// Core for https://cloud.google.com/logging/docs/agent/configuration#special-fields compatible format.
type Core struct {
	Core zapcore.Core
}

// Enabled returns specified level will emit.
func (core *Core) Enabled(level zapcore.Level) bool {
	return core.Core.Enabled(level)
}

// With adds structured context to the Core.
func (core *Core) With(fields []zapcore.Field) zapcore.Core {
	return &Core{
		Core: core.Core.With(fields),
	}
}

// Check determines whether the supplied Entry should be logged (using the
// embedded LevelEnabler and possibly some extra logic). If the entry
// should be logged, the Core adds itself to the CheckedEntry and returns
// the result.
//
// Callers must use Check before calling Write.
func (core *Core) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if core.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, core)
	}

	return checkedEntry
}

// Write serializes the Entry and any Fields supplied at the log site and
// writes them to their destination.
//
// If called, Write should always log the Entry and Fields; it should not
// replicate the logic of Check.
func (core *Core) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	if entry.Stack != "" {
		// https://cloud.google.com/logging/docs/agent/configuration#special-fields
		// > If your log entry contains an exception stack trace,
		// > the exception stack trace should be set in this message JSON log field,
		// > so that the exception stack trace can be parsed and saved to Stackdriver Error Reporting.
		if entry.Message == "" {
			entry.Message = entry.Stack
		} else {
			entry.Message += "\n\n" + entry.Stack
		}
		entry.Stack = ""
	}

	loc := core.sourceLocationFromEntry(entry)

	if loc != nil {
		fields = append(fields, LogEntrySourceLocation(loc))
	}

	fields = core.dedupe(
		fields,
		"logging.googleapis.com/trace",
		"logging.googleapis.com/spanId",
		"logging.googleapis.com/operation",
		"logging.googleapis.com/sourceLocation",
	)

	return core.Core.Write(entry, fields)
}

// Sync flushes buffered logs (if any).
func (core *Core) Sync() error {
	return core.Core.Sync()
}

func (core *Core) dedupe(fields []zapcore.Field, targetNames ...string) []zapcore.Field {

	var m map[int]bool
	for _, targetName := range targetNames {
		latest := -1
		for idx, field := range fields {
			if targetName != field.Key {
				continue
			}
			if latest != -1 {
				if m == nil {
					m = make(map[int]bool)
				}
				m[idx] = true
			}
			latest = idx
		}
	}
	if len(m) == 0 {
		return fields
	}
	newFields := make([]zapcore.Field, 0, len(fields)-len(m))
	for idx, field := range fields {
		if !m[idx] {
			continue
		}
		newFields = append(newFields, field)
	}

	return newFields
}

func (core *Core) sourceLocationFromEntry(entry zapcore.Entry) *buildlog.LogEntrySourceLocation {
	if !entry.Caller.Defined {
		return nil
	}

	loc := &buildlog.LogEntrySourceLocation{
		File: entry.Caller.File,
		Line: int64(entry.Caller.Line),
	}

	if function := runtime.FuncForPC(entry.Caller.PC); function != nil {
		loc.Function = function.Name()
	}

	return loc
}
