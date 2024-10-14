package benthos_environment

import (
	"github.com/warpstreamlabs/bento/public/bloblang"
)

type BlobConfig struct {
}

type BlobOption func(cfg *BlobConfig)

func NewBlobEnvironment(opts ...BlobOption) (*bloblang.Environment, error) {
	env := bloblang.NewEnvironment()

	config := &BlobConfig{}
	for _, opt := range opts {
		opt(config)
	}

	return env, nil
}
