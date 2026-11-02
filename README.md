# Cloud Logging compatible logger

[![Go Reference](https://pkg.go.dev/badge/github.com/vvakame/sdlog/v2.svg)](https://pkg.go.dev/github.com/vvakame/sdlog/v2)

This logger helps you to emit [special fields in structured payloads](https://docs.cloud.google.com/logging/docs/structured-logging#special-payload-fields).
You can emit LogEntry compatible JSON via stdout or stderr.

```sh
go get github.com/vvakame/sdlog/v2
```

## Packages

* [gcpslog](https://pkg.go.dev/github.com/vvakame/sdlog/v2/gcpslog) - [log/slog](https://pkg.go.dev/log/slog) handler.
* [buildlog](https://pkg.go.dev/github.com/vvakame/sdlog/v2/buildlog) - building blocks for logging.

## How to use

```go
slog.SetDefault(slog.New(gcpslog.HandlerOptions{}.NewHandler(os.Stdout)))

slog.InfoContext(ctx, "info message", slog.String("key", "value"))
```

When `HandlerOptions.ProjectID` is empty, the project ID is taken from `GOOGLE_CLOUD_PROJECT` or the metadata server,
and the trace and span IDs are taken from the OpenTelemetry span in the context by default.

For buildlog, see the examples in [the package documentation](https://pkg.go.dev/github.com/vvakame/sdlog/v2/buildlog#pkg-examples).
