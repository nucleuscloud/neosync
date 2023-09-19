package nucleusdb

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
)

func (d *NucleusDb) SetUserByAuth0Id(
	ctx context.Context,
	auth0UserId string,
) (*db_queries.NeosyncApiUser, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Println(err)
		}
	}()
	q := d.Q.WithTx(tx)

	user, err := q.GetUserByAuth0Id(ctx, auth0UserId)
	if err != nil && !IsNoRows(err) {
		return nil, err
	} else if err != nil && IsNoRows(err) {
		association, err := q.GetUserAssociationByAuth0Id(ctx, auth0UserId)
		if err != nil && !IsNoRows(err) {
			return nil, err
		} else if err != nil && IsNoRows(err) {
			// create user, create association
			user, err = q.CreateUser(ctx)
			if err != nil {
				return nil, err
			}
			association, err = q.CreateAuth0IdentityProviderAssociation(ctx, db_queries.CreateAuth0IdentityProviderAssociationParams{
				UserID:          user.ID,
				Auth0ProviderID: auth0UserId,
			})
			if err != nil {
				return nil, err
			}
		} else {
			user, err = q.GetUser(ctx, association.UserID)
			if err != nil && !IsNoRows(err) {
				if err != nil {
					return nil, err
				}
			} else if err != nil && IsNoRows(err) {
				user, err = q.CreateUser(ctx)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *NucleusDb) SetPersonalAccount(
	ctx context.Context,
	userId pgtype.UUID,
) (*db_queries.NeosyncApiAccount, error) {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Println("called rollback", err)
		}
	}()

	q := d.Q.WithTx(tx)
	account, err := q.GetPersonalAccountByUserId(ctx, userId)
	if err != nil && !IsNoRows(err) {
		return nil, err
	} else if err != nil && IsNoRows(err) {
		account, err = q.CreatePersonalAccount(ctx, "personal")
		if err != nil {
			return nil, err
		}
		_, err = q.CreateAccountUserAssociation(ctx, db_queries.CreateAccountUserAssociationParams{
			AccountID: account.ID,
			UserID:    userId,
		})
		if err != nil {
			return nil, err
		}
	} else {
		_, err = q.GetAccountUserAssociation(ctx, db_queries.GetAccountUserAssociationParams{
			AccountId: account.ID,
			UserId:    userId,
		})
		if err != nil && !IsNoRows(err) {
			return nil, err
		} else if err != nil && IsNoRows(err) {
			_, err = q.CreateAccountUserAssociation(ctx, db_queries.CreateAccountUserAssociationParams{
				AccountID: account.ID,
				UserID:    userId,
			})
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &account, nil
}
