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
	"strconv"
	"strings"
	"sync"
	"time"

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
	attrApmplusFinishReason     = "gen_ai.response.finish_reason"
	attrRunInfoName             = "runinfo.name"
	attrRunInfoType             = "runinfo.type"
	attrRunInfoComponent        = "runinfo.component"

	attrOpenInferencePromptTokens     = "llm.token_count.prompt"
	attrOpenInferenceCompletionTokens = "llm.token_count.completion"
	attrOpenInferenceModelName        = "llm.model_name"
)

const pgExporterFlushInterval = 5 * time.Second

// PGExporter exports completed OpenTelemetry spans into Pomclaw's PostgreSQL tables.
type PGExporter struct {
	traceStore model.TracesModel
	spanStore  model.SpansModel

	traces sync.Map // map[otelTraceID]*cachedTrace

	stopCh       chan struct{}
	doneCh       chan struct{}
	shutdownOnce sync.Once
	shutdownErr  error
}

type cachedTrace struct {
	mu    sync.Mutex
	model *model.Traces
	spans []*model.Spans

	persisted bool
	dirty     bool
	version   int64

	firstInputSpanTime time.Time
	lastOutputSpanTime time.Time
}

// NewPGExporter creates a PostgreSQL OpenTelemetry span exporter.
func NewPGExporter(traceStore model.TracesModel, spanStore model.SpansModel) sdktrace.SpanExporter {
	exporter := &PGExporter{
		traceStore: traceStore,
		spanStore:  spanStore,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
	go exporter.flushLoop()
	return exporter
}

func (e *PGExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		if err := e.exportSpan(span); err != nil {
			return err
		}
	}
	return nil
}

func (e *PGExporter) Shutdown(ctx context.Context) error {
	e.shutdownOnce.Do(func() {
		close(e.stopCh)
		<-e.doneCh
		e.shutdownErr = e.flush(ctx)
	})
	return e.shutdownErr
}

func (e *PGExporter) exportSpan(span sdktrace.ReadOnlySpan) error {
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	attrs := spanAttrs(span.Attributes())

	spanModel := e.toSpanModel(span, attrs)
	traceCache := e.ensureCachedTrace(traceID, span, attrs)
	e.updateTraceAggregate(traceCache, spanID, spanModel, span, attrs)
	return nil
}

func (e *PGExporter) ensureCachedTrace(traceID string, span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) *cachedTrace {
	if value, ok := e.traces.Load(traceID); ok {
		return value.(*cachedTrace)
	}
	meta := map[string]any{
		"otel_trace_id": traceID,
	}
	metaJSON, _ := json.Marshal(meta)

	traceModel := &model.Traces{
		TraceId:    sql.NullString{String: traceID, Valid: true},
		UserId:     nullString(attrString(attrs, attrApmplusUserID)),
		SessionKey: nullString(attrString(attrs, attrApmplusSessionID)),
		StartTime:  span.StartTime(),
		Name:       nullString(span.Name()),
		Status:     "running",
		Metadata:   sql.NullString{String: string(metaJSON), Valid: true},
	}

	traceCache := &cachedTrace{
		model: traceModel,
		spans: make([]*model.Spans, 0),
		dirty: true,
	}
	actual, _ := e.traces.LoadOrStore(traceID, traceCache)
	return actual.(*cachedTrace)
}

func (e *PGExporter) flushLoop() {
	defer close(e.doneCh)

	ticker := time.NewTicker(pgExporterFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := e.flush(suppressedExportCtx); err != nil {
				logx.Errorf("[PGExporter] flush failed: %v", err)
			}
		case <-e.stopCh:
			return
		}
	}
}

func (e *PGExporter) flush(ctx context.Context) error {
	var firstErr error
	const inactiveTimeout = 30 * time.Second

	e.traces.Range(func(key, value any) bool {
		traceID, ok := key.(string)
		if !ok {
			return true
		}
		traceCache, ok := value.(*cachedTrace)
		if !ok {
			return true
		}

		if err := e.flushTrace(ctx, traceID, traceCache); err != nil {
			logx.Errorf("[PGExporter] failed to flush trace traceID=%s: %v", traceID, err)
			if firstErr == nil {
				firstErr = err
			}
			return true
		}

		// Delete inactive traces to prevent memory leak
		traceCache.mu.Lock()
		lastActivity := traceCache.lastOutputSpanTime
		if lastActivity.IsZero() {
			lastActivity = traceCache.model.StartTime
		}
		isInactive := time.Since(lastActivity) > inactiveTimeout && traceCache.persisted && len(traceCache.spans) == 0
		traceCache.mu.Unlock()

		if isInactive {
			e.traces.Delete(traceID)
			logx.Debugf("[PGExporter] cleaned up inactive trace traceID=%s", traceID)
		}

		return true
	})

	return firstErr
}

func (e *PGExporter) flushTrace(ctx context.Context, traceID string, traceCache *cachedTrace) error {
	traceCache.mu.Lock()
	if !traceCache.dirty && traceCache.persisted && len(traceCache.spans) == 0 {
		traceCache.mu.Unlock()
		return nil
	}
	traceModel := *traceCache.model
	spans := make([]*model.Spans, len(traceCache.spans))
	copy(spans, traceCache.spans)
	version := traceCache.version
	persisted := traceCache.persisted
	traceCache.mu.Unlock()

	if !persisted {
		existingTrace, err := e.traceStore.FindByOtelTraceID(ctx, traceID)
		switch {
		case err == nil:
			traceModel.Id = existingTrace.Id
			if err := e.traceStore.Update(ctx, &traceModel); err != nil {
				return err
			}
		case errors.Is(err, model.ErrNotFound):
			if _, err := e.traceStore.Insert(ctx, &traceModel); err != nil {
				return err
			}
			insertedTrace, err := e.traceStore.FindByOtelTraceID(ctx, traceID)
			if err != nil {
				return err
			}
			traceModel.Id = insertedTrace.Id
		default:
			return err
		}
	} else if err := e.traceStore.Update(ctx, &traceModel); err != nil {
		return err
	}

	// Batch insert all spans
	if len(spans) > 0 {
		if err := e.spanStore.BatchInsert(ctx, spans); err != nil {
			return fmt.Errorf("batch insert spans for traceID=%s: %w", traceID, err)
		}
	}

	traceCache.mu.Lock()
	traceCache.model.Id = traceModel.Id
	traceCache.persisted = true
	traceCache.spans = traceCache.spans[:0] // Clear spans slice after flushing
	if traceCache.version == version {
		traceCache.dirty = false
	}
	traceCache.mu.Unlock()
	return nil
}

func (e *PGExporter) toSpanModel(span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) *model.Spans {
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
		Provider:      nullString(spanProvider(attrs)),
		InputTokens:   nullInt64(inputTokens(attrs)),
		OutputTokens:  nullInt64(outputTokens(attrs)),
		InputPreview:  nullString(preview(firstNonEmpty(lastRoleContent(attrs, "gen_ai.prompt", "user"), attrString(attrs, attrApmplusInput)))),
		OutputPreview: nullString(preview(firstNonEmpty(lastRoleContent(attrs, "gen_ai.completion", "assistant"), attrString(attrs, attrApmplusOutput)))),
		FinishReason:  nullString(attrString(attrs, attrApmplusFinishReason)),
		ToolName:      nullString(spanToolName(attrs)),
		ToolCallId:    nullString(spanToolCallID(attrs)),
		Metadata:      sql.NullString{String: string(metaJSON), Valid: true},
	}
}

func (e *PGExporter) updateTraceAggregate(traceCache *cachedTrace, spanID string, spanModel *model.Spans, span sdktrace.ReadOnlySpan, attrs map[string]attribute.Value) {
	traceCache.mu.Lock()
	defer traceCache.mu.Unlock()

	traceModel := traceCache.model
	traceCache.spans = append(traceCache.spans, spanModel)

	if traceModel.StartTime.IsZero() || span.StartTime().Before(traceModel.StartTime) {
		traceModel.StartTime = span.StartTime()
		traceModel.Name = nullString(span.Name())
	}

	traceModel.EndTime = sql.NullTime{Time: span.EndTime(), Valid: !span.EndTime().IsZero()}
	if !span.EndTime().IsZero() {
		traceModel.DurationMs = sql.NullInt64{Int64: span.EndTime().Sub(traceModel.StartTime).Milliseconds(), Valid: true}
	}
	traceModel.TotalInputTokens += inputTokens(attrs)
	traceModel.TotalOutputTokens += outputTokens(attrs)
	traceModel.TotalCost += float64(totalTokens(attrs))
	traceModel.SpanCount++

	if userInput := lastRoleContent(attrs, "gen_ai.prompt", "user"); userInput != "" && localSpanType(attrs) == "llm_call" {
		if !traceModel.InputPreview.Valid {
			traceCache.firstInputSpanTime = span.StartTime()
			traceModel.InputPreview = nullString(preview(userInput))
		}
	}
	if assistantOutput := lastRoleContent(attrs, "gen_ai.completion", "assistant"); assistantOutput != "" {
		outputTime := span.EndTime()
		if outputTime.IsZero() {
			outputTime = span.StartTime()
		}
		if traceCache.lastOutputSpanTime.IsZero() || outputTime.After(traceCache.lastOutputSpanTime) || outputTime.Equal(traceCache.lastOutputSpanTime) {
			traceCache.lastOutputSpanTime = outputTime
			traceModel.OutputPreview = nullString(preview(assistantOutput))
		}
	}

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

	traceCache.version++
	traceCache.dirty = true
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

func spanProvider(attrs map[string]attribute.Value) string {
	if provider := attrString(attrs, attrApmplusProviderName); provider != "" {
		return provider
	}
	if localSpanType(attrs) == "llm_call" {
		if provider := attrString(attrs, attrRunInfoType); provider != "" {
			return provider
		}
	}
	return attrString(attrs, "gen_ai.system")
}

func spanToolName(attrs map[string]attribute.Value) string {
	if localSpanType(attrs) == "tool_call" {
		if name := attrString(attrs, attrRunInfoName); name != "" && name != "ToolNode" {
			return name
		}
	}
	for _, value := range roleContents(attrs, "gen_ai.completion") {
		if name := firstJSONField(value, "tool_name", "name"); name != "" {
			return name
		}
	}
	for _, value := range roleContents(attrs, "gen_ai.prompt") {
		if name := firstJSONField(value, "tool_name", "name"); name != "" {
			return name
		}
	}
	return ""
}

func spanToolCallID(attrs map[string]attribute.Value) string {
	for _, value := range roleContents(attrs, "gen_ai.completion") {
		if id := firstJSONField(value, "tool_call_id", "id"); id != "" {
			return id
		}
	}
	for _, value := range roleContents(attrs, "gen_ai.prompt") {
		if id := firstJSONField(value, "tool_call_id", "id"); id != "" {
			return id
		}
	}
	return ""
}

func inputTokens(attrs map[string]attribute.Value) int64 {
	return firstIntAttr(attrs, attrApmplusPromptTokens, attrOpenInferencePromptTokens)
}

func outputTokens(attrs map[string]attribute.Value) int64 {
	return firstIntAttr(attrs, attrApmplusCompletionTokens, attrOpenInferenceCompletionTokens)
}

func totalTokens(attrs map[string]attribute.Value) int64 {
	return firstIntAttr(attrs, attrApmplusTotalTokens)
}

func preview(value string) string {
	const maxPreviewLen = 4096
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func roleContents(attrs map[string]attribute.Value, prefix string) []string {
	type indexedContent struct {
		index   int
		content string
	}

	var values []indexedContent
	contentSuffix := ".content"
	for key, value := range attrs {
		if !strings.HasPrefix(key, prefix+".") || !strings.HasSuffix(key, contentSuffix) {
			continue
		}
		indexPart := strings.TrimSuffix(strings.TrimPrefix(key, prefix+"."), contentSuffix)
		index, err := strconv.Atoi(indexPart)
		if err != nil {
			continue
		}
		values = append(values, indexedContent{index: index, content: value.AsString()})
	}

	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j-1].index > values[j].index; j-- {
			values[j-1], values[j] = values[j], values[j-1]
		}
	}

	contents := make([]string, 0, len(values))
	for _, value := range values {
		if value.content != "" {
			contents = append(contents, value.content)
		}
	}
	return contents
}

func lastRoleContent(attrs map[string]attribute.Value, prefix, role string) string {
	maxIndex := -1
	content := ""
	roleSuffix := ".role"

	for key, value := range attrs {
		if !strings.HasPrefix(key, prefix+".") || !strings.HasSuffix(key, roleSuffix) {
			continue
		}
		if value.AsString() != role {
			continue
		}
		indexPart := strings.TrimSuffix(strings.TrimPrefix(key, prefix+"."), roleSuffix)
		index, err := strconv.Atoi(indexPart)
		if err != nil {
			continue
		}
		if index < maxIndex {
			continue
		}
		if candidate := attrString(attrs, fmt.Sprintf("%s.%d.content", prefix, index)); candidate != "" {
			maxIndex = index
			content = candidate
		}
	}

	return content
}

func firstJSONField(raw string, fields ...string) string {
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return ""
	}
	return firstJSONFieldFromAny(value, fields...)
}

func firstJSONFieldFromAny(value any, fields ...string) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, field := range fields {
			if found, ok := typed[field]; ok {
				if text, ok := found.(string); ok && text != "" {
					return text
				}
			}
		}
		for _, nested := range typed {
			if text := firstJSONFieldFromAny(nested, fields...); text != "" {
				return text
			}
		}
	case []any:
		for _, item := range typed {
			if text := firstJSONFieldFromAny(item, fields...); text != "" {
				return text
			}
		}
	}
	return ""
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

var _ sdktrace.SpanExporter = (*PGExporter)(nil)
