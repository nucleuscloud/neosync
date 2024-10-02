package v1alpha_anonymizationservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func (s *Service) AnonymizeMany(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.AnonymizeManyRequest],
) (*connect.Response[mgmtv1alpha1.AnonymizeManyResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.AnonymizeManyResponse{}), nil
}
