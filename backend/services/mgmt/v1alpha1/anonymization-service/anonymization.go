package v1alpha_anonymizationservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	anon "github.com/nucleuscloud/neosync/internal/anonymizer"
)

func (s *Service) AnonymizeMany(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AnonymizeManyRequest],
) (*connect.Response[mgmtv1alpha1.AnonymizeManyResponse], error) {
	anonymizer, err := anon.NewAnonymizer(
		anon.WithTransformerMappings(req.Msg.TransformerMappings),
		anon.WithDefaultTransformers(req.Msg.DefaultTransformers),
		anon.WithHaltOnFailure(req.Msg.HaltOnFailure),
	)
	if err != nil {
		return nil, err
	}

	outputData, errors, err := anonymizer.AnonymizeMany(req.Msg.InputData)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&mgmtv1alpha1.AnonymizeManyResponse{
		OutputData: outputData,
		Errors:     errors,
	}), nil
}

// // AnonymizeSingle endpoint using the Anonymizer
// func (s *Service) AnonymizeSingle(
// 	ctx context.Context,
// 	req *connect.Request[mgmtv1alpha1.AnonymizeSingleRequest],
// ) (*connect.Response[mgmtv1alpha1.AnonymizeSingleResponse], error) {

// 	anonymizer := NewAnonymizer(
// 		WithTransformerMappings(req.Msg.TransformerMappings),
// 		WithDefaultTransformers(req.Msg.DefaultTransformers),
// 		WithHaltOnFailure(req.Msg.HaltOnFailure),
// 	)

// 	outputData, err := anonymizer.AnonymizeSingle(req.Msg.InputData)
// 	var errors []*mgmtv1alpha1.AnonymizeManyErrors
// 	if err != nil {
// 		errors = append(errors, &mgmtv1alpha1.AnonymizeManyErrors{
// 			InputIndex:   0,
// 			FieldPath:    "",
// 			ErrorMessage: err.Error(),
// 		})
// 	}

// 	resp := &mgmtv1alpha1.AnonymizeSingleResponse{
// 		OutputData: outputData,
// 		Errors:     errors,
// 	}

// 	return connect.NewResponse(resp), nil
// }
