// Copyright 2023 The Connect Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package validate provides a [connect.Interceptor] that validates messages
// against constraints specified in their Protobuf schemas. Because the
// interceptor is powered by [protovalidate], validation is flexible,
// efficient, and consistent across languages - without additional code
// generation.
package validate

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"
)

// An Option configures an [Interceptor].
type Option interface {
	apply(*Interceptor)
}

// WithValidator configures the [Interceptor] to use a customized
// [protovalidate.Validator]. See [protovalidate.ValidatorOption] for the range
// of available customizations.
func WithValidator(validator protovalidate.Validator) Option {
	return optionFunc(func(i *Interceptor) {
		i.validator = validator
	})
}

// Interceptor is a [connect.Interceptor] that ensures that RPC request
// messages match the constraints expressed in their Protobuf schemas. It does
// not validate response messages.
//
// By default, Interceptors use a validator that lazily compiles constraints
// and works with any Protobuf message. This is a simple, widely-applicable
// configuration: after compiling and caching the constraints for a Protobuf
// message type once, validation is very efficient. To customize the validator,
// use [WithValidator] and [protovalidate.ValidatorOption].
//
// RPCs with invalid request messages short-circuit with an error. The error
// always uses [connect.CodeInvalidArgument] and has a [detailed representation
// of the error] attached as a [connect.ErrorDetail].
//
// This interceptor is primarily intended for use on handlers. Client-side use
// is possible, but discouraged unless the client always has an up-to-date
// schema.
//
// [detailed representation of the error]: https://pkg.go.dev/buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate#Violations
type Interceptor struct {
	validator protovalidate.Validator
}

// NewInterceptor builds an Interceptor. The default configuration is
// appropriate for most use cases.
func NewInterceptor(opts ...Option) (*Interceptor, error) {
	var interceptor Interceptor
	for _, opt := range opts {
		opt.apply(&interceptor)
	}

	if interceptor.validator == nil {
		validator, err := protovalidate.New()
		if err != nil {
			return nil, fmt.Errorf("construct validator: %w", err)
		}
		interceptor.validator = validator
	}

	return &interceptor, nil
}

// WrapUnary implements connect.Interceptor.
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if err := validate(i.validator, req.Any()); err != nil {
			return nil, err
		}
		return next(ctx, req)
	}
}

// WrapStreamingClient implements connect.Interceptor.
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return &streamingClientInterceptor{
			StreamingClientConn: next(ctx, spec),
			validator:           i.validator,
		}
	}
}

// WrapStreamingHandler implements connect.Interceptor.
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		return next(ctx, &streamingHandlerInterceptor{
			StreamingHandlerConn: conn,
			validator:            i.validator,
		})
	}
}

type streamingClientInterceptor struct {
	connect.StreamingClientConn

	validator protovalidate.Validator
}

func (s *streamingClientInterceptor) Send(msg any) error {
	if err := validate(s.validator, msg); err != nil {
		return err
	}
	return s.StreamingClientConn.Send(msg)
}

type streamingHandlerInterceptor struct {
	connect.StreamingHandlerConn

	validator protovalidate.Validator
}

func (s *streamingHandlerInterceptor) Receive(msg any) error {
	if err := s.StreamingHandlerConn.Receive(msg); err != nil {
		return err
	}
	return validate(s.validator, msg)
}

type optionFunc func(*Interceptor)

func (f optionFunc) apply(i *Interceptor) { f(i) }

func validate(validator protovalidate.Validator, msg any) error {
	protoMsg, ok := msg.(proto.Message)
	if !ok {
		return fmt.Errorf("expected proto.Message, got %T", msg)
	}
	err := validator.Validate(protoMsg)
	if err == nil {
		return nil
	}
	connectErr := connect.NewError(connect.CodeInvalidArgument, err)
	if validationErr := new(protovalidate.ValidationError); errors.As(err, &validationErr) {
		if detail, err := connect.NewErrorDetail(validationErr.ToProto()); err == nil {
			connectErr.AddDetail(detail)
		}
	}
	return connectErr
}
