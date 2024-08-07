
// Code generated by Neosync neosync_transformer_generator.go. DO NOT EDIT.
// source: generate_categorical.go

package transformers

import (
	"fmt"
	
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	
)

type GenerateCategorical struct{}

type GenerateCategoricalOpts struct {
	randomizer     rng.Rand
	
	categories string
}

func NewGenerateCategorical() *GenerateCategorical {
	return &GenerateCategorical{}
}

func NewGenerateCategoricalOpts(
	categories string,
  seedArg *int64,
) (*GenerateCategoricalOpts, error) {
	seed, err := transformer_utils.GetSeedOrDefault(seedArg)
  if err != nil {
    return nil, fmt.Errorf("unable to generate seed: %w", err)
	}
	
	return &GenerateCategoricalOpts{
		categories: categories,
		randomizer: rng.New(seed),	
	}, nil
}

func (t *GenerateCategorical) GetJsTemplateData() (*TemplateData, error) {
	return &TemplateData{
		Name: "generateCategorical",
		Description: "Randomly selects a value from a defined set of categorical values.",
		Example: "",
	}, nil
}

func (t *GenerateCategorical) ParseOptions(opts map[string]any) (any, error) {
	transformerOpts := &GenerateCategoricalOpts{}

	if _, ok := opts["categories"].(string); !ok {
		return nil, fmt.Errorf("missing required argument. function: %s argument: %s", "generateCategorical", "categories")
	}
	categories := opts["categories"].(string)
	transformerOpts.categories = categories

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
