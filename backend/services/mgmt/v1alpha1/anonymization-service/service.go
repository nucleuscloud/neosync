package v1alpha_anonymizationservice

import "go.opentelemetry.io/otel/metric"

type Service struct {
	cfg   *Config
	meter metric.Meter // optional
}

type Config struct{}

func New(
	cfg *Config,
	meter metric.Meter,
) *Service {
	return &Service{
		cfg:   cfg,
		meter: meter,
	}
}
