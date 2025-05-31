package buildlog_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vvakame/sdlog/buildlog"
	"go.opentelemetry.io/otel/trace/noop"
)

func Example_emitJSONPayload() {
	ctx := context.Background()
	tp := noop.NewTracerProvider()
	ctx, span := tp.Tracer("test").Start(ctx, "test")
	defer span.End()

	logEntry := buildlog.NewLogEntry(ctx)

	// for stable log output
	logEntry.Trace = "projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738"
	logEntry.SpanID = "2325d572b51a4ba6"
	logEntry.Time = buildlog.Time(time.Date(2019, 5, 18, 13, 47, 0, 0, time.UTC))
	logEntry.SourceLocation.File = "/tmp/123456/sdlog/buildlog/example_test.go"
	logEntry.SourceLocation.Line = 10

	b, err := json.Marshal(logEntry)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	// Output:
	// {"severity":"DEFAULT","time":"2019-05-18T13:47:00Z","logging.googleapis.com/trace":"projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738","logging.googleapis.com/spanId":"2325d572b51a4ba6","logging.googleapis.com/sourceLocation":{"file":"/tmp/123456/sdlog/buildlog/example_test.go","line":"10","function":"github.com/vvakame/sdlog/buildlog_test.Example_emitJSONPayload"}}
}

func Example_emitJSONPayloadWithEmbed() {
	ctx := context.Background()
	tp := noop.NewTracerProvider()
	ctx, span := tp.Tracer("test").Start(ctx, "test")
	defer span.End()

	type MyLog struct {
		Message string
		buildlog.LogEntry
	}

	buildMyLog := func(message string) *MyLog {
		myLog := &MyLog{
			Message:  message,
			LogEntry: *buildlog.NewLogEntry(ctx, buildlog.WithSourceLocationSkip(4)),
		}
		return myLog
	}

	logEntry := buildMyLog("Hi!")

	// for stable log output
	logEntry.Trace = "projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738"
	logEntry.SpanID = "2325d572b51a4ba6"
	logEntry.Time = buildlog.Time(time.Date(2019, 5, 18, 13, 47, 0, 0, time.UTC))
	logEntry.SourceLocation.File = "/tmp/123456/sdlog/buildlog/example_test.go"
	logEntry.SourceLocation.Line = 55

	b, err := json.Marshal(logEntry)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	// Output:
	// {"Message":"Hi!","severity":"DEFAULT","time":"2019-05-18T13:47:00Z","logging.googleapis.com/trace":"projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738","logging.googleapis.com/spanId":"2325d572b51a4ba6","logging.googleapis.com/sourceLocation":{"file":"/tmp/123456/sdlog/buildlog/example_test.go","line":"55","function":"github.com/vvakame/sdlog/buildlog_test.Example_emitJSONPayloadWithEmbed"}}
}

func Example_emitTextPayload() {
	ctx := context.Background()
	tp := noop.NewTracerProvider()
	ctx, span := tp.Tracer("test").Start(ctx, "test")
	defer span.End()

	logEntry := buildlog.NewLogEntry(ctx)
	logEntry.Message = "Hi!"

	// for stable log output
	logEntry.Trace = "projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738"
	logEntry.SpanID = "2325d572b51a4ba6"
	logEntry.Time = buildlog.Time(time.Date(2019, 5, 18, 13, 47, 0, 0, time.UTC))
	logEntry.SourceLocation.File = "/tmp/123456/sdlog/buildlog/example_test.go"
	logEntry.SourceLocation.Line = 55

	b, err := json.Marshal(logEntry)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	// Output:
	// {"severity":"DEFAULT","time":"2019-05-18T13:47:00Z","logging.googleapis.com/trace":"projects/foobar/traces/65ed3bb1ceb342ba0ca62fa64076c738","logging.googleapis.com/spanId":"2325d572b51a4ba6","logging.googleapis.com/sourceLocation":{"file":"/tmp/123456/sdlog/buildlog/example_test.go","line":"55","function":"github.com/vvakame/sdlog/buildlog_test.Example_emitTextPayload"},"message":"Hi!"}
}
