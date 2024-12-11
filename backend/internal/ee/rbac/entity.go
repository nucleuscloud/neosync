package rbac

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	db_queries "github.com/nucleuscloud/neosync/backend/gen/go/db"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

const (
	Wildcard = "*"
)

var (
	JobWildcard        = NewEntity("jobs", Wildcard)
	ConnectionWildcard = NewEntity("connections", Wildcard)
)

type Entity struct {
	prefix string
	value  string
}

type EntityString interface {
	String() string
}

func NewEntity(prefix, value string) *Entity {
	return &Entity{prefix: prefix, value: value}
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s/%s", e.prefix, e.value)
}

func NewAccountIdEntity(value string) *Entity {
	return NewEntity("accounts", value)
}
func NewPgAccountIdEntity(value pgtype.UUID) *Entity {
	return NewAccountIdEntity(neosyncdb.UUIDString(value))
}

func NewPgAccountEntity(account *db_queries.NeosyncApiAccount) *Entity {
	return NewPgAccountIdEntity(account.ID)
}
func NewDtoAccountEntity(account *mgmtv1alpha1.UserAccount) *Entity {
	return NewAccountIdEntity(account.GetId())
}

func NewJobIdEntity(value string) *Entity {
	return NewEntity("jobs", value)
}
func NewPgJobIdEntity(value pgtype.UUID) *Entity {
	return NewJobIdEntity(neosyncdb.UUIDString(value))
}

func NewDtoJobEntity(job *mgmtv1alpha1.Job) *Entity {
	return NewJobIdEntity(job.GetId())
}
func NewPgJobEntity(job *db_queries.NeosyncApiJob) *Entity {
	return NewPgJobIdEntity(job.ID)
}

func NewUserIdEntity(value string) *Entity {
	return NewEntity("users", value)
}
func NewPgUserIdEntity(value pgtype.UUID) *Entity {
	return NewUserIdEntity(neosyncdb.UUIDString(value))
}

func NewPgUserEntity(user *db_queries.NeosyncApiUser) *Entity {
	return NewPgUserIdEntity(user.ID)
}

func NewConnectionIdEntity(value string) *Entity {
	return NewEntity("connections", value)
}
func NewPgConnectionIdEntity(value pgtype.UUID) *Entity {
	return NewConnectionIdEntity(neosyncdb.UUIDString(value))
}

func NewDtoConnectionEntity(connection *mgmtv1alpha1.Connection) *Entity {
	return NewConnectionIdEntity(connection.GetId())
}
func NewPgConnectionEntity(connection *db_queries.NeosyncApiConnection) *Entity {
	return NewPgConnectionIdEntity(connection.ID)
}
