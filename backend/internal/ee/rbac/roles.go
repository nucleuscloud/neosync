package rbac

type Role string

const (
	Role_AccountAdmin Role = "account_admin"
	Role_JobDeveloper Role = "job_developer"
	Role_JobViewer    Role = "job_viewer"
)

func (r Role) String() string {
	return string(r)
}
