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

var _ buildlog.Configurator = (*aeConfigurator)(nil)

type aeConfigurator struct{}

func (aeConfigurator) ProjectID() string {
	if v := os.Getenv("GAE_APPLICATION"); v != "" {
		return v
	}
	if v := os.Getenv("GAE_LONG_APP_ID"); v != "" {
		return v
	}
	if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	}

	panic("environment variable GAE_APPLICATION is empty")
}

func (aeConfigurator) TraceInfo(ctx context.Context) (traceID, spanID string) {
	if span := trace.FromContext(ctx); span != nil {
		return span.SpanContext().TraceID.String(), span.SpanContext().SpanID.String()
	}

	r, ok := ctx.Value(contextHTTPRequestKey{}).(*http.Request)
	if !ok {
		panic("ctx doesn't have trace & spanId info")
	}

	sc, ok := (&propagation.HTTPFormat{}).SpanContextFromRequest(r)
	if !ok {
		panic("ctx doesn't have trace & spanId info")
	}

	return sc.TraceID.String(), sc.SpanID.String()
}

func (aeConfigurator) RemoteIP(r *http.Request) string {
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
