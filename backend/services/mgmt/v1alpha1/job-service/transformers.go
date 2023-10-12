package v1alpha1_jobservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Transformation string

const (
	Invalid     Transformation = "invalid"
	Passthrough Transformation = "passthrough"
	UuidV4      Transformation = "uuid_v4"
	FirstName   Transformation = "first_name"
	PhoneNumber Transformation = "phone_number"
	Email       Transformation = "email"
)

func (s *Service) GetTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Value: string(Passthrough)},
			{Value: string(UuidV4)},
			{Value: string(FirstName)},
			{Value: string(PhoneNumber)},
			{
				Value: string(Email),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
						EmailConfig: &mgmtv1alpha1.EmailConfig{
							PreserveDomain: true,
							PreserveLength: true,
						},
					},
				}},
		},
	}), nil
}
