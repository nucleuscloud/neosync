package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
)

func ToCustomTransformerDto(
	input *db_queries.NeosyncApiTransformer,
) *mgmtv1alpha1.CustomTransformer {
	return &mgmtv1alpha1.CustomTransformer{
		Id:     nucleusdb.UUIDString(input.ID),
		Name:   input.Name,
		Config: input.TransformerConfig.ToTransformerConfigDto(input.TransformerConfig),
		// CreatedAt: timestamppb.New(input.CreatedAt.Time),
		// UpdatedAt: timestamppb.New(input.UpdatedAt.Time),
		// CreatedByUserId: nucleusdb.UUIDString(input.CreatedByID),
		// UpdatedByUserId: nucleusdb.UUIDString(input.UpdatedByID),
		AccountId: nucleusdb.UUIDString(input.AccountID),
	}
}
