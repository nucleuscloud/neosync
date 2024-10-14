package v1alpha_anonymizationservice

import presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"

type Service struct {
	cfg       *Config
	analyze   presidioapi.AnalyzeInterface
	anonymize presidioapi.AnonymizeInterface
}

type Config struct {
	IsPresidioEnabled bool
}

func New(
	cfg *Config,
	analyzeclient presidioapi.AnalyzeInterface,
	anonymizeclient presidioapi.AnonymizeInterface,
) *Service {
	return &Service{
		cfg:       cfg,
		analyze:   analyzeclient,
		anonymize: anonymizeclient,
	}
}
