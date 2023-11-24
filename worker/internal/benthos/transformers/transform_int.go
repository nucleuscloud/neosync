package transformers

import (
	_ "github.com/benthosdev/benthos/v4/public/components/io"
)

// const defaultIntLength = 4

// func init() {

// 	spec := bloblang.NewPluginSpec().
// 		Param(bloblang.NewInt64Param("value").Optional()).
// 		Param(bloblang.NewBoolParam("preserve_length").Optional()).
// 		Param(bloblang.NewInt64Param("int_length").Optional())

// 	// register the plugin
// 	err := bloblang.RegisterFunctionV2("randominttransformer", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {

// 		valuePtr, err := args.GetOptionalInt64("value")
// 		if err != nil {
// 			return nil, err
// 		}

// 		var value int64
// 		if valuePtr != nil {
// 			value = *valuePtr
// 		}
// 		preserveLengthPtr, err := args.GetOptionalBool("preserve_length")
// 		if err != nil {
// 			return nil, err
// 		}

// 		var preserveLength bool
// 		if preserveLengthPtr != nil {
// 			preserveLength = *preserveLengthPtr
// 		}

// 		intLengthPtr, err := args.GetOptionalInt64("int_length")
// 		if err != nil {
// 			return nil, err
// 		}

// 		var intLength int64
// 		if intLengthPtr != nil {
// 			intLength = *intLengthPtr
// 		}

// 		return func() (any, error) {
// 			res, err := GenerateRandomInt(value, preserveLength, intLength)
// 			return res, err
// 		}, nil
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// }

// func GenerateRandomInt(value int64, preserveLength bool, intLength int64) (int64, error) {
// 	var returnValue int64

// 	if preserveLength && intLength > 0 {
// 		return 0, fmt.Errorf("preserve length and int length params cannot both be true")
// 	}

// 	if value != 0 {

// 		if preserveLength {

// 			val, err := transformer_utils.GenerateRandomInt(int(transformer_utils.GetIntLength(value)))

// 			if err != nil {
// 				return 0, fmt.Errorf("unable to generate a random string with length")
// 			}

// 			returnValue = int64(val)

// 		} else if intLength > 0 {

// 			val, err := transformer_utils.GenerateRandomInt(int(intLength))

// 			if err != nil {
// 				return 0, fmt.Errorf("unable to generate a random string with length")
// 			}

// 			returnValue = int64(val)

// 		} else {

// 			val, err := transformer_utils.GenerateRandomInt(defaultIntLength)

// 			if err != nil {
// 				return 0, fmt.Errorf("unable to generate a random string with length")
// 			}

// 			returnValue = int64(val)

// 		}
// 	} else if intLength != 0 {

// 		val, err := transformer_utils.GenerateRandomInt(int(intLength))

// 		if err != nil {
// 			return 0, fmt.Errorf("unable to generate a random string with length")
// 		}

// 		returnValue = int64(val)

// 	} else {
// 		val, err := transformer_utils.GenerateRandomInt(defaultIntLength)

// 		if err != nil {
// 			return 0, fmt.Errorf("unable to generate a random string with length")
// 		}

// 		returnValue = int64(val)
// 	}

// 	return returnValue, nil
// }
