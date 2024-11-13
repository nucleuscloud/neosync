-- name: GetUser :one
SELECT * FROM neosync_api.users
WHERE id = $1;

-- name: GetUserAssociationByProviderSub :one
SELECT * from neosync_api.user_identity_provider_associations
WHERE provider_sub = $1;

-- name: GetUserByProviderSub :one
SELECT u.* from neosync_api.users u
INNER JOIN neosync_api.user_identity_provider_associations uipa ON uipa.user_id = u.id
WHERE uipa.provider_sub = $1 and u.user_type = 0;

-- name: CreateNonMachineUser :one
INSERT INTO neosync_api.users (
  id, created_at, updated_at, user_type
) VALUES (
  DEFAULT, DEFAULT, DEFAULT, 0
)
RETURNING *;

-- name: CreateMachineUser :one
INSERT INTO neosync_api.users (
  id, created_at, updated_at, user_type
) VALUES (
  DEFAULT, DEFAULT, DEFAULT, 1
)
RETURNING *;

-- name: GetUserIdentitiesByTeamAccount :many
SELECT aipa.* FROM neosync_api.user_identity_provider_associations aipa
JOIN neosync_api.account_user_associations aua ON aua.user_id = aipa.user_id
JOIN neosync_api.accounts a ON a.id = aua.account_id
WHERE aua.account_id = sqlc.arg('accountId') AND a.account_type = 1;

-- name: GetUserIdentityByUserId :one
SELECT aipa.* FROM neosync_api.user_identity_provider_associations aipa
WHERE aipa.user_id = $1;

-- name: CreateIdentityProviderAssociation :one
INSERT INTO neosync_api.user_identity_provider_associations (
  user_id, provider_sub
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetUserIdentityAssociationsByUserIds :many
SELECT * from neosync_api.user_identity_provider_associations
WHERE user_id = ANY($1::uuid[]);

-- name: GetAccount :one
SELECT * from neosync_api.accounts
WHERE id = $1;

-- name: GetPersonalAccountByUserId :one
SELECT a.* from neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE u.id = sqlc.arg('userId') AND a.account_type = 0;

-- name: GetTeamAccountsByUserId :many
SELECT a.* from neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE u.id = sqlc.arg('userId') AND a.account_type = 1;

-- name: CreatePersonalAccount :one
INSERT INTO neosync_api.accounts (
  account_type, account_slug, max_allowed_records
) VALUES (
  0, $1, $2
)
RETURNING *;

-- name: CreateTeamAccount :one
INSERT INTO neosync_api.accounts (
  account_type, account_slug
) VALUES (
  1, $1
)
RETURNING *;

-- name: GetAccountsByUser :many
SELECT a.*
FROM neosync_api.accounts a
INNER JOIN neosync_api.account_api_keys aak ON aak.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aak.user_id
WHERE u.id = $1

UNION

SELECT a.*
FROM neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE u.id = $1;


-- name: CreateAccountUserAssociation :exec
INSERT INTO neosync_api.account_user_associations (
  account_id, user_id
) VALUES (
  $1, $2
)
ON CONFLICT (account_id, user_id) DO NOTHING;

-- name: GetAccountUserAssociation :one
SELECT aua.* from neosync_api.account_user_associations aua
INNER JOIN neosync_api.accounts a ON a.id = aua.account_id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE a.id = sqlc.arg('accountId') AND u.id = sqlc.arg('userId');

-- name: IsUserInAccount :one
SELECT count(aua.id) from neosync_api.account_user_associations aua
INNER JOIN neosync_api.accounts a ON a.id = aua.account_id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE a.id = sqlc.arg('accountId') AND u.id = sqlc.arg('userId');

-- name: GetAnonymousUser :one
SELECT * from neosync_api.users
WHERE id = '00000000-0000-0000-0000-000000000000';

-- name: SetAnonymousUser :one
INSERT INTO neosync_api.users (
  id, created_at, updated_at
) VALUES (
  '00000000-0000-0000-0000-000000000000', DEFAULT, DEFAULT
)
ON CONFLICT (id)
DO
  UPDATE SET updated_at = current_timestamp
RETURNING *;

-- name: GetTemporalConfigByAccount :one
SELECT temporal_config
FROM neosync_api.accounts
WHERE id = $1;

-- name: UpdateTemporalConfigByAccount :one
UPDATE neosync_api.accounts
SET temporal_config = $1
WHERE id = sqlc.arg('accountId')
RETURNING *;

-- name: GetTemporalConfigByUserAccount :one
SELECT a.temporal_config
FROM neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE a.id = sqlc.arg('accountId') AND u.id = sqlc.arg('userId');

-- name: RemoveAccountUser :exec
DELETE FROM neosync_api.account_user_associations
WHERE account_id = sqlc.arg('accountId') AND user_id = sqlc.arg('userId');

-- name: CreateAccountInvite :one
INSERT INTO neosync_api.account_invites (
  account_id, sender_user_id, email, expires_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetActiveAccountInvites :many
SELECT * FROM neosync_api.account_invites
WHERE account_id = sqlc.arg('accountId') AND expires_at > CURRENT_TIMESTAMP AND accepted = false;

-- name: UpdateActiveAccountInvitesToExpired :one
UPDATE neosync_api.account_invites
SET expires_at = CURRENT_TIMESTAMP
WHERE account_id = sqlc.arg('accountId') AND email = sqlc.arg('email') AND expires_at > CURRENT_TIMESTAMP
RETURNING *;

-- name: UpdateAccountInviteToAccepted :one
UPDATE neosync_api.account_invites
SET accepted = true
WHERE id = $1
RETURNING *;

-- name: GetAccountInvite :one
SELECT * FROM neosync_api.account_invites
WHERE id = $1;

-- name: GetAccountInviteByToken :one
SELECT * FROM neosync_api.account_invites
WHERE token = $1;

-- name: RemoveAccountInvite :exec
DELETE FROM neosync_api.account_invites
WHERE id = $1;

-- name: GetAccountOnboardingConfig :one
SELECT onboarding_config
FROM neosync_api.accounts
WHERE id = $1;


-- name: UpdateAccountOnboardingConfig :one
UPDATE neosync_api.accounts
SET onboarding_config = $1
WHERE id = sqlc.arg('accountId')
RETURNING *;

-- name: SetAccountCreatedAt :one
UPDATE neosync_api.accounts
SET created_at = $1
WHERE id = sqlc.arg('accountId')
RETURNING *;

-- name: SetNewAccountStripeCustomerId :one
UPDATE neosync_api.accounts
SET stripe_customer_id = $1
WHERE id = sqlc.arg('accountId') AND stripe_customer_id IS NULL
RETURNING *;

-- name: GetBilledAccounts :many
SELECT *
FROM neosync_api.accounts
WHERE stripe_customer_id IS NOT NULL AND (sqlc.arg('accountIds')::uuid[] = '{}' OR id = ANY(sqlc.arg('accountIds')::uuid[]));

-- name: ConvertPersonalAccountToTeam :one
UPDATE neosync_api.accounts
SET account_slug = sqlc.arg('teamName'),
    account_type = 1,
    max_allowed_records = NULL
WHERE id = sqlc.arg('accountId')
RETURNING *;
