// Pomclaw - Ultra-lightweight personal AI agent
// Powered by Eino Framework
// License: MIT
//
// Copyright (c) 2026 Pomclaw contributors

package callback

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"strings"
)

// LogExporter exports completed OpenTelemetry spans into Pomclaw's PostgreSQL tables.
type LogExporter struct {
}

// NewLogExporter creates a PostgreSQL OpenTelemetry span exporter.
func NewLogExporter() sdktrace.SpanExporter {
	return &LogExporter{}
}

func (e *LogExporter) ExportSpans(_ context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		if err := e.exportSpan(suppressedExportCtx, span); err != nil {
			return err
		}
	}
	return nil
}

func (e *LogExporter) Shutdown(context.Context) error {
	return nil
}

func (e *LogExporter) exportSpan(ctx context.Context, span sdktrace.ReadOnlySpan) error {
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	attrs := spanAttrs(span.Attributes())

	// Get parent span info
	parentSpanID := ""
	if parent := span.Parent(); parent.IsValid() {
		parentSpanID = parent.SpanID().String()
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("================== SPAN START ==================\n")
	sb.WriteString(fmt.Sprintf("**TraceID**: `%s`\n", traceID))
	sb.WriteString(fmt.Sprintf("**SpanID**: `%s`\n", spanID))
	if parentSpanID != "" {
		sb.WriteString(fmt.Sprintf("**ParentSpanID**: `%s`\n", parentSpanID))
	}
	sb.WriteString(fmt.Sprintf("**Name**: %s\n", span.Name()))
	sb.WriteString(fmt.Sprintf("**Kind**: %s\n", span.SpanKind().String()))
	sb.WriteString(fmt.Sprintf("**StartTime**: %s\n", span.StartTime().Format("2006-01-02 15:04:05.000")))
	sb.WriteString(fmt.Sprintf("**EndTime**: %s\n", span.EndTime().Format("2006-01-02 15:04:05.000")))
	sb.WriteString(fmt.Sprintf("**Duration**: %dms\n", span.EndTime().Sub(span.StartTime()).Milliseconds()))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n", span.Status().Code.String()))
	if span.Status().Description != "" {
		sb.WriteString(fmt.Sprintf("**StatusDesc**: %s\n", span.Status().Description))
	}

	// Log all attributes
	if len(attrs) > 0 {
		sb.WriteString("\n#### Attributes:\n")
		for _, keyVal := range span.Attributes() {
			key := string(keyVal.Key)
			val := keyVal.Value.AsInterface()
			sb.WriteString(fmt.Sprintf("- `%s`: %v\n", key, val))
		}
	}

	// Log events if any
	if len(span.Events()) > 0 {
		sb.WriteString("\n#### Events:\n")
		for _, event := range span.Events() {
			sb.WriteString(fmt.Sprintf("- **%s** at %s\n", event.Name, event.Time.Format("2006-01-02 15:04:05.000")))
			if len(event.Attributes) > 0 {
				for _, attr := range event.Attributes {
					sb.WriteString(fmt.Sprintf("  - `%s`: %v\n", string(attr.Key), attr.Value.AsInterface()))
				}
			}
		}
	}

	sb.WriteString("================== SPAN END ==================\n")

	logx.Info(sb.String())

	return nil
}

var _ sdktrace.SpanExporter = (*LogExporter)(nil)
