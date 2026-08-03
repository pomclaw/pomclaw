package callback

import (
	"context"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// exportSampler records only spans with gen_ai attributes (Eino LLM-related).
// Drops all other spans including:
//   - go-zero sqlx database operations
//   - middleware spans
//   - other non-LLM instrumentation
//
// This ensures only relevant AI/LLM execution chains are persisted.
type exportSampler struct{}

func (exportSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	psc := trace.SpanContextFromContext(p.ParentContext)

	// Drop remote not-sampled spans (exporter's own DB calls via suppressedExportCtx)
	// to prevent infinite recursion: span → export → DB query → span → ...
	if psc.IsValid() && psc.IsRemote() && !psc.IsSampled() {
		return sdktrace.SamplingResult{Decision: sdktrace.Drop}
	}

	// Drop SQL spans (go-zero sqlx instrumentation)
	if p.Name == "sql" {
		return sdktrace.SamplingResult{Decision: sdktrace.Drop}
	}

	return sdktrace.SamplingResult{Decision: sdktrace.RecordAndSample}
}

func (exportSampler) Description() string {
	return "GenAISampler"
}

// suppressedExportCtx is a context with a remote not-sampled span.
// Used for all DB calls inside ExportSpans to prevent go-zero's sqlx
// instrumentation from creating new recorded spans, which would cause
// infinite recursion (span → export → DB query → span → ...).
// With the default ParentBased(AlwaysSample()) sampler, a remote
// not-sampled parent causes all child spans to be dropped.
var suppressedExportCtx = func() context.Context {
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x01},
		SpanID:     trace.SpanID{0x01},
		TraceFlags: trace.TraceFlags(0), // not sampled
		Remote:     true,
	})
	return trace.ContextWithRemoteSpanContext(context.Background(), sc)
}()

// NewLocalTracerProvider creates a tracer provider for local DB export.
// Uses exportSampler to record only gen_ai spans (LLM-related), filtering out
// database operations, middleware, and other non-relevant instrumentation.
func NewLocalTracerProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(exportSampler{}),
	)
}
