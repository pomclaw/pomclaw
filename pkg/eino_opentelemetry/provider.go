package opentelemetry

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type OtelProvider struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
}

var providerRegistry struct {
	sync.Mutex
	provider *OtelProvider
}

// SetProvider registers the OpenTelemetry providers returned to upstream apmplus.
// It is a Pomclaw-local extension of the upstream package surface, enabled via
// go.mod replace for local trace persistence.
func SetProvider(tracerProvider *sdktrace.TracerProvider, meterProvider *metric.MeterProvider) {
	providerRegistry.Lock()
	defer providerRegistry.Unlock()
	providerRegistry.provider = &OtelProvider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}
}

func (p *OtelProvider) Shutdown(ctx context.Context) error {
	var err error
	if p.TracerProvider != nil {
		if err = p.TracerProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}
	if p.MeterProvider != nil {
		if err = p.MeterProvider.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}
	return err
}

// NewOpenTelemetryProvider mirrors the upstream constructor, but returns the
// locally registered provider when present. This lets the official apmplus
// callback run unchanged while spans are exported to Pomclaw's local database.
func NewOpenTelemetryProvider(opts ...Option) (*OtelProvider, error) {
	cfg := newConfig(opts)

	providerRegistry.Lock()
	registered := providerRegistry.provider
	providerRegistry.Unlock()
	if registered != nil {
		return registered, nil
	}

	var tracerProvider *sdktrace.TracerProvider
	if cfg.enableTracing {
		tracerProvider = cfg.sdkTracerProvider
		if tracerProvider == nil {
			tracerProvider = sdktrace.NewTracerProvider(sdktrace.WithSampler(cfg.sampler))
		}
	}

	var meterProvider *metric.MeterProvider
	if cfg.enableMetrics {
		meterProvider = cfg.meterProvider
		if meterProvider == nil {
			meterProvider = metric.NewMeterProvider()
		}
	}

	return &OtelProvider{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}
