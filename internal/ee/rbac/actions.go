package rbac

type AccountAction string

const (
	AccountAction_Create AccountAction = "create"
	AccountAction_Delete AccountAction = "delete"
	AccountAction_View   AccountAction = "view"
	AccountAction_Edit   AccountAction = "edit"
)

func (a AccountAction) String() string {
	return string(a)
}

type ConnectionAction string

const (
	ConnectionAction_Create ConnectionAction = "create"
	ConnectionAction_Delete ConnectionAction = "delete"
	ConnectionAction_View   ConnectionAction = "view"
	ConnectionAction_Edit   ConnectionAction = "edit"
)

func (c ConnectionAction) String() string {
	return string(c)
}

type JobAction string

const (
	JobAction_Create  JobAction = "create"
	JobAction_Delete  JobAction = "delete"
	JobAction_Execute JobAction = "execute"
	JobAction_View    JobAction = "view"
	JobAction_Edit    JobAction = "edit"
)

func (a JobAction) String() string {
	return string(a)
}
