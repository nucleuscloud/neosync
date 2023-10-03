package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToTransformerDto(
	input *db_queries.NeosyncApiTransformer,
) *mgmtv1alpha1.Transformer {
	return &mgmtv1alpha1.Transformer{
		Id:              nucleusdb.UUIDString(input.ID),
		Name:            input.Name,
		Config:          input.TransformerConfig.ToDto(),
		CreatedAt:       timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:       timestamppb.New(input.UpdatedAt.Time),
		CreatedByUserId: nucleusdb.UUIDString(input.CreatedByID),
		UpdatedByUserId: nucleusdb.UUIDString(input.UpdatedByID),
		AccountId:       nucleusdb.UUIDString(input.AccountID),
	}
}
