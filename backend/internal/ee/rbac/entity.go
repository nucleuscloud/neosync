package rbac

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
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

func NewJobIdEntity(value string) *Entity {
	return NewEntity("jobs", value)
}

func NewUserIdEntity(value string) *Entity {
	return NewEntity("users", value)
}
func NewPgUserIdEntity(value pgtype.UUID) *Entity {
	return NewUserIdEntity(neosyncdb.UUIDString(value))
}

func NewConnectionIdEntity(value string) *Entity {
	return NewEntity("connections", value)
}
