package neosyncdb

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
)

type CreateAccountApiKeyRequest struct {
	KeyName           string
	KeyValue          string
	AccountUuid       pgtype.UUID
	CreatedByUserUuid pgtype.UUID
	ExpiresAt         pgtype.Timestamp
}

func (d *NeosyncDb) CreateAccountApikey(
	ctx context.Context,
	req *CreateAccountApiKeyRequest,
) (*db_queries.NeosyncApiAccountApiKey, error) {
	var createdApiKey *db_queries.NeosyncApiAccountApiKey
	if err := d.WithTx(ctx, nil, func(tx BaseDBTX) error {
		// create machine user
		user, err := d.Q.CreateMachineUser(ctx, tx)
		if err != nil {
			return err
		}
		newApiKey, err := d.Q.CreateAccountApiKey(
			ctx,
			tx,
			db_queries.CreateAccountApiKeyParams{
				KeyName:     req.KeyName,
				KeyValue:    req.KeyValue,
				AccountID:   req.AccountUuid,
				ExpiresAt:   req.ExpiresAt,
				CreatedByID: req.CreatedByUserUuid,
				UpdatedByID: req.CreatedByUserUuid,
				UserID:      user.ID,
			},
		)
		if err != nil {
			return err
		}
		createdApiKey = &newApiKey
		return nil
	}); err != nil {
		return nil, err
	}

	return createdApiKey, nil
}
