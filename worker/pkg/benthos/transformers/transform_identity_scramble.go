package transformers

import (
	context "context"
	"errors"
	"fmt"
	"reflect"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

func RegisterTransformIdentityScramble(
	env *bloblang.Environment,
	allocator tablesync_shared.IdentityAllocator,
) error {
	spec := bloblang.NewPluginSpec().
		Description("Scrambles the identity of the input").
		Category("int64").
		Param(bloblang.NewAnyParam("value").Description("The value to scramble").Optional()).
		Param(bloblang.NewStringParam("token").Description("The token used to exchange for a block of identity values"))

	err := env.RegisterFunctionV2(
		"transform_identity_scramble",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			value, err := args.Get("value")
			if err != nil {
				return nil, err
			}
			token, err := args.GetString("token")
			if err != nil {
				return nil, err
			}
			return func() (any, error) {
				return transformIdentityScramble(allocator, token, value)
			}, nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to register transform_identity_scramble: %w", err)
	}
	return nil
}

func NewTransformIdentityScrambleOptsFromConfig(
	config *mgmtv1alpha1.TransformScrambleIdentity,
) (*TransformIdentityScrambleOpts, error) {
	if config == nil {
		return NewTransformIdentityScrambleOpts("token-not-implemented")
	}
	return NewTransformIdentityScrambleOpts("token-not-implemented")
}

func NewTransformIdentityScrambleOptsFromConfigWithToken(
	token string,
) (*TransformIdentityScrambleOpts, error) {
	return NewTransformIdentityScrambleOpts(token)
}

func (t *TransformIdentityScramble) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformIdentityScrambleOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	_ = parsedOpts
	return transformIdentityScramble(nil, "token-not-implemented", value)
}

func transformIdentityScramble(
	allocator tablesync_shared.IdentityAllocator,
	token string,
	value any,
) (any, error) {
	if value == nil {
		return nil, nil // todo: we should instead return a new scrambled value
	}

	var identity uint
	var err error
	// todo: should add better support for different integer types to ensure overflow does not occur
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var input uint
		if v.Int() < 0 {
			input = 0
		} else {
			input = uint(v.Int()) //nolint:gosec // safe to convert since we check for negative values above
		}
		identity, err = allocator.GetIdentity(context.Background(), token, &input)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to get identity from value with token %s: %w",
				token,
				err,
			)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		input := uint(v.Uint())
		identity, err = allocator.GetIdentity(context.Background(), token, &input)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to get identity from value with token %s: %w",
				token,
				err,
			)
		}
	default:
		return nil, fmt.Errorf("unable to get identity from value as input was %T", value)
	}
	if identity == 0 {
		return nil, errors.New("unable to get identity from value as generated identity was 0")
	}

	return identity, nil
}
