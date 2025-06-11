package transformers

import (
	"crypto/md5"  //nolint:gosec
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
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
	defaultAlgo := TransformHashAlgo_Md5.String()
	if config == nil {
		return NewTransformHashOpts(&defaultAlgo)
	}
	var algo *string
	if config.Algo != nil {
		algoTypeStr := NewTransformHashAlgoFromDto(config.GetAlgo()).String()
		algo = &algoTypeStr
	}
	return NewTransformHashOpts(algo)
}

func (t *TransformHash) Transform(value, opts any) (any, error) {
	parsedOpts, ok := opts.(*TransformHashOpts)
	if !ok {
		return nil, fmt.Errorf("invalid parsed opts: %T", opts)
	}
	hashFunc := hashToFunction(TransformHashAlgo(parsedOpts.algo))
	return hashFunc(value)
}

func hashToFunction(algo TransformHashAlgo) func(any) (*string, error) {
	switch algo {
	case TransformHashAlgo_Md5:
		return func(a any) (*string, error) {
			return generateHash(md5.New(), a) //nolint:gosec
		}
	case TransformHashAlgo_Sha1:
		return func(a any) (*string, error) {
			return generateHash(sha1.New(), a) //nolint:gosec
		}
	case TransformHashAlgo_Sha256:
		return func(a any) (*string, error) {
			return generateHash(sha256.New(), a)
		}
	case TransformHashAlgo_Sha512:
		return func(a any) (*string, error) {
			return generateHash(sha512.New(), a)
		}
	default:
		return func(a any) (*string, error) {
			return generateHash(md5.New(), a) //nolint:gosec
		}
	}
}

func generateHash(hasher hash.Hash, value any) (*string, error) {
	if value == nil {
		return nil, nil
	} else if value == "" {
		emptyStr := ""
		return &emptyStr, nil
	}

	bits, err := convertAnyToBytes(value)
	if err != nil {
		return nil, fmt.Errorf("failed to convert value to bytes: %w", err)
	}
	hashStr, err := hashToString(hasher, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to hash value: %w", err)
	}
	return &hashStr, nil
}

func convertAnyToBytes(v any) ([]byte, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		return x, nil
	case string:
		return []byte(x), nil
	default:
		return json.Marshal(x)
	}
}

func hashToString(hasher hash.Hash, input []byte) (string, error) {
	_, err := hasher.Write(input)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
