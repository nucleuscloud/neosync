package transformers

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

// +neosyncTransformerBuilder:transform:transformHash

type TransformHashAlgo string

const (
	TransformHashAlgo_Md5    TransformHashAlgo = "md5"
	TransformHashAlgo_Sha1   TransformHashAlgo = "sha1"
	TransformHashAlgo_Sha256 TransformHashAlgo = "sha256"
	TransformHashAlgo_Sha512 TransformHashAlgo = "sha512"
)

func (a TransformHashAlgo) String() string {
	return string(a)
}

func isValidTransformHashAlgo(algo string) bool {
	return algo == string(TransformHashAlgo_Md5) ||
		algo == string(TransformHashAlgo_Sha1) ||
		algo == string(TransformHashAlgo_Sha256) ||
		algo == string(TransformHashAlgo_Sha512)
}

func NewTransformHashAlgoFromDto(dto mgmtv1alpha1.TransformHash_HashType) TransformHashAlgo {
	switch dto {
	case mgmtv1alpha1.TransformHash_HASH_TYPE_MD5:
		return TransformHashAlgo_Md5
	case mgmtv1alpha1.TransformHash_HASH_TYPE_SHA1:
		return TransformHashAlgo_Sha1
	case mgmtv1alpha1.TransformHash_HASH_TYPE_SHA256:
		return TransformHashAlgo_Sha256
	case mgmtv1alpha1.TransformHash_HASH_TYPE_SHA512:
		return TransformHashAlgo_Sha512
	default:
		return TransformHashAlgo_Md5
	}
}

func init() {
	spec := bloblang.NewPluginSpec().
		Description("Transforms input into a hash based on the configured algorithm").
		Category("any").
		Param(bloblang.NewAnyParam("value").Optional()).
		Param(
			bloblang.NewStringParam("algo").
				Description("The hashing algorithm to use. Oneof: md5, sha1, sha256, sha512. Defaults to md5.").
				Default(TransformHashAlgo_Md5.String()),
		)

	err := bloblang.RegisterFunctionV2("transform_hash", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		valuePtr, err := args.GetOptionalString("value")
		if err != nil {
			return nil, err
		}

		var value string
		if valuePtr != nil {
			value = *valuePtr
		}

		algoPtr, err := args.GetOptionalString("algo")
		if err != nil {
			return nil, err
		}

		var algo string
		if algoPtr != nil {
			algo = *algoPtr
		}
		if !isValidTransformHashAlgo(algo) {
			return nil, fmt.Errorf("invalid algo: %s", algo)
		}

		hashFunc := hashToFunction(TransformHashAlgo(algo))

		return func() (any, error) {
			res, err := hashFunc(value)
			if err != nil {
				return nil, fmt.Errorf("unable to run transform_hash: %w", err)
			}
			return res, nil
		}, nil
	})

	if err != nil {
		panic(err)
	}
}

func NewTransformHashOptsFromConfig(config *mgmtv1alpha1.TransformHash) (*TransformHashOpts, error) {
	if config == nil {
		return NewTransformHashOpts("md5")
	}
	return NewTransformHashOpts(config.GetAlgo().String())
}

func (t *TransformHash) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformHashOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	_ = parsedOpts

	valueStr, ok := value.(string)
	if !ok {
		return nil, errors.New("value is not a string")
	}
	_ = valueStr

	// todo: implement this
	return nil, nil
}

func hashToFunction(algo TransformHashAlgo) func(any) (string, error) {
	switch algo {
	case TransformHashAlgo_Md5:
		return func(a any) (string, error) {
			return generateHash(md5.New(), a) //nolint:gosec
		}
	case TransformHashAlgo_Sha1:
		return func(a any) (string, error) {
			return generateHash(sha1.New(), a) //nolint:gosec
		}
	case TransformHashAlgo_Sha256:
		return func(a any) (string, error) {
			return generateHash(sha256.New(), a)
		}
	case TransformHashAlgo_Sha512:
		return func(a any) (string, error) {
			return generateHash(sha512.New(), a)
		}
	default:
		return func(a any) (string, error) {
			return generateHash(md5.New(), a) //nolint:gosec
		}
	}
}

func generateHash(hasher hash.Hash, value any) (string, error) {
	valueStr, err := valueToStringForHash(value)
	if err != nil {
		return "", err
	}
	return hashToString(hasher, valueStr)
}

func valueToStringForHash(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	// todo: handle more types
	default:
		return "", fmt.Errorf("invalid value type: %T", value)
	}
}

func hashToString(hasher hash.Hash, input string) (string, error) {
	_, err := hasher.Write([]byte(input))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
