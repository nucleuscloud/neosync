package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToCustomTransformerDto(
	input *db_queries.NeosyncApiTransformer,
) *mgmtv1alpha1.CustomTransformer {
	return &mgmtv1alpha1.CustomTransformer{
		Id:          nucleusdb.UUIDString(input.ID),
		Name:        input.Name,
		Type:        input.Type,
		Config:      input.TransformerConfig.ToTransformerConfigDto(input.TransformerConfig),
		Description: input.Description,
		CreatedAt:   timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(input.UpdatedAt.Time),
		AccountId:   nucleusdb.UUIDString(input.AccountID),
	}
}
