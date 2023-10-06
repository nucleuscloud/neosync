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
			{Title: "Passthrough", Value: string(Passthrough), Description: "The Passthrough transformer just passes the input value through to the desination with no changes."},
			{Title: "Uuid V4", Value: string(UuidV4), Description: "The UUID tranformer generates a new UUIDv4 id."},
			{Title: "First Name", Value: string(FirstName), Description: "The First Name tranformer can anonymize or generate a new phone number."},
			{Title: "Phone Number", Value: string(PhoneNumber), Description: "The Phone Number tranformer can anonymize or generate a new phone number."},
			{Title: "Email", Value: string(Email), Description: "The Email transformer can anonymize or generate a new email address.", Config: &mgmtv1alpha1.TransformerConfig{
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
