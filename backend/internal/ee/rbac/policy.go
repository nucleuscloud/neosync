package rbac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type Rbac struct {
	e casbin.IEnforcer
}

func New(
	e casbin.IEnforcer,
) *Rbac {
	return &Rbac{e: e}
}

// Interface used by rbac engine to make necessary calls to the database
type Db interface {
	GetAccountIds(ctx context.Context) ([]string, error)
}

type EntityEnforcer interface {
	Job(ctx context.Context, user EntityString, account EntityString, job EntityString, action JobAction) (bool, error)
	Connection(ctx context.Context, user EntityString, account EntityString, connection EntityString, action ConnectionAction) (bool, error)
	Account(ctx context.Context, user EntityString, account EntityString, action AccountAction) (bool, error)
}

// Initialize default policies for existing accounts at startup
func (r *Rbac) InitPolicies(
	ctx context.Context,
	db Db,
	logger *slog.Logger,
) error {
	accountIds, err := db.GetAccountIds(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve account ids during casbin policy init: %w", err)
	}
	logger.Debug(fmt.Sprintf("found %d account ids to associate with rbac policies", len(accountIds)))

	policyRules := [][]string{}
	for _, accountId := range accountIds {
		accountKey := NewAccountIdEntity(accountId).String()

		policyRules = append(
			policyRules,
			[]string{
				"account_admin",
				accountKey,
				Wildcard, // any resource in the account
				Wildcard, // all actions in the account
			},
			[]string{
				"job_developer",
				accountKey,
				JobWildcard.String(), // all jobs in the account
				Wildcard,             // all job actions
			},
			[]string{
				"job_developer",
				accountKey,
				ConnectionWildcard.String(), // all connections in the account
				Wildcard,                    // all connection actions
			},
			[]string{
				"job_viewer",
				accountKey,
				JobWildcard.String(),
				"view",
			},
			[]string{
				"job_viewer",
				accountKey,
				JobWildcard.String(),
				"execute",
			},
		)
	}
	if len(policyRules) > 0 {
		logger.Debug(fmt.Sprintf("adding %d policy rules to rbac engine", len(policyRules)))
		for _, policy := range policyRules {
			err := setPolicy(r.e, policy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Rbac) SetAccountRole(
	ctx context.Context,
	user EntityString,
	account EntityString,
	role mgmtv1alpha1.AccountRole,
) error {
	roleName, err := toRoleName(role)
	if err != nil {
		return err
	}

	_, err = r.e.DeleteRolesForUserInDomain(user.String(), account.String())
	if err != nil {
		return fmt.Errorf("unable to delete roles for user in domain: %w", err)
	}

	_, err = r.e.AddRoleForUserInDomain(user.String(), roleName, account.String())
	return err
}

func (r *Rbac) RemoveAccountRole(
	ctx context.Context,
	user EntityString,
	account EntityString,
	role mgmtv1alpha1.AccountRole,
) error {
	roleName, err := toRoleName(role)
	if err != nil {
		return err
	}
	_, err = r.e.DeleteRoleForUserInDomain(user.String(), roleName, account.String())
	return err
}

func (r *Rbac) RemoveAccountUser(
	ctx context.Context,
	user EntityString,
	account EntityString,
) error {
	_, err := r.e.DeleteRolesForUserInDomain(user.String(), account.String())
	return err
}

func (r *Rbac) Job(
	ctx context.Context,
	user EntityString,
	account EntityString,
	job EntityString,
	action JobAction,
) (bool, error) {
	return r.e.Enforce(user.String(), account.String(), job.String(), action.String())
}

func (r *Rbac) Connection(
	ctx context.Context,
	user EntityString,
	account EntityString,
	connection EntityString,
	action ConnectionAction,
) (bool, error) {
	return r.e.Enforce(user.String(), account.String(), connection.String(), action.String())
}

func (r *Rbac) Account(
	ctx context.Context,
	user EntityString,
	account EntityString,
	action AccountAction,
) (bool, error) {
	return r.e.Enforce(user.String(), account.String(), account.String(), action.String())
}

func toRoleName(role mgmtv1alpha1.AccountRole) (string, error) {
	switch role {
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN:
		return "account_admin", nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER:
		return "job_developer", nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER:
		return "job_viewer", nil
	default:
		return "", fmt.Errorf("account role provided has not be mapped to a casbin role name: %d", role)
	}
}

func setPolicy(e casbin.IEnforcer, policy []string) error {
	// AddPoliciesEx is what should be uesd here but is resulting in duplicates (and errors with unique constraint)
	// AddPolicies handles the unique constraint but fails if even one policy already exists..

	// This logic here seems to handle what I want it to do instead strangely...
	ok, err := e.HasPolicy(policy)
	if err != nil {
		return fmt.Errorf("unable to check if policy exists: %w", err)
	}
	if !ok {
		_, err = e.AddPolicy(policy) // always resolves to true even if it was not added, may be adapter dependent
		if err != nil {
			return fmt.Errorf("unable to add policy: %w", err)
		}
	}
	return nil
}
