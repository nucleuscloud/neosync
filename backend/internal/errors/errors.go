package nucleuserrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Wraps an error as an internal error. If the error is already an RPC error, returns that instead
func New(err error) error {
	if err == nil {
		return status.Error(codes.Internal, "unknown error")
	}
	if e, ok := status.FromError(err); ok {
		return e.Err()
	}
	return NewInternalError(err.Error())
}

func NewNotFound(message string) error {
	return connect.NewError(connect.CodeNotFound, fmt.Errorf(message))
	// return status.Error(codes.NotFound, message)
}

func NewInternalError(message string) error {
	return connect.NewError(connect.CodeInternal, fmt.Errorf(message))
}

func NewBadRequest(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(message))
}

func NewAlreadyExists(message string) error {
	return connect.NewError(connect.CodeAlreadyExists, fmt.Errorf(message))
}

func NewForbidden(message string) error {
	return connect.NewError(connect.CodePermissionDenied, fmt.Errorf(message))
}

func NewUnauthenticated(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, fmt.Errorf(message))
}

// Identical to NewUnauthenticated
func NewUnauthorized(message string) error {
	return NewUnauthenticated(message)
}

func NewNotImplemented(message string) error {
	return connect.NewError(connect.CodeUnimplemented, fmt.Errorf(message))
}

func IsNotFound(err error) bool {
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return true
		}
		if connectErr := new(connect.Error); errors.As(err, &connectErr) && connectErr.Code() == connect.CodeNotFound {
			return true
		}
	}
	return false
}
