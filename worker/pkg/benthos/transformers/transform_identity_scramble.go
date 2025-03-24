package transformers

import (
	context "context"
	"fmt"
	"reflect"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformIdentityScramble

func RegisterTransformIdentityScramble(env *bloblang.Environment, allocator tablesync_shared.IdentityAllocator) error {
	spec := bloblang.NewPluginSpec().
		Description("Scrambles the identity of the input").
		Category("int64").
		Param(bloblang.NewStringParam("value").Description("The value to scramble").Optional()).
		Param(bloblang.NewStringParam("token").Description("The token used to exchange for a block of identity values"))

	err := env.RegisterFunctionV2("transform_identity_scramble", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
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

func NewTransformIdentityScrambleOptsFromConfig(config *mgmtv1alpha1.ScrambleIdentity) (*TransformIdentityScrambleOpts, error) {
	if config == nil {
		return NewTransformIdentityScrambleOpts("")
	}
	return NewTransformIdentityScrambleOpts("")
}

func NewTransformIdentityScrambleOptsFromConfigWithToken(token string) (*TransformIdentityScrambleOpts, error) {
	return NewTransformIdentityScrambleOpts(token)
}

func (t *TransformIdentityScramble) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformIdentityScrambleOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	_ = parsedOpts
	return transformIdentityScramble(nil, "", nil)
}

func transformIdentityScramble(allocator tablesync_shared.IdentityAllocator, token string, value any) (any, error) {
	if value == nil {
		return nil, nil // todo: we should instead return a new scrambled value
	}

	var identity uint
	var err error
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.String:
		// return allocator.GetIdentity(context.Background(), token, v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// todo: check if value is negative
		input := uint(v.Uint())
		identity, err = allocator.GetIdentity(context.Background(), token, &input)
		if err != nil {
			return nil, fmt.Errorf("unable to get identity from value: %w", err)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		input := uint(v.Uint())
		identity, err = allocator.GetIdentity(context.Background(), token, &input)
		if err != nil {
			return nil, fmt.Errorf("unable to get identity from value: %w", err)
		}
	}
	if identity == 0 {
		return nil, fmt.Errorf("unable to get identity from value: %v", value)
	}

	return identity, nil
}
