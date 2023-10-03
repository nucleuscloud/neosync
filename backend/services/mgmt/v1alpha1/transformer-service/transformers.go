package v1alpha1_transformerservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
)

type Transformation string

const (
	Invalid     Transformation = "invalid"
	Passthrough Transformation = "passthrough"
	UuidV4      Transformation = "uuid_v4"
	FirstName   Transformation = "first_name"
	PhoneNumber Transformation = "phone_number"
	Email       Transformation = "email "
)

func (s *Service) GetTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetTransformersResponse], error) {

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	fmt.Println("getting through to the API")

	transformers, err := s.db.Q.GetTransformersByAccount(ctx, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoTransformers := []*mgmtv1alpha1.Transformer{}
	for idx := range transformers {
		transformer := transformers[idx]
		dtoTransformers = append(dtoTransformers, dtomaps.ToTransformerDto(&transformer))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetTransformersResponse{
		Transformers: dtoTransformers,
	}), nil
}
