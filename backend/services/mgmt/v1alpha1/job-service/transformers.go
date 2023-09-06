package v1alpha1_jobservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	neosyncdevv1alpha1 "github.com/nucleuscloud/neosync/k8s-operator/api/v1alpha1"
	operator_transformers "github.com/nucleuscloud/neosync/k8s-operator/pkg/transformers"
)

type Transformation string

const (
	Unspecified Transformation = "unspecified"
	Passthrough Transformation = "passthrough"
	UuidV4      Transformation = "uuid_v4"
	FirstName   Transformation = "first_name"
	PhoneNumber Transformation = "phone_number"
)

func (s *Service) GetTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Title: "Unspecified", Value: string(Unspecified)},
			{Title: "Passthrough", Value: string(Passthrough)},
			{Title: "Uuid V4", Value: string(UuidV4)},
			{Title: "First Name", Value: string(FirstName)},
			{Title: "Phone Number", Value: string(PhoneNumber)},
		},
	}), nil
}

func getColumnTransformer(value string) (*neosyncdevv1alpha1.ColumnTransformer, error) {
	if value == "passthrough" {
		return nil, nil
	}
	name, err := toOperatorTransformer(value)
	if err != nil {
		return nil, err
	}
	return &neosyncdevv1alpha1.ColumnTransformer{
		Name: name,
	}, nil
}

func toOperatorTransformer(value string) (string, error) {
	switch value {
	case string(UuidV4):
		return string(operator_transformers.UuidV4), nil
	case string(FirstName):
		return string(operator_transformers.FirstName), nil
	case string(PhoneNumber):
		return string(operator_transformers.PhoneNumber), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}

func fromOperatorTransformer(value string) (string, error) {
	switch value {
	case string(operator_transformers.UuidV4):
		return string(UuidV4), nil
	case string(operator_transformers.FirstName):
		return string(FirstName), nil
	case string(operator_transformers.PhoneNumber):
		return string(PhoneNumber), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
