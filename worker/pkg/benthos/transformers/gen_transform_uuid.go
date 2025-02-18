
// Code generated by Neosync neosync_transformer_generator.go. DO NOT EDIT.
// source: transform_uuid.go

package transformers

import (
	"strings"
	"fmt"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	
)

type TransformUuid struct{}

type TransformUuidOpts struct {
	randomizer     rng.Rand
	
}

func NewTransformUuid() *TransformUuid {
	return &TransformUuid{}
}

func NewTransformUuidOpts(
  seedArg *int64,
) (*TransformUuidOpts, error) {
	seed, err := transformer_utils.GetSeedOrDefault(seedArg)
  if err != nil {
    return nil, fmt.Errorf("unable to generate seed: %w", err)
	}
	
	return &TransformUuidOpts{
		randomizer: rng.New(seed),
	}, nil
}

func (o *TransformUuidOpts) BuildBloblangString(
	valuePath string,
) string {
	fnStr := []string{
		"value:this.%s",
	}

	params := []any{
		valuePath,
	}

	

	template := fmt.Sprintf("transform_uuid(%s)", strings.Join(fnStr, ","))
	return fmt.Sprintf(template, params...)
}

func (t *TransformUuid) GetJsTemplateData() (*TemplateData, error) {
	return &TemplateData{
		Name: "transformUuid",
		Description: "Transforms an existing UUID to a UUID v5",
		Example: "",
	}, nil
}

func (t *TransformUuid) ParseOptions(opts map[string]any) (any, error) {
	transformerOpts := &TransformUuidOpts{}

	var seedArg *int64
	if seedValue, ok := opts["seed"].(int64); ok {
			seedArg = &seedValue
	}
	seed, err := transformer_utils.GetSeedOrDefault(seedArg)
	if err != nil {
		return nil, fmt.Errorf("unable to generate seed: %w", err)
	}
	transformerOpts.randomizer = rng.New(seed)

	return transformerOpts, nil
}
