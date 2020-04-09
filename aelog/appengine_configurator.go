package aelog

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/vvakame/sdlog/buildlog"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

var _ buildlog.Configurator = (*AppEngineConfigurator)(nil)

// AppEngineConfigurator works on AppEngine.
type AppEngineConfigurator struct{}

// ProjectID returns current GCP project id.
func (*AppEngineConfigurator) ProjectID() string {
	if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	}
	if v := os.Getenv("GAE_APPLICATION"); v != "" {
		// incoming `b~foobar` format
		ss := strings.SplitN(v, "~", 2)
		if len(ss) == 2 {
			return ss[1]
		}
		return v
	}
	if v := os.Getenv("GAE_LONG_APP_ID"); v != "" {
		return v
	}

	return ""
}

// TraceInfo returns TraceID and SpanID.
func (*AppEngineConfigurator) TraceInfo(ctx context.Context) (traceID, spanID string) {
	if span := trace.FromContext(ctx); span != nil {
		return span.SpanContext().TraceID.String(), span.SpanContext().SpanID.String()
	}

	r, ok := ctx.Value(contextHTTPRequestKey{}).(*http.Request)
	if !ok {
		// this case is common pattern in unit test.
		return "", ""
	}

	sc, ok := (&propagation.HTTPFormat{}).SpanContextFromRequest(r)
	if ok {
		return sc.TraceID.String(), sc.SpanID.String()
	}

	return "", ""
}

// RemoteIP of client.
func (*AppEngineConfigurator) RemoteIP(r *http.Request) string {
	remoteIP := ""
	if v := r.Header.Get("X-AppEngine-User-IP"); v != "" {
		remoteIP = v
	} else if v := r.Header.Get("X-Forwarded-For"); v != "" {
		remoteIP = v
	} else {
		remoteIP = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	return remoteIP
}
