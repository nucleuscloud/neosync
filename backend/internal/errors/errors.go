package nucleuserrors

import (
	"errors"
	"fmt"
	"strings"

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
	return connect.NewError(connect.CodeNotFound, errors.New(message))
}

func NewInternalError(message string) error {
	return connect.NewError(connect.CodeInternal, errors.New(message))
}

func NewBadRequest(message string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(message))
}

func NewAlreadyExists(message string) error {
	return connect.NewError(connect.CodeAlreadyExists, errors.New(message))
}

// Identical to NewUnauthorized
func NewForbidden(message string) error {
	return NewUnauthorized(message)
}

func NewUnauthenticated(message string) error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New(message))
}

func NewUnauthorized(message string) error {
	return connect.NewError(connect.CodePermissionDenied, errors.New(message))
}

func NewNotImplemented(message string) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New(message))
}

func NewNotImplementedProcedure(procedure string) error {
	return NewNotImplemented(
		fmt.Sprintf("%s is not implemented", strings.TrimPrefix(procedure, "/")),
	)
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
