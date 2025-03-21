package transformers

import (
	context "context"
	"fmt"
	"reflect"

	tablesync_shared "github.com/nucleuscloud/neosync/worker/pkg/workflows/tablesync/shared"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

func RegisterTransformIdentityScramble(env *bloblang.Environment, allocator tablesync_shared.IdentityAllocator) error {
	spec := bloblang.NewPluginSpec().
		Description("Scrambles the identity of the input").
		Category("identity").
		Param(bloblang.NewStringParam("value").Description("The value to scramble").Optional()).
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
