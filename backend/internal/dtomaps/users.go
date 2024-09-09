package dtomaps

import (
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/nucleusdb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AccountType int16

const (
	AccountType_Personal AccountType = iota
	AccountType_Team
	AccountType_Enterprise
)

func ToAccountTypeDto(aType AccountType) mgmtv1alpha1.UserAccountType {
	switch aType {
	case 0:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_PERSONAL
	case 1:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_TEAM
	case 2:
		return mgmtv1alpha1.UserAccountType_USER_ACCOUNT_TYPE_ENTERPRISE
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
		Accepted:     input.Accepted.Bool,
		CreatedAt:    timestamppb.New(input.CreatedAt.Time),
		UpdatedAt:    timestamppb.New(input.UpdatedAt.Time),
		ExpiresAt:    timestamppb.New(input.ExpiresAt.Time),
	}
}

func ToUserAccount(input *db_queries.NeosyncApiAccount) *mgmtv1alpha1.UserAccount {
	return &mgmtv1alpha1.UserAccount{
		Id:   nucleusdb.UUIDString(input.ID),
		Name: input.AccountSlug,
		Type: ToAccountTypeDto(AccountType(input.AccountType)),
	}
}
