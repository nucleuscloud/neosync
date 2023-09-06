package v1alpha1_jobservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	"github.com/nucleuscloud/neosync/k8s-operator/pkg/transformers"
)

func (s *Service) GetTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Title: "Unspecified", Value: "unspecified"},
			{Title: "Passthrough", Value: "passthrough"},
			{Title: "Uuid V4", Value: "uuidV4"},
			{Title: "First Name", Value: "firstName"},
			{Title: "Phone Number", Value: "phoneNumber"},
		},
	}), nil
}

func getColumnTransformer(value string) (*neosyncdevv1alpha1.ColumnTransformer, error) {
	if value == "passthrough" {
		return nil, nil
	}
	name, err := getTransformerName(value)
	if err != nil {
		return nil, err
	}
	return &neosyncdevv1alpha1.ColumnTransformer{
		Name: name,
	}, nil
}

func getTransformerName(value string) (string, error) {
	switch value {
	case "uuidV4":
		return string(transformers.UuidV4), nil
	case "firstName":
		return string(transformers.FirstName), nil
	case "phoneNumber":
		return string(transformers.PhoneNumber), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
