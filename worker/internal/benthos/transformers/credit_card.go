package neosync_transformers

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/benthosdev/benthos/v4/public/bloblang"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	transformer_utils "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers/utils"
)

const defualtCCLength = 16
const defaultIIN = 400000

func init() {

	spec := bloblang.NewPluginSpec().
		Param(bloblang.NewBoolParam("luhn_check"))

	// register the plugin
	err := bloblang.RegisterFunctionV2("creditcardtransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

		luhn, err := args.GetBool("luhn_check")
		if err != nil {
			return nil, err
		}

		return func() (any, error) {
			res, err := GenerateCreditCard(luhn)
			return res, err
		}, nil
	})

	if err != nil {
		panic(err)
	}

}

// main transformer logic goes here
func GenerateCreditCard(luhn bool) (int64, error) {
	var returnValue int64

	if luhn {

		val, err := GenerateValidVLuhnCheckCreditCard()

		if err != nil {
			return 0, fmt.Errorf("unable to generate a luhn valid credit card number")
		}

		returnValue = val

	} else {

		val, err := transformer_utils.GenerateRandomInt(defualtCCLength)

		if err != nil {
			return 0, fmt.Errorf("unable to generate a random credit card number")
		}

		returnValue = val

	}

	return returnValue, nil
}

// generates a credit card number that passes luhn validation
func GenerateValidVLuhnCheckCreditCard() (int64, error) {

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
	seventh15 := rand.Perm(9)[:9]
	for _, i := range seventh15 {
		cardNo = append(cardNo, i)
		cardNum = append(cardNum, i)
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

	// convert the card number to a string
	cardNumStr := ""
	for _, d := range cardNum {
		cardNumStr += fmt.Sprintf("%d", d)
	}

	cardNumInt, err := strconv.ParseInt(cardNumStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return cardNumInt, nil
}
