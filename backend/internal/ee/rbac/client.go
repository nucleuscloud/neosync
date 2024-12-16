package rbac

import "github.com/casbin/casbin/v2"

type Rbac struct {
	e casbin.IEnforcer
}

// Combines RBAC interface that handles entity enforcement and role management
type Interface interface {
	EntityEnforcer
	RoleAdmin
}

var _ Interface = (*Rbac)(nil)

func New(
	e casbin.IEnforcer,
) *Rbac {
	return &Rbac{e: e}
}
