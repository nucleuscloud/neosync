package v1alpha_anonymizationservice

type Service struct {
	cfg *Config
}

type Config struct{}

func New(
	cfg *Config,
) *Service {
	return &Service{
		cfg: cfg,
	}
}
