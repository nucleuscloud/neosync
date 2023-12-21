package aws_s3

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func IsNotFound(err error) bool {
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				return true
			}
		}

		var notFound *types.NoSuchKey
		if ok := errors.As(err, &notFound); ok {
			return true
		}

		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return true
		}
	}
	return false
}
