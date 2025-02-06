package userdata

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nucleuscloud/neosync/internal/ee/rbac"
	"github.com/nucleuscloud/neosync/internal/neosyncdb"
)

// Domain entity interface that mimics the domain model of the mgmt service
type DomainEntity interface {
	Identifier
	GetAccountId() string
}

type DomainEntityImpl struct {
	id        string
	accountId string
	isWild    bool
}

type Identifier interface {
	GetId() string
}

func (j *DomainEntityImpl) GetId() string {
	return j.id
}
func (j *DomainEntityImpl) GetAccountId() string {
	return j.accountId
}

// Used for things like mgmtv1alpha1.Job, mgmtv1alpha1.Connection, etc
func NewDomainEntity(accountId, id string) DomainEntity {
	return &DomainEntityImpl{
		id:        id,
		accountId: accountId,
	}
}

// Used for things like mgmtv1alpha1.Job, mgmtv1alpha1.Connection, etc
// But for checking wildcard or account-level access
func NewWildcardDomainEntity(accountId string) DomainEntity {
	return &DomainEntityImpl{
		id:        rbac.Wildcard,
		accountId: accountId,
		isWild:    true,
	}
}

// Helper function that can be used when dealing with the DB entities instead of the domain entities
func NewDbDomainEntity(accountId, id pgtype.UUID) DomainEntity {
	return &DomainEntityImpl{
		id:        neosyncdb.UUIDString(id),
		accountId: neosyncdb.UUIDString(accountId),
	}
}

type IdentifierImpl struct {
	id string
}

// Helper function that creates just an identitier. Generally used when working with the account object
func NewIdentifier(id string) Identifier {
	return &IdentifierImpl{
		id: id,
	}
}

func (i *IdentifierImpl) GetId() string {
	return i.id
}
