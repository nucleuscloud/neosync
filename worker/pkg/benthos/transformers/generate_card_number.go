package transformers

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	transformer_utils "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers/utils"
)

// +neosyncTransformerBuilder:generate:generateCardNumber

const defaultCCLength = 16
const defaultIIN = 400000

func init() {
	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("valid_luhn"))

	err := bloblang.RegisterFunctionV2("generate_card_number", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		luhn, err := args.GetBool("valid_luhn")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := generateCardNumber(luhn)
			if err != nil {
				return nil, fmt.Errorf("unable to run generate_card_number: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func (t *GenerateCardNumber) Generate(opts any) (any, error) {
	parsedOpts, ok := opts.(*GenerateCardNumberOpts)
	if !ok {
		return nil, errors.New("invalid parse opts")
	}

	return generateCardNumber(parsedOpts.validLuhn)
}

// Generates a 16 digit card number that can pass a luhn check if the validLuhn param is set to true. Otherwise will generate a random 16 digit card number.
func generateCardNumber(luhn bool) (int64, error) {
	var returnValue int64

	if luhn {
		val, err := generateValidLuhnCheckCardNumber()
		if err != nil {
			return 0, err
		}

		returnValue = val
	} else {
		val, err := transformer_utils.GenerateRandomInt64FixedLength(defaultCCLength)
		if err != nil {
			return 0, err
		}

		returnValue = val
	}

	return returnValue, nil
}

// generates a card number that passes luhn validation
func generateValidLuhnCheckCardNumber() (int64, error) {
	// To find the checksum digit on
	cardNo := make([]int, 0)
	for _, c := range fmt.Sprintf("%d", defaultIIN) {
		cardNo = append(cardNo, int(c-'0'))
	}

	// Actual account number
	cardNum := make([]int, 0)
	for _, c := range fmt.Sprintf("%d", defaultIIN) {
		cardNum = append(cardNum, int(c-'0'))
	}

	// Acc no (9 digits)
	nineDigits, err := transformer_utils.GenerateRandomInt64FixedLength(9)
	if err != nil {
		return 0, err
	}
	strNine := strconv.FormatInt(nineDigits, 10)
	for _, k := range strNine {
		digit, err := strconv.Atoi(string(k))
		if err != nil {
			continue
		}
		cardNo = append(cardNo, digit)
		cardNum = append(cardNum, digit)
	}

	// odd position digits
	for i := 0; i < len(cardNo); i += 2 {
		cardNo[i] *= 2
	}

	// deduct 9 from numbers greater than 9
	for i := 0; i < len(cardNo); i++ {
		if cardNo[i] > 9 {
			cardNo[i] -= 9
		}
	}

	// sum the digits
	s := 0
	for _, d := range cardNo {
		s += d
	}

	// calculate the checksum
	mod := s % 10
	checkSum := 0
	if mod != 0 {
		checkSum = 10 - mod
	}

	// append the checksum to the card number
	cardNum = append(cardNum, checkSum)

	// convert the card number to a string to build the card number format
	cardNumStr := ""
	for _, d := range cardNum {
		cardNumStr += fmt.Sprintf("%d", d)
	}

	// convert the card number back to an int to return it
	cardNumInt, err := strconv.ParseInt(cardNumStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return cardNumInt, nil
}
