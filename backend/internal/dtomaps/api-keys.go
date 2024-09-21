package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToAccountApiKeyDto(
	input *db_queries.NeosyncApiAccountApiKey,
	cleartextKeyValue *string,
) *mgmtv1alpha1.AccountApiKey {
	return &mgmtv1alpha1.AccountApiKey{
		Id:          neosyncdb.UUIDString(input.ID),
		Name:        input.KeyName,
		AccountId:   neosyncdb.UUIDString(input.AccountID),
		CreatedById: neosyncdb.UUIDString(input.CreatedByID),
		CreatedAt:   timestamppb.New(input.CreatedAt.Time),
		UpdatedById: neosyncdb.UUIDString(input.UpdatedByID),
		UpdatedAt:   timestamppb.New(input.UpdatedAt.Time),
		KeyValue:    cleartextKeyValue,
		UserId:      neosyncdb.UUIDString(input.UserID),
		ExpiresAt:   timestamppb.New(input.ExpiresAt.Time),
	}
}
