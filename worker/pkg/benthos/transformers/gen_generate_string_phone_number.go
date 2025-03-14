
// Code generated by Neosync neosync_transformer_generator.go. DO NOT EDIT.
// source: generate_string_phone_number.go

package transformers

import (
	"strings"
	"fmt"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
	
)

type GenerateStringPhoneNumber struct{}

type GenerateStringPhoneNumberOpts struct {
	randomizer     rng.Rand
	
	min int64
	max int64
}

func NewGenerateStringPhoneNumber() *GenerateStringPhoneNumber {
	return &GenerateStringPhoneNumber{}
}

func NewGenerateStringPhoneNumberOpts(
	minArg *int64,
	maxArg *int64,
  seedArg *int64,
) (*GenerateStringPhoneNumberOpts, error) {
	min := int64(9)
	if minArg != nil {
		min = *minArg
	}
	
	max := int64(15)
	if maxArg != nil {
		max = *maxArg
	}
	
	seed, err := transformer_utils.GetSeedOrDefault(seedArg)
  if err != nil {
    return nil, fmt.Errorf("unable to generate seed: %w", err)
	}
	
	return &GenerateStringPhoneNumberOpts{
		min: min,
		max: max,
		randomizer: rng.New(seed),
	}, nil
}

func (o *GenerateStringPhoneNumberOpts) BuildBloblangString(
) string {
	fnStr := []string{
		"min:%v",
		"max:%v",
	}

	params := []any{
	 	o.min,
	 	o.max,
	}

	

	template := fmt.Sprintf("generate_string_phone_number(%s)", strings.Join(fnStr, ","))
	return fmt.Sprintf(template, params...)
}

func (t *GenerateStringPhoneNumber) GetJsTemplateData() (*TemplateData, error) {
	return &TemplateData{
		Name: "generateStringPhoneNumber",
		Description: "Generates a random 10 digit phone number and returns it as a string with no hyphens.",
		Example: "",
	}, nil
}

func (t *GenerateStringPhoneNumber) ParseOptions(opts map[string]any) (any, error) {
	transformerOpts := &GenerateStringPhoneNumberOpts{}

	min, ok := opts["min"].(int64)
	if !ok {
		min = 9
	}
	transformerOpts.min = min

	max, ok := opts["max"].(int64)
	if !ok {
		max = 15
	}
	transformerOpts.max = max

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
