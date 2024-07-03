package transformers

import (
	"errors"
	"fmt"
	"time"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	"github.com/nucleuscloud/neosync/worker/pkg/rng"
)

// +neosyncTransformerBuilder:generate:generateSSN

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewInt64Param("seed").Default(time.Now().UnixNano()))

	err := bloblang.RegisterFunctionV2("generate_ssn", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		seed, err := args.GetInt64("seed")
		if err != nil {
			return nil, err
		}
		randomizer := rng.New(seed)
		return func() (any, error) {
			val := generateRandomSSN(randomizer)
			return val, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}

func (t *GenerateSSN) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateSSNOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateRandomSSN(parsedOpts.randomizer), nil
}

/*
Generates a random social security number in the format AAA-GG-SSSS

An SSN is a nine-digit number typically formatted as "AAA-GG-SSSS". The three parts of an SSN are:

Area Number (AAA) - The first three digits, which historically represented the state or location where the SSN was issued.
However, post 2011, this is randomized due to the "randomization initiative".
Group Number (GG) - The next two digits, which are used to break down numbers into blocks available for assignment in a particular area.
Serial Number (SSSS) - The final four digits, which are assigned sequentially within each group.

This method ensures that the number does not start with 666, 000 or fall in the range 900-999, and does not use 00 in the group number or 0000 in the serial number
This is done to conform with how the US govt typically generates SSNs.
*/
func generateRandomSSN(randomizer rng.Rand) string {
	area := randomizer.Intn(899) + 100
	if area == 666 {
		area = 665
	}

	group := randomizer.Intn(89) + 10
	serial := randomizer.Intn(9999) + 1

	return fmt.Sprintf("%03d-%02d-%04d", area, group, serial)
}
