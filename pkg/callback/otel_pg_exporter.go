// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package callback

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	attrApmplusSpanKind         = "gen_ai.span.kind"
	attrApmplusSessionID        = "gen_ai.session.id"
	attrApmplusUserID           = "gen_ai.user.id"
	attrApmplusInput            = "gen_ai.input"
	attrApmplusOutput           = "gen_ai.output"
	attrApmplusPromptTokens     = "gen_ai.usage.prompt_tokens"
	attrApmplusCompletionTokens = "gen_ai.usage.completion_tokens"
	attrApmplusTotalTokens      = "gen_ai.usage.total_tokens"
	attrApmplusRequestModel     = "gen_ai.request.model"
	attrApmplusResponseModel    = "gen_ai.response.model"
	attrApmplusProviderName     = "gen_ai.provider.name"
	attrRunInfoComponent        = "runinfo.component"

	attrOpenInferencePromptTokens     = "llm.token_count.prompt"
	attrOpenInferenceCompletionTokens = "llm.token_count.completion"
	attrOpenInferenceModelName        = "llm.model_name"
)

// OTelPGExporter exports completed OpenTelemetry spans into Pomclaw's PostgreSQL tables.
type OTelPGExporter struct {
	traceStore model.TracesModel
	spanStore  model.SpansModel

	mu     sync.Mutex
	traces map[string]struct{}
}

// NewOTelPGExporter creates a PostgreSQL OpenTelemetry span exporter.
func NewOTelPGExporter(traceStore model.TracesModel, spanStore model.SpansModel) sdktrace.SpanExporter {
	return &OTelPGExporter{
		traceStore: traceStore,
		spanStore:  spanStore,
		traces:     make(map[string]struct{}),
	}
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

// exportSampler records all spans EXCEPT those whose parent is a remote
// not-sampled span (which is how suppressedExportCtx is identified).
//
// Why a custom sampler instead of ParentBased(AlwaysSample()):
//
//	ParentBased follows the local parent's sampling decision.
//	If any span in the Eino call-chain happens to be non-recording
//	(e.g. from go-zero middleware before our SDK was registered),
//	ParentBased would drop all child spans too.
//	This sampler ignores local-parent sampling and always records,
//	while still dropping the DB-query spans spawned by suppressedExportCtx.
type exportSampler struct{}

func (exportSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	psc := trace.SpanContextFromContext(p.ParentContext)
	// Drop only when the parent is a remote not-sampled span
	// (the marker used by suppressedExportCtx to break recursion).
	if psc.IsValid() && psc.IsRemote() && !psc.IsSampled() {
		return sdktrace.SamplingResult{Decision: sdktrace.Drop}
	}
	return sdktrace.SamplingResult{
		Decision:   sdktrace.RecordAndSample,
		Tracestate: psc.TraceState(),
	}
}

func (exportSampler) Description() string { return "EinoExportSampler" }

// NewLocalTracerProvider creates a tracer provider for local DB export.
func NewLocalTracerProvider(exporter sdktrace.SpanExporter) *sdktrace.TracerProvider {
	return sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
		sdktrace.WithSampler(exportSampler{}),
	)
}

func (e *OTelPGExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		if err := e.exportSpan(suppressedExportCtx, span); err != nil {
			return err
		}
	}
	return nil
}

func (e *OTelPGExporter) Shutdown(context.Context) error {
	return nil
}

func (e *OTelPGExporter) exportSpan(ctx context.Context, span sdktrace.ReadOnlySpan) error {
	traceID := span.SpanContext().TraceID().String()
	attrs := spanAttrs(span.Attributes())

	traceModel, err := e.ensureTrace(ctx, traceID, span, attrs)
	if err != nil {
		logx.Errorf("[OTelPGExporter] ensureTrace failed traceID=%s: %v", traceID, err)
		return err
	}

	spanModel := e.toSpanModel(span, attrs)
	if _, err := e.spanStore.Insert(ctx, spanModel); err != nil {
		logx.Errorf("[OTelPGExporter] failed to insert span: %v", err)
		return err
	}

	e.updateTraceAggregate(traceModel, span, attrs)
	if err := e.traceStore.Update(ctx, traceModel); err != nil {
		logx.Errorf("[OTelPGExporter] failed to update trace aggregate: %v", err)
		return err
	}

	return nil
}

func (e *OTelPGExporter) ensureTrace(ctx context.Context, traceID string, span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) (*model.Traces, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.traces[traceID]; ok {
		traceModel, err := e.traceStore.FindByOtelTraceID(ctx, traceID)
		if err == nil {
			return traceModel, nil
		}
	}

	traceModel, err := e.traceStore.FindByOtelTraceID(ctx, traceID)
	if err == nil {
		e.traces[traceID] = struct{}{}
		return traceModel, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	meta := map[string]any{
		"otel_trace_id": traceID,
	}

	metaJSON, _ := json.Marshal(meta)
	traceModel = &model.Traces{
		UserId:     nullString(attrString(attrs, attrApmplusUserID)),
		SessionKey: nullString(attrString(attrs, attrApmplusSessionID)),
		StartTime:  span.StartTime(),
		Name:       nullString(span.Name()),
		Status:     "running",
		Metadata:   sql.NullString{String: string(metaJSON), Valid: true},
	}

	if _, err := e.traceStore.Insert(ctx, traceModel); err != nil {
		logx.Errorf("[OTelPGExporter] failed to insert trace: %v", err)
		return nil, err
	}

	traceModel, err = e.traceStore.FindByOtelTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	e.traces[traceID] = struct{}{}
	return traceModel, nil
}

func (e *OTelPGExporter) toSpanModel(span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) *model.Spans {
	spanType := localSpanType(attrs)
	if spanType == "" {
		spanType = "chain"
	}

	status := "completed"
	level := "DEFAULT"
	errMsg := ""
	if span.Status().Code == codes.Error {
		status = "error"
		level = "ERROR"
		errMsg = span.Status().Description
	}
	if errMsg == "" {
		for _, event := range span.Events() {
			if event.Name == "exception" {
				for _, attr := range event.Attributes {
					if string(attr.Key) == "exception.message" {
						errMsg = attr.Value.AsString()
						break
					}
				}
			}
		}
	}

	meta := map[string]any{
		"otel_trace_id": traceIDString(span.SpanContext()),
		"otel_span_id":  span.SpanContext().SpanID().String(),
		"otel_kind":     span.SpanKind().String(),
	}
	if parent := span.Parent(); parent.IsValid() {
		meta["otel_parent_span_id"] = parent.SpanID().String()
	}
	metaJSON, _ := json.Marshal(meta)

	return &model.Spans{
		TraceId:       traceIDAsUUID(span.SpanContext().TraceID()),
		SpanType:      spanType,
		Name:          nullString(span.Name()),
		StartTime:     span.StartTime(),
		EndTime:       sql.NullTime{Time: span.EndTime(), Valid: !span.EndTime().IsZero()},
		DurationMs:    sql.NullInt64{Int64: span.EndTime().Sub(span.StartTime()).Milliseconds(), Valid: !span.EndTime().IsZero()},
		Status:        status,
		Error:         nullString(errMsg),
		Level:         level,
		Model:         nullString(firstStringAttr(attrs, attrApmplusRequestModel, attrApmplusResponseModel, attrOpenInferenceModelName)),
		Provider:      nullString(attrString(attrs, attrApmplusProviderName)),
		InputTokens:   nullInt64(inputTokens(attrs)),
		OutputTokens:  nullInt64(outputTokens(attrs)),
		InputPreview:  nullString(previewAttr(attrs, attrApmplusInput)),
		OutputPreview: nullString(previewAttr(attrs, attrApmplusOutput)),
		Metadata:      sql.NullString{String: string(metaJSON), Valid: true},
	}
}

func (e *OTelPGExporter) updateTraceAggregate(traceModel *model.Traces, span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) {
	traceModel.EndTime = sql.NullTime{Time: span.EndTime(), Valid: !span.EndTime().IsZero()}
	if !span.EndTime().IsZero() {
		traceModel.DurationMs = sql.NullInt64{Int64: span.EndTime().Sub(traceModel.StartTime).Milliseconds(), Valid: true}
	}
	traceModel.TotalInputTokens += inputTokens(attrs)
	traceModel.TotalOutputTokens += outputTokens(attrs)
	traceModel.SpanCount++

	switch localSpanType(attrs) {
	case "llm_call":
		traceModel.LlmCallCount++
	case "tool_call":
		traceModel.ToolCallCount++
	}

	if span.Status().Code == codes.Error {
		traceModel.Status = "error"
		traceModel.Error = nullString(span.Status().Description)
		return
	}
	if traceModel.Status == "" || traceModel.Status == "running" {
		traceModel.Status = "completed"
	}
}

func localSpanType(attrs map[string]attribute.Value) string {
	switch attrString(attrs, attrApmplusSpanKind) {
	case "workflow":
		return "agent"
	case "chatmodel":
		return "llm_call"
	case "toolsnode", "tool":
		return "tool_call"
	}
	switch attrString(attrs, attrRunInfoComponent) {
	case "ChatModel":
		return "llm_call"
	case "ToolsNode":
		return "tool_call"
	}
	return "chain"
}

func inputTokens(attrs map[string]attribute.Value) int64 {
	return firstIntAttr(attrs, attrApmplusPromptTokens, attrOpenInferencePromptTokens)
}

func outputTokens(attrs map[string]attribute.Value) int64 {
	return firstIntAttr(attrs, attrApmplusCompletionTokens, attrOpenInferenceCompletionTokens)
}

func previewAttr(attrs map[string]attribute.Value, key string) string {
	const maxPreviewLen = 4096
	value := attrString(attrs, key)
	if len(value) <= maxPreviewLen {
		return value
	}
	return value[:maxPreviewLen] + "...(truncated)"
}

func spanAttrs(attrs []attribute.KeyValue) map[string]attribute.Value {
	out := make(map[string]attribute.Value, len(attrs))
	for _, attr := range attrs {
		out[string(attr.Key)] = attr.Value
	}
	return out
}

func attrString(attrs map[string]attribute.Value, key string) string {
	value, ok := attrs[key]
	if !ok {
		return ""
	}
	return value.AsString()
}

func firstStringAttr(attrs map[string]attribute.Value, keys ...string) string {
	for _, key := range keys {
		if value := attrString(attrs, key); value != "" {
			return value
		}
	}
	return ""
}

func firstIntAttr(attrs map[string]attribute.Value, keys ...string) int64 {
	for _, key := range keys {
		value, ok := attrs[key]
		if ok {
			return value.AsInt64()
		}
	}
	return 0
}

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func nullInt64(value int64) sql.NullInt64 {
	return sql.NullInt64{Int64: value, Valid: value > 0}
}

func traceIDString(sc trace.SpanContext) string {
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

func traceIDAsUUID(traceID trace.TraceID) string {
	raw := traceID.String()
	if len(raw) != 32 {
		return raw
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", raw[0:8], raw[8:12], raw[12:16], raw[16:20], raw[20:32])
}

var _ sdktrace.SpanExporter = (*OTelPGExporter)(nil)
