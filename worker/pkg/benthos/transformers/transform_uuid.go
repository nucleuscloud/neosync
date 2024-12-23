package transformers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

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
			res, err := transformUuid(randomizer, value)
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

	return transformUuid(parsedOpts.randomizer, valueStr)
}

// Transforms an existing Uuid into a new UUid v5. This is mainly used to deterministically transform UUIDs using seed values into new UUIDs in situations where the existing UUIDs are considered sensitive.
func transformUuid(randomizer rng.Rand, value string) (*string, error) {
	if value == "" {
		return &value, nil
	}

	// output := uuid.NewSHA1(uuid.MustParse(value), []byte{randomizer.Float64()}).String()

	fmt.Println("the entry value", value)

	randomFloat := randomizer.Float64()
	randomBytes := make([]byte, 8)

	// Convert the float to its bit representation to ensure consistent byte conversion
	randomBits := math.Float64bits(randomFloat)
	binary.LittleEndian.PutUint64(randomBytes, randomBits)

	output := uuid.NewSHA1(uuid.MustParse(value), randomBytes).String()

	return &output, nil
}
