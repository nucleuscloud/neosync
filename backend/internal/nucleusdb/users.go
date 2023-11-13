package nucleusdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
)

func (d *NucleusDb) SetUserByAuth0Id(
	ctx context.Context,
	auth0UserId string,
) (*db_queries.NeosyncApiUser, error) {
	var userResp *db_queries.NeosyncApiUser
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		user, err := d.Q.GetUserByAuth0Id(ctx, dbtx, auth0UserId)
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			association, err := d.Q.GetUserAssociationByAuth0Id(ctx, dbtx, auth0UserId)
			if err != nil && !IsNoRows(err) {
				return err
			} else if err != nil && IsNoRows(err) {
				// create user, create association
				user, err = d.Q.CreateNonMachineUser(ctx, dbtx)
				if err != nil {
					return err
				}
				userResp = &user
				association, err = d.Q.CreateAuth0IdentityProviderAssociation(ctx, dbtx, db_queries.CreateAuth0IdentityProviderAssociationParams{
					UserID:          user.ID,
					Auth0ProviderID: auth0UserId,
				})
				if err != nil {
					return err
				}
			} else {
				user, err = d.Q.GetUser(ctx, dbtx, association.UserID)
				if err != nil && !IsNoRows(err) {
					if err != nil {
						return err
					}
				} else if err != nil && IsNoRows(err) {
					user, err = d.Q.CreateNonMachineUser(ctx, dbtx)
					if err != nil {
						return err
					}
				}
				userResp = &user
			}
			return nil
		}
		userResp = &user
		return nil
	}); err != nil {
		return nil, err
	}
	return userResp, nil
}

func (d *NucleusDb) SetPersonalAccount(
	ctx context.Context,
	userId pgtype.UUID,
) (*db_queries.NeosyncApiAccount, error) {
	var personalAccount *db_queries.NeosyncApiAccount
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		account, err := d.Q.GetPersonalAccountByUserId(ctx, dbtx, userId)
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			account, err = d.Q.CreatePersonalAccount(ctx, dbtx, "personal")
			if err != nil {
				return err
			}
			personalAccount = &account
			_, err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
				AccountID: account.ID,
				UserID:    userId,
			})
			if err != nil {
				return err
			}
		} else {
			personalAccount = &account
			_, err = d.Q.GetAccountUserAssociation(ctx, dbtx, db_queries.GetAccountUserAssociationParams{
				AccountId: account.ID,
				UserId:    userId,
			})
			if err != nil && !IsNoRows(err) {
				return err
			} else if err != nil && IsNoRows(err) {
				_, err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
					AccountID: account.ID,
					UserID:    userId,
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return personalAccount, nil
}

func (d *NucleusDb) CreateTeamAccount(
	ctx context.Context,
	userId pgtype.UUID,
	teamName string,
) (*db_queries.NeosyncApiAccount, error) {
	var teamAccount *db_queries.NeosyncApiAccount
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		accounts, err := d.Q.GetAccountsByUser(ctx, dbtx, userId)
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			accounts = []db_queries.NeosyncApiAccount{}
		}
		for _, account := range accounts {
			if strings.EqualFold(account.AccountSlug, teamName) {
				return nucleuserrors.NewAlreadyExists(fmt.Sprintf("team account with the name %s already exists", teamName))
			}
		}
		account, err := d.Q.CreateTeamAccount(ctx, dbtx, teamName)
		if err != nil {
			return err
		}
		teamAccount = &account
		_, err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
			AccountID: account.ID,
			UserID:    userId,
		})
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return teamAccount, nil
}

func (d *NucleusDb) CreateTeamAccountInvite(
	ctx context.Context,
	accountId pgtype.UUID,
	userId pgtype.UUID,
	email string,
	expiresAt pgtype.Timestamp,
) (*db_queries.NeosyncApiAccountInvite, error) {
	var accountInvite *db_queries.NeosyncApiAccountInvite
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		account, err := d.Q.GetAccount(ctx, dbtx, accountId)
		if err != nil {
			return err
		}
		if account.AccountType != 1 {
			return nucleuserrors.NewForbidden("unable to create team account invite: account type is not team")
		}

		// update any active invites for user to expired before creating new invite
		_, err = d.Q.UpdateActiveAccountInvitesToExpired(ctx, dbtx, db_queries.UpdateActiveAccountInvitesToExpiredParams{
			AccountId: accountId,
			Email:     email,
		})
		if err != nil {
			return err
		}

		invite, err := d.Q.CreateAccountInvite(ctx, dbtx, db_queries.CreateAccountInviteParams{
			AccountID:    accountId,
			SenderUserID: userId,
			Email:        email,
			ExpiresAt:    expiresAt,
		})
		if err != nil {
			return err
		}
		accountInvite = &invite
		return nil
	}); err != nil {

		return nil, err
	}
	return accountInvite, nil

}
