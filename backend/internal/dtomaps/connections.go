package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToConnectionDto(
	input *db_queries.NeosyncApiConnection,
	canViewSensitive bool,
) (*mgmtv1alpha1.Connection, error) {
	ccDto, err := input.ConnectionConfig.ToDto(canViewSensitive)
	if err != nil {
		return nil, err
	}
	return &mgmtv1alpha1.Connection{
		Id:               neosyncdb.UUIDString(input.ID),
		Name:             input.Name,
		ConnectionConfig: ccDto,
		CreatedAt:        timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:        timestamppb.New(input.UpdatedAt.Time),
		CreatedByUserId:  neosyncdb.UUIDString(input.CreatedByID),
		UpdatedByUserId:  neosyncdb.UUIDString(input.UpdatedByID),
		AccountId:        neosyncdb.UUIDString(input.AccountID),
	}, nil
}
