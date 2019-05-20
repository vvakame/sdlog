package buildlog

import (
	"context"
	"go.opencensus.io/trace"
	"net/http"
	"os"
	"strings"
)

var _ Configurator = (*appengineConfigurator)(nil)

type appengineConfigurator struct {}

func (cfg *appengineConfigurator) ProjectID() string {
	// TODO ちゃんと何らかのルールに従う

	if v := os.Getenv("GCP_PROJECT"); v != "" {
		return v
	} else if v := os.Getenv("GOOGLE_CLOUD_PROJECT"); v != "" {
		return v
	} else if v := os.Getenv("GCLOUD_PROJECT"); v != "" {
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
