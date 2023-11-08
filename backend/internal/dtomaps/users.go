package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToAccountTypeDto(aType int16) mgmtv1alpha1.UserAccountType {
	switch aType {
	case 0:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_PERSONAL
	case 1:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_TEAM
	default:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func ToAccountInviteDto(input *db_queries.NeosyncApiAccountInvite) *mgmtv1alpha1.AccountInvite {
	return &mgmtv1alpha1.AccountInvite{
		Id:           nucleusdb.UUIDString(input.ID),
		AccountId:    nucleusdb.UUIDString(input.AccountID),
		SenderUserId: nucleusdb.UUIDString(input.SenderUserID),
		Email:        input.Email,
		Token:        input.Token,
		Accepted:     input.Accepted.Valid,
		CreatedAt:    timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:    timestamppb.New(input.UpdatedAt.Time),
		ExpiresAt:    timestamppb.New(input.ExpiresAt.Time),
	}
}
