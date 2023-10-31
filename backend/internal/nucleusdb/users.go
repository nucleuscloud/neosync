package nucleusdb

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
)

func (d *NucleusDb) SetUserByAuth0Id(
	ctx context.Context,
	auth0UserId string,
) (*db_queries.NeosyncApiUser, error) {
	var user *db_queries.NeosyncApiUser
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
				user, err = d.Q.CreateUser(ctx, dbtx)
				if err != nil {
					return err
				}
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
					user, err = d.Q.CreateUser(ctx, dbtx)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return user, nil
}

func (d *NucleusDb) SetPersonalAccount(
	ctx context.Context,
	userId pgtype.UUID,
) (*db_queries.NeosyncApiAccount, error) {

	var account *db_queries.NeosyncApiAccount
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		account, err := d.Q.GetPersonalAccountByUserId(ctx, dbtx, userId)
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			account, err = d.Q.CreatePersonalAccount(ctx, dbtx, "personal")
			if err != nil {
				return err
			}
			_, err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
				AccountID: account.ID,
				UserID:    userId,
			})
			if err != nil {
				return err
			}
		} else {
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
	return account, nil
}
