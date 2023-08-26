package v1alpha1_transformerservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func (s *Service) GetTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Title: "Unspecified", Value: "JOB_MAPPING_TRANSFORMER_UNSPECIFIED"},
			{Title: "Passthrough", Value: "JOB_MAPPING_TRANSFORMER_PASSTHROUGH"},
			{Title: "Social Security Number", Value: "JOB_MAPPING_TRANSFORMER_SSN"},
			{Title: "Scramble", Value: "JOB_MAPPING_TRANSFORMER_SCRAMBLE"},
		},
	}), nil
}
