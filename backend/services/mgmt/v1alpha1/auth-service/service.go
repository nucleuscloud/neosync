package v1alpha1_authservice

type Service struct {
	cfg *Config
}

type Config struct {
	IsAuthEnabled bool
}

func New(
	cfg *Config,
) *Service {
	return &Service{cfg: cfg}
}
