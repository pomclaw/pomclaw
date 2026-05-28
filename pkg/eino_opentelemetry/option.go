package opentelemetry

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Option mirrors the upstream opentelemetry package option surface used by apmplus.
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	enableTracing bool
	enableMetrics bool

	exportInsecure bool
	exportEndpoint string
	exportHeaders  map[string]string

	resource          *resource.Resource
	sdkTracerProvider *sdktrace.TracerProvider

	sampler sdktrace.Sampler

	resourceAttributes []attribute.KeyValue
	resourceDetectors  []resource.Detector

	meterProvider *metric.MeterProvider
}

func newConfig(opts []Option) *config {
	cfg := &config{
		enableTracing: true,
		enableMetrics: true,
		sampler:       sdktrace.AlwaysSample(),
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}
	return cfg
}

func WithServiceName(serviceName string) Option {
	return WithResourceAttribute(attribute.String("service.name", serviceName))
}

func WithDeploymentEnvironment(env string) Option {
	return WithResourceAttribute(attribute.String("deployment.environment.name", env))
}

func WithServiceNamespace(namespace string) Option {
	return WithResourceAttribute(attribute.String("service.namespace", namespace))
}

func WithResourceAttribute(rAttr attribute.KeyValue) Option {
	return option(func(cfg *config) {
		cfg.resourceAttributes = append(cfg.resourceAttributes, rAttr)
	})
}

func WithResourceAttributes(rAttrs []attribute.KeyValue) Option {
	return option(func(cfg *config) {
		cfg.resourceAttributes = rAttrs
	})
}

func WithResource(resource *resource.Resource) Option {
	return option(func(cfg *config) {
		cfg.resource = resource
	})
}

func WithExportEndpoint(endpoint string) Option {
	return option(func(cfg *config) {
		cfg.exportEndpoint = endpoint
	})
}

func WithEnableTracing(enableTracing bool) Option {
	return option(func(cfg *config) {
		cfg.enableTracing = enableTracing
	})
}

func WithEnableMetrics(enableMetrics bool) Option {
	return option(func(cfg *config) {
		cfg.enableMetrics = enableMetrics
	})
}

func WithResourceDetector(detector resource.Detector) Option {
	return option(func(cfg *config) {
		cfg.resourceDetectors = append(cfg.resourceDetectors, detector)
	})
}

func WithHeaders(headers map[string]string) Option {
	return option(func(cfg *config) {
		cfg.exportHeaders = headers
	})
}

func WithInsecure() Option {
	return option(func(cfg *config) {
		cfg.exportInsecure = true
	})
}

func WithSampler(sampler sdktrace.Sampler) Option {
	return option(func(cfg *config) {
		cfg.sampler = sampler
	})
}

func WithSdkTracerProvider(sdkTracerProvider *sdktrace.TracerProvider) Option {
	return option(func(cfg *config) {
		cfg.sdkTracerProvider = sdkTracerProvider
	})
}

func WithMeterProvider(meterProvider *metric.MeterProvider) Option {
	return option(func(cfg *config) {
		cfg.meterProvider = meterProvider
	})
}
