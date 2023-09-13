-- name: GetUser :one
SELECT * FROM neosync_api.users
WHERE id = $1;

-- name: GetUserAssociationByAuth0Id :one
SELECT * from neosync_api.user_identity_provider_associations
WHERE auth0_provider_id = $1;

-- name: GetUserByAuth0Id :one
SELECT u.* from neosync_api.users u
INNER JOIN neosync_api.user_identity_provider_associations uipa ON uipa.user_id = u.id
WHERE uipa.auth0_provider_id = $1;

-- name: CreateUser :one
INSERT INTO neosync_api.users (
  id, created_at, updated_at
) VALUES (
  DEFAULT, DEFAULT, DEFAULT
)
RETURNING *;

-- name: CreateAuth0IdentityProviderAssociation :one
INSERT INTO neosync_api.user_identity_provider_associations (
  user_id, auth0_provider_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetAccount :one
SELECT * from neosync_api.accounts
WHERE id = $1;

-- name: GetPersonalAccountByUserId :one
SELECT a.* from neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE u.id = sqlc.arg('userId') AND a.account_type = 0;

-- name: CreatePersonalAccount :one
INSERT INTO neosync_api.accounts (
  account_type, account_slug
) VALUES (
  0, $1
)
RETURNING *;

-- name: GetAccountsByUser :many
SELECT a.* from neosync_api.accounts a
INNER JOIN neosync_api.account_user_associations aua ON aua.account_id = a.id
INNER JOIN neosync_api.users u ON u.id = aua.user_id
WHERE u.id = $1;

-- name: CreateAccountUserAssociation :one
INSERT INTO neosync_api.account_user_associations (
  account_id, user_id
) VALUES (
  $1, $2
)
RETURNING *;

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
