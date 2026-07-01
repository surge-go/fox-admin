package application

import (
	"context"

	tracecore "github.com/surge-go/fox/core/tracing"
)

// initTracing 初始化 OpenTelemetry tracing provider。
func (app *Application) initTracing() error {
	cfg := app.toTracingConfig()
	if cfg == nil {
		return nil
	}

	provider, err := tracecore.New(context.Background(), cfg)
	if err != nil {
		return err
	}
	app.tracing = provider
	return nil
}

// toTracingConfig 将应用配置转换为 fox/core/tracing 配置。
func (app *Application) toTracingConfig() *tracecore.Config {
	if app.cfg == nil || app.cfg.Tracing == nil {
		return nil
	}

	cfg := app.cfg.Tracing
	return &tracecore.Config{
		Service:  toTracingServiceConfig(cfg.Service),
		Exporter: tracecore.Exporter(cfg.Exporter),
		OTLP:     toTracingOTLPConfig(cfg.OTLP),
		Sampler:  toTracingSamplerConfig(cfg.Sampler),
		Resource: toTracingResourceConfig(cfg.Resource),
		Batch:    toTracingBatchConfig(cfg.Batch),
	}
}

func toTracingServiceConfig(cfg *TracingServiceConfig) *tracecore.ServiceConfig {
	if cfg == nil {
		return nil
	}

	return &tracecore.ServiceConfig{
		Name:        cfg.Name,
		Namespace:   cfg.Namespace,
		Version:     cfg.Version,
		InstanceID:  cfg.InstanceID,
		Environment: cfg.Environment,
	}
}

func toTracingOTLPConfig(cfg *TracingOTLPConfig) *tracecore.OTLPConfig {
	if cfg == nil {
		return nil
	}

	return &tracecore.OTLPConfig{
		Endpoint:    cfg.Endpoint,
		Insecure:    cfg.Insecure,
		Headers:     cfg.Headers,
		Timeout:     cfg.Timeout,
		Compression: tracecore.Compression(cfg.Compression),
	}
}

func toTracingSamplerConfig(cfg *TracingSamplerConfig) *tracecore.SamplerConfig {
	if cfg == nil {
		return nil
	}

	return &tracecore.SamplerConfig{
		Type:  tracecore.Sampler(cfg.Type),
		Ratio: cfg.Ratio,
	}
}

func toTracingResourceConfig(cfg *TracingResourceConfig) *tracecore.ResourceConfig {
	if cfg == nil {
		return nil
	}

	return &tracecore.ResourceConfig{
		Attributes: cfg.Attributes,
	}
}

func toTracingBatchConfig(cfg *TracingBatchConfig) *tracecore.BatchConfig {
	if cfg == nil {
		return nil
	}

	return &tracecore.BatchConfig{
		MaxQueueSize:       cfg.MaxQueueSize,
		BatchTimeout:       cfg.BatchTimeout,
		ExportTimeout:      cfg.ExportTimeout,
		MaxExportBatchSize: cfg.MaxExportBatchSize,
	}
}
