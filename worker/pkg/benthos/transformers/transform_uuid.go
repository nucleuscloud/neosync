package transformers

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/google/uuid"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformUuid

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms an existing UUID to a UUID v5").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(bloblang.NewInt64Param("seed").Optional().Description("An optional seed value used for generating deterministic transformations."))

	err := bloblang.RegisterFunctionV2("transform_uuid", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		seedArg, err := args.GetOptionalInt64("seed")
		if err != nil {
			return nil, err
		}

		seed, err := transformer_utils.GetSeedOrDefault(seedArg)
		if err != nil {
			return nil, err
		}

		randomizer := rng.New(seed)

		return func() (any, error) {
			res := transformUuid(randomizer, value)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_uuid: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewTransformUuidOptsFromConfig(config *mgmtv1alpha1.TransformUuid) (*TransformUuidOpts, error) {
	if config == nil {
		return NewTransformUuidOpts(nil)
	}
	return NewTransformUuidOpts(
		nil,
	)
}

func (t *TransformUuid) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformUuidOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}

	return transformUuid(parsedOpts.randomizer, valueStr), nil
}

// Transforms an existing Uuid into a new UUid v5. This is mainly used to deterministically transform UUIDs using seed values into new UUIDs in situations where the existing UUIDs are considered sensitive.
func transformUuid(randomizer rng.Rand, value string) *string {
	if value == "" {
		return &value
	}

	inputUuid, err := uuid.Parse(value)
	if err != nil {
		return &value
	}

	// Generate a deterministic byte slice based on the input UUID and the randomizer
	seedBytes := make([]byte, 16)

	// Use the first 8 bytes from the input UUID
	copy(seedBytes, inputUuid[:8])

	randomInt := randomizer.Float64()
	binary.LittleEndian.PutUint64(seedBytes[8:], uint64(randomInt))

	// Create a new UUID using SHA1 namespace
	output := uuid.NewSHA1(uuid.Nil, seedBytes).String()

	return &output
}
