package aws_s3

import (
	"errors"

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
	}
	return false
}
