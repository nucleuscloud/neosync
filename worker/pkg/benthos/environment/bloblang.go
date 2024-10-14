package benthos_environment

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	presidioapi "github.com/nucleuscloud/neosync/internal/ee/presidio"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

type BlobConfig struct {
	piitext *transformPiiTextConfig
}

type transformPiiTextConfig struct {
	analyzeclient   presidioapi.AnalyzeInterface
	anonymizeclient presidioapi.AnonymizeInterface
	config          *mgmtv1alpha1.TransformPiiText
}

type BlobOption func(cfg *BlobConfig)

func NewBlobEnvironment(opts ...BlobOption) (*bloblang.Environment, error) {
	env := bloblang.NewEnvironment()

	config := &BlobConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.piitext != nil {
		err := transformers.NewBloblTransformPiiText(
			env,
			config.piitext.analyzeclient,
			config.piitext.anonymizeclient,
			config.piitext.config,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to register transform_pii_text blobl function: %w", err)
		}
	}

	return env, nil
}

func WithBlobTransformPiiText(
	analyzeclient presidioapi.AnalyzeInterface,
	anonymizeclient presidioapi.AnonymizeInterface,
	config *mgmtv1alpha1.TransformPiiText,
) BlobOption {
	return func(cfg *BlobConfig) {
		cfg.piitext = &transformPiiTextConfig{
			analyzeclient:   analyzeclient,
			anonymizeclient: anonymizeclient,
			config:          config,
		}
	}
}
