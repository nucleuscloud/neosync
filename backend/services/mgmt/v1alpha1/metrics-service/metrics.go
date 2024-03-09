package v1alpha1_metricsservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func (s *Service) GetRecordsReceivedCount(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetRecordsReceivedCountRequest],
) (*connect.Response[mgmtv1alpha1.GetRecordsReceivedCountResponse], error) {

	return connect.NewResponse(&mgmtv1alpha1.GetRecordsReceivedCountResponse{}), nil
}
