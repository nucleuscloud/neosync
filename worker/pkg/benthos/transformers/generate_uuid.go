package transformers

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
)

// +neosyncTransformerBuilder:generate:generateUUID

func init() {
	spec := bloblang.NewPluginSpec().Description("Generates a new UUIDv4 id.").
		Category("string").
		Param(bloblang.NewBoolParam("include_hyphens").
			Default(true).
			Description("Determines whether the generated UUID should include hyphens. If set to true, the UUID will be formatted with hyphens (e.g., d853d251-e135-4fe4-a4eb-0aea6bfaf645). If set to false, the hyphens will be omitted (e.g., d853d251e1354fe4a4eb0aea6bfaf645)."),
		)

	err := bloblang.RegisterFunctionV2(
		"generate_uuid",
		spec,
		func(args *bloblang.ParsedParams) (bloblang.Function, error) {
			include_hyphen, err := args.GetBool("include_hyphens")
			if err != nil {
				return nil, err
			}

			return func() (any, error) {
				val := generateUuid(include_hyphen)
				return val, nil
			}, nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func NewGenerateUUIDOptsFromConfig(config *mgmtv1alpha1.GenerateUuid) (*GenerateUUIDOpts, error) {
	if config == nil {
		return NewGenerateUUIDOpts(nil)
	}
	return NewGenerateUUIDOpts(config.IncludeHyphens)
}

func (t *GenerateUUID) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateUUIDOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}

	return generateUuid(parsedOpts.includeHyphens), nil
}

func generateUuid(includeHyphens bool) string {
	newuuid := uuid.NewString()
	if includeHyphens {
		return newuuid
	}
	/*generates uuid with no hyphens
	for postgres, if the dest column is defined as a UUID column then it will automatically
	convert the UUID with no hyphens to having hyphens
	so this is more useful for string columns or other dbs that won't do the automatic
	conversion if you want don't want your UUIDs to have hyphens on purpose */
	return strings.ReplaceAll(newuuid, "-", "")
}
