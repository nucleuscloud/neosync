package rbac

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Role string

const (
	Role_AccountAdmin Role = "account_admin"
	Role_JobDeveloper Role = "job_developer"
	Role_JobExecutor  Role = "job_executor"
	Role_JobViewer    Role = "job_viewer"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) ToDto() mgmtv1alpha1.AccountRole {
	switch r {
	case Role_AccountAdmin:
		return mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN
	case Role_JobDeveloper:
		return mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER
	case Role_JobExecutor:
		return mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_EXECUTOR
	case Role_JobViewer:
		return mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER
	default:
		return mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_UNSPECIFIED
	}
}

func fromRoleDto(role mgmtv1alpha1.AccountRole) (string, error) {
	switch role {
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN:
		return Role_AccountAdmin.String(), nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER:
		return Role_JobDeveloper.String(), nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_EXECUTOR:
		return Role_JobExecutor.String(), nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER:
		return Role_JobViewer.String(), nil
	default:
		return "", fmt.Errorf("account role provided has not be mapped to a casbin role name: %d", role)
	}
}
