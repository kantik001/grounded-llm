package httpapi

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type pathTraceKey struct{}

// PathTrace correlates RAG pipeline steps under one request_id.
type PathTrace struct {
	reqID  string
	tenant string
	start  time.Time
	span   trace.Span
}

var (
	otelOnce   sync.Once
	otelTracer trace.Tracer
)

// InitTracing sets up OTLP HTTP exporter when OTEL_EXPORTER_OTLP_ENDPOINT is set.
// Safe to call multiple times; no-op when endpoint is empty.
func InitTracing(ctx context.Context) {
	otelOnce.Do(func() {
		endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
		if endpoint == "" {
			return
		}
		opts := []otlptracehttp.Option{}
		// Endpoint may be host:port or full URL; otlptracehttp accepts both via WithEndpoint / WithEndpointURL.
		if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
			opts = append(opts, otlptracehttp.WithEndpointURL(endpoint))
		} else {
			opts = append(opts, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			log.Printf("OTel exporter init failed: %v", err)
			return
		}
		svc := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
		if svc == "" {
			svc = "grounded-llm-server"
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(svc),
			)),
		)
		otel.SetTracerProvider(tp)
		otelTracer = tp.Tracer("grounded-llm")
		log.Printf("OpenTelemetry tracing enabled (endpoint=%s)", endpoint)
	})
}

// BeginPathTrace starts a debug trail (nil if reqID empty).
func BeginPathTrace(reqID, tenant string) *PathTrace {
	if reqID == "" {
		return nil
	}
	tr := &PathTrace{reqID: reqID, tenant: tenant, start: time.Now()}
	if otelTracer != nil {
		_, span := otelTracer.Start(context.Background(), "rag.request",
			trace.WithAttributes(
				attribute.String("request_id", reqID),
				attribute.String("tenant_id", tenant),
			),
		)
		tr.span = span
	}
	return tr
}

// ContextWithPathTrace stores the trace on ctx.
func ContextWithPathTrace(ctx context.Context, tr *PathTrace) context.Context {
	if tr == nil {
		return ctx
	}
	return context.WithValue(ctx, pathTraceKey{}, tr)
}

// PathTraceFrom returns the trace from ctx, if any.
func PathTraceFrom(ctx context.Context) *PathTrace {
	if ctx == nil {
		return nil
	}
	tr, _ := ctx.Value(pathTraceKey{}).(*PathTrace)
	return tr
}

// Step logs a named pipeline step and records an OTel child span when enabled.
func (t *PathTrace) Step(name string, fields map[string]any) {
	if t == nil {
		return
	}
	msg := "req=" + t.reqID + " step=" + name
	if t.tenant != "" {
		msg += " tenant=" + t.tenant
	}
	for k, v := range fields {
		msg += " " + k + "=" + FormatLogValue(v)
	}
	log.Print(msg)

	if otelTracer != nil {
		parent := context.Background()
		if t.span != nil {
			parent = trace.ContextWithSpan(parent, t.span)
		}
		_, span := otelTracer.Start(parent, "rag."+name)
		for k, v := range fields {
			span.SetAttributes(attribute.String(k, FormatLogValue(v)))
		}
		span.End()
	}
}

// End finishes the root OTel span, if any.
func (t *PathTrace) End() {
	if t != nil && t.span != nil {
		t.span.End()
		t.span = nil
	}
}

// ElapsedMS returns ms since BeginPathTrace.
func (t *PathTrace) ElapsedMS() int64 {
	if t == nil {
		return 0
	}
	return time.Since(t.start).Milliseconds()
}

// RequestID returns the correlated request id.
func (t *PathTrace) RequestID() string {
	if t == nil {
		return ""
	}
	return t.reqID
}

// AttachRequestID adds request_id to a JSON body when present.
func AttachRequestID(c *gin.Context, body gin.H) gin.H {
	if body == nil {
		body = gin.H{}
	}
	if c == nil {
		return body
	}
	if id := CtxRequestID(c); id != "" {
		body["request_id"] = id
	}
	return body
}

// MSSince returns elapsed milliseconds since start.
func MSSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
