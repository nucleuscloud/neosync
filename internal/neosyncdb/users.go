package neosyncdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/internal/errors"
)

func (d *NeosyncDb) SetUserByAuthSub(
	ctx context.Context,
	authSub string,
) (*db_queries.NeosyncApiUser, error) {
	var userResp *db_queries.NeosyncApiUser
	if err := d.WithTx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(dbtx BaseDBTX) error {
		user, err := d.Q.GetUserByProviderSub(ctx, dbtx, authSub)
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			association, err := d.Q.GetUserAssociationByProviderSub(ctx, dbtx, authSub)
			if err != nil && !IsNoRows(err) {
				return err
			} else if err != nil && IsNoRows(err) {
				// create user, create association
				user, err = d.Q.CreateNonMachineUser(ctx, dbtx)
				if err != nil {
					return err
				}
				userResp = &user
				association, err = d.Q.CreateIdentityProviderAssociation(ctx, dbtx, db_queries.CreateIdentityProviderAssociationParams{
					UserID:      user.ID,
					ProviderSub: authSub,
				})
				if err != nil {
					return err
				}
			} else {
				user, err = d.Q.GetUser(ctx, dbtx, association.UserID)
				if err != nil && !IsNoRows(err) {
					return err
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

func (d *NeosyncDb) SetPersonalAccount(
	ctx context.Context,
	userId pgtype.UUID,
	maxAllowedRecords *int64, // only used when personal account is created
) (*db_queries.NeosyncApiAccount, error) {
	var personalAccount *db_queries.NeosyncApiAccount
	if err := d.WithTx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(dbtx BaseDBTX) error {
		resp, err := upsertPersonalAccount(ctx, d.Q, dbtx, &upsertPersonalAccountRequest{
			UserId:            userId,
			MaxAllowedRecords: maxAllowedRecords,
		})
		if err != nil {
			return err
		}
		personalAccount = resp.Account
		return nil
	}); err != nil {
		return nil, err
	}
	return personalAccount, nil
}

type upsertPersonalAccountRequest struct {
	UserId            pgtype.UUID
	MaxAllowedRecords *int64
}

type upsertPersonalAccountResponse struct {
	Account *db_queries.NeosyncApiAccount
}

func upsertPersonalAccount(ctx context.Context, q db_queries.Querier, dbtx BaseDBTX, req *upsertPersonalAccountRequest) (*upsertPersonalAccountResponse, error) {
	resp := &upsertPersonalAccountResponse{}
	account, err := q.GetPersonalAccountByUserId(ctx, dbtx, req.UserId)
	if err != nil && !IsNoRows(err) {
		return nil, err
	} else if err != nil && IsNoRows(err) {
		pgMaxAllowedRecords, err := int64ToPgInt8(req.MaxAllowedRecords)
		if err != nil {
			return nil, err
		}
		account, err = q.CreatePersonalAccount(ctx, dbtx, db_queries.CreatePersonalAccountParams{AccountSlug: "personal", MaxAllowedRecords: pgMaxAllowedRecords})
		if err != nil {
			return nil, err
		}
		resp.Account = &account
	} else {
		resp.Account = &account
	}
	err = q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
		AccountID: account.ID,
		UserID:    req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func int64ToPgInt8(val *int64) (pgtype.Int8, error) {
	output := pgtype.Int8{}
	if val != nil && *val > 0 {
		err := output.Scan(*val)
		if err != nil {
			return pgtype.Int8{}, fmt.Errorf("value was not scannable to pgtype.Int8: %w", err)
		}
	}
	return output, nil
}

func (d *NeosyncDb) CreateTeamAccount(
	ctx context.Context,
	userId pgtype.UUID,
	teamName string,
	logger *slog.Logger,
) (*db_queries.NeosyncApiAccount, error) {
	var teamAccount *db_queries.NeosyncApiAccount
	if err := d.WithTx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(dbtx BaseDBTX) error {
		accounts, err := d.Q.GetAccountsByUser(ctx, dbtx, userId)
		if err != nil && !IsNoRows(err) {
			return fmt.Errorf("unable to get account(s) by user id: %w", err)
		} else if err != nil && IsNoRows(err) {
			accounts = []db_queries.NeosyncApiAccount{}
		}
		logger.Debug(fmt.Sprintf("found %d accounts for user during team account creation", len(accounts)))
		if err := verifyAccountNameUnique(accounts, teamName); err != nil {
			return err
		}

		account, err := d.Q.CreateTeamAccount(ctx, dbtx, teamName)
		if err != nil {
			return fmt.Errorf("unable to create team account: %w", err)
		}
		teamAccount = &account
		err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
			AccountID: account.ID,
			UserID:    userId,
		})
		if err != nil {
			return fmt.Errorf("unable to associate user to newly created team account: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return teamAccount, nil
}

func verifyAccountNameUnique(accounts []db_queries.NeosyncApiAccount, name string) error {
	for idx := range accounts {
		if strings.EqualFold(accounts[idx].AccountSlug, name) {
			return nucleuserrors.NewAlreadyExists(fmt.Sprintf("team account with the name %s already exists", name))
		}
	}
	return nil
}

func getAccountById(accounts []db_queries.NeosyncApiAccount, id pgtype.UUID) (*db_queries.NeosyncApiAccount, error) {
	for idx := range accounts {
		if accounts[idx].ID.Valid && id.Valid && UUIDString(accounts[idx].ID) == UUIDString(id) {
			return &accounts[idx], nil
		}
	}
	return nil, nucleuserrors.NewNotFound("could not find id in list of neosync accounts")
}

type ConvertPersonalToTeamAccountRequest struct {
	// The id of the user
	UserId pgtype.UUID
	// The personal account id that will be converted to a team account
	PersonalAccountId pgtype.UUID
	// The name that the account slug will be updated to
	TeamName string
}

type ConvertPersonalToTeamAccountResponse struct {
	PersonalAccount *db_queries.NeosyncApiAccount
	TeamAccount     *db_queries.NeosyncApiAccount
}

func (d *NeosyncDb) ConvertPersonalToTeamAccount(
	ctx context.Context,
	req *ConvertPersonalToTeamAccountRequest,
	logger *slog.Logger,
) (*ConvertPersonalToTeamAccountResponse, error) {
	var resp *ConvertPersonalToTeamAccountResponse
	if err := d.WithTx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(dbtx BaseDBTX) error {
		accounts, err := d.Q.GetAccountsByUser(ctx, dbtx, req.UserId)
		if err != nil {
			return err
		}
		logger.DebugContext(ctx, fmt.Sprintf("found %d accounts for user during personal account conversion", len(accounts)))
		if err := verifyAccountNameUnique(accounts, req.TeamName); err != nil {
			return err
		}
		logger.DebugContext(ctx, "verified that new team name is unique within the users current account list")
		personalAccount, err := getAccountById(accounts, req.PersonalAccountId)
		if err != nil {
			return err
		}
		logger.DebugContext(ctx, "verified that requested personal account id is owned by the user")
		if personalAccount.AccountType != int16(AccountType_Personal) {
			return nucleuserrors.NewBadRequest("requested account conversion is not a personal account and thus cannot be converted")
		}

		// update personal account to be team account.
		// set max allowed records to nil, update slug, update account type
		updatedAccount, err := d.Q.ConvertPersonalAccountToTeam(ctx, dbtx, db_queries.ConvertPersonalAccountToTeamParams{
			TeamName:  req.TeamName,
			AccountId: personalAccount.ID,
		})
		if err != nil {
			return err
		}

		// create new personal account
		newPersonalAccountResp, err := upsertPersonalAccount(ctx, d.Q, dbtx, &upsertPersonalAccountRequest{
			UserId:            req.UserId,
			MaxAllowedRecords: &personalAccount.MaxAllowedRecords.Int64,
		})
		if err != nil {
			return err
		}
		resp = &ConvertPersonalToTeamAccountResponse{
			PersonalAccount: newPersonalAccountResp.Account,
			TeamAccount:     &updatedAccount,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (d *NeosyncDb) UpsertStripeCustomerId(
	ctx context.Context,
	accountId pgtype.UUID,
	getStripeCustomerId func(ctx context.Context, account db_queries.NeosyncApiAccount) (string, error),
	logger *slog.Logger,
) (*db_queries.NeosyncApiAccount, error) {
	var account *db_queries.NeosyncApiAccount

	// Serializable here to ensure the highest level of data integrity and avoid race conditions
	if err := d.WithTx(ctx, &pgx.TxOptions{IsoLevel: pgx.Serializable}, func(dbtx BaseDBTX) error {
		acc, err := d.Q.GetAccount(ctx, dbtx, accountId)
		if err != nil {
			return err
		}
		if acc.AccountType == int16(AccountType_Personal) {
			return errors.New("unsupported account type may not associate a stripe customer id")
		}

		if acc.StripeCustomerID.Valid {
			account = &acc
			logger.Debug("during stripe customer id upsert, found valid stripe customer id")
			return nil
		}
		customerId, err := getStripeCustomerId(ctx, acc)
		if err != nil {
			return fmt.Errorf("unable to get stripe customer id: %w", err)
		}
		logger.Debug("created new stripe customer id")
		updatedAcc, err := d.Q.SetNewAccountStripeCustomerId(ctx, dbtx, db_queries.SetNewAccountStripeCustomerIdParams{
			StripeCustomerID: pgtype.Text{String: customerId, Valid: true},
			AccountId:        accountId,
		})
		if err != nil {
			return fmt.Errorf("unable to update account with stripe customer id: %w", err)
		}
		// this shouldn't happen unless the write query is bad
		if !updatedAcc.StripeCustomerID.Valid {
			return errors.New("tried to update account with stripe customer id but received invalid response")
		}

		if updatedAcc.StripeCustomerID.String != customerId {
			logger.Warn("orphaned stripe customer id was created", "accountId", accountId)
		}
		account = &updatedAcc
		return nil
	}); err != nil {
		return nil, err
	}

	return account, nil
}

func (d *NeosyncDb) CreateTeamAccountInvite(
	ctx context.Context,
	accountId pgtype.UUID,
	userId pgtype.UUID,
	email string,
	expiresAt pgtype.Timestamp,
	role pgtype.Int4,
) (*db_queries.NeosyncApiAccountInvite, error) {
	var accountInvite *db_queries.NeosyncApiAccountInvite
	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		account, err := d.Q.GetAccount(ctx, dbtx, accountId)
		if err != nil {
			return err
		}
		if account.AccountType != int16(AccountType_Team) &&
			account.AccountType != int16(AccountType_Enterprise) {
			return nucleuserrors.NewForbidden("unable to create team account invite: account type is not team, enterprise")
		}

		// update any active invites for user to expired before creating new invite
		_, err = d.Q.UpdateActiveAccountInvitesToExpired(ctx, dbtx, db_queries.UpdateActiveAccountInvitesToExpiredParams{
			AccountId: accountId,
			Email:     email,
		})
		if err != nil && !IsNoRows(err) {
			return err
		}

		invite, err := d.Q.CreateAccountInvite(ctx, dbtx, db_queries.CreateAccountInviteParams{
			AccountID:    accountId,
			SenderUserID: userId,
			Email:        email,
			ExpiresAt:    expiresAt,
			Role:         role,
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

type ValidateInviteAddUserToAccountResponse struct {
	AccountId pgtype.UUID
	Role      mgmtv1alpha1.AccountRole
}

func (d *NeosyncDb) ValidateInviteAddUserToAccount(
	ctx context.Context,
	userId pgtype.UUID,
	token string,
	userEmail string,
) (*ValidateInviteAddUserToAccountResponse, error) {
	resp := &ValidateInviteAddUserToAccountResponse{}

	if err := d.WithTx(ctx, nil, func(dbtx BaseDBTX) error {
		invite, err := d.Q.GetAccountInviteByToken(ctx, dbtx, token)
		if err != nil && !IsNoRows(err) {
			return nucleuserrors.New(err)
		} else if err != nil && IsNoRows(err) {
			return nucleuserrors.NewBadRequest("invalid invite. unable to accept invite")
		}
		if invite.Email != userEmail {
			return nucleuserrors.NewBadRequest("invalid invite email. unable to accept invite")
		}
		if !invite.Accepted.Bool {
			_, err = d.Q.UpdateAccountInviteToAccepted(ctx, dbtx, invite.ID)
			if err != nil {
				return err
			}
		}
		resp.AccountId = invite.AccountID
		if invite.Role.Valid {
			resp.Role = mgmtv1alpha1.AccountRole(invite.Role.Int32)
		} else {
			resp.Role = mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER
		}

		_, err = d.Q.GetAccountUserAssociation(ctx, dbtx, db_queries.GetAccountUserAssociationParams{
			AccountId: invite.AccountID,
			UserId:    userId,
		})
		if err != nil && !IsNoRows(err) {
			return err
		} else if err != nil && IsNoRows(err) {
			if invite.Accepted.Bool {
				return nucleuserrors.NewBadRequest("account invitation already accepted")
			}

			if invite.ExpiresAt.Time.Before(time.Now().UTC()) {
				return nucleuserrors.NewForbidden("account invitation expired")
			}

			err = d.Q.CreateAccountUserAssociation(ctx, dbtx, db_queries.CreateAccountUserAssociationParams{
				AccountID: invite.AccountID,
				UserID:    userId,
			})
			if err != nil {
				return err
			}
		} else {
			return nil
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}
