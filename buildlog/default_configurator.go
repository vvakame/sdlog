package buildlog

import (
	"context"
	"net/http"
	"os"
	"strings"

	"go.opencensus.io/trace"
)

var _ Configurator = (*appengineConfigurator)(nil)

type appengineConfigurator struct{}

func (cfg *appengineConfigurator) ProjectID() string {
	if v := os.Getenv("GAE_APPLICATION"); v != "" {
		return v
	}
	if v := os.Getenv("GAE_LONG_APP_ID"); v != "" {
		return v
	}
	if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	}
	if v := os.Getenv("GCP_PROJECT"); v != "" {
		return v
	}
	if v := os.Getenv("GCLOUD_PROJECT"); v != "" {
		return v
	}

	return ""
}

func (cfg *appengineConfigurator) TraceInfo(ctx context.Context) (string, string) {
	if span := trace.FromContext(ctx); span != nil {
		return span.SpanContext().TraceID.String(), span.SpanContext().SpanID.String()
	}

	return "", ""
}

func (cfg *appengineConfigurator) RemoteIP(r *http.Request) string {
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
