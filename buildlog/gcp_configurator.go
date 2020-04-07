package buildlog

import (
	"context"
	"net/http"
	"os"
	"strings"

	"go.opencensus.io/trace"
)

var _ Configurator = (*GCPDefaultConfigurator)(nil)

// GCPDefaultConfigurator works on AppEngine, Cloud Run, Compute Engine etc.
type GCPDefaultConfigurator struct{}

// ProjectID returns current GCP project id.
func (cfg *GCPDefaultConfigurator) ProjectID() string {
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

// TraceInfo returns TraceID and SpanID.
func (cfg *GCPDefaultConfigurator) TraceInfo(ctx context.Context) (string, string) {
	if span := trace.FromContext(ctx); span != nil {
		return span.SpanContext().TraceID.String(), span.SpanContext().SpanID.String()
	}

	return "", ""
}

// RemoteIP of client.
func (cfg *GCPDefaultConfigurator) RemoteIP(r *http.Request) string {
	return strings.SplitN(r.RemoteAddr, ":", 2)[0]
}
