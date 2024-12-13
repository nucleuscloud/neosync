package rbac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	nucleuserrors "github.com/nucleuscloud/neosync/backend/internal/errors"
	"github.com/nucleuscloud/neosync/backend/internal/neosyncdb"
)

// Interface used by rbac engine to make necessary calls to the database
type Db interface {
	GetAccountIds(ctx context.Context) ([]string, error)
	GetAccountUsers(ctx context.Context, accountId string) ([]string, error)
}

// Interface that handles enforcing entity level policies
type EntityEnforcer interface {
	Job(ctx context.Context, user EntityString, account EntityString, job EntityString, action JobAction) (bool, error)
	EnforceJob(ctx context.Context, user EntityString, account EntityString, job EntityString, action JobAction) error
	Connection(ctx context.Context, user EntityString, account EntityString, connection EntityString, action ConnectionAction) (bool, error)
	EnforceConnection(ctx context.Context, user EntityString, account EntityString, connection EntityString, action ConnectionAction) error
	Account(ctx context.Context, user EntityString, account EntityString, action AccountAction) (bool, error)
	EnforceAccount(ctx context.Context, user EntityString, account EntityString, action AccountAction) error
}

// Interface that handles setting and removing roles for users
type RoleAdmin interface {
	SetAccountRole(ctx context.Context, user EntityString, account EntityString, role mgmtv1alpha1.AccountRole) error
	RemoveAccountRole(ctx context.Context, user EntityString, account EntityString, role mgmtv1alpha1.AccountRole) error
	RemoveAccountUser(ctx context.Context, user EntityString, account EntityString) error
	SetupNewAccount(ctx context.Context, accountId string, logger *slog.Logger) error
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
	err = setupAccountPolicies(r.e, accountIds, logger)
	if err != nil {
		return err
	}

	err = setupUserAssignments(ctx, db, r.e, accountIds, logger)
	if err != nil {
		return err
	}

	return nil
}

func setupAccountPolicies(enforcer casbin.IEnforcer, accountIds []string, logger *slog.Logger) error {
	logger.Debug(fmt.Sprintf("found %d account ids to associate with rbac policies", len(accountIds)))

	policyRules := [][]string{}
	for _, accountId := range accountIds {
		accountRules := getAccountPolicyRules(accountId)
		policyRules = append(
			policyRules,
			accountRules...,
		)
	}

	if len(policyRules) > 0 {
		logger.Debug(fmt.Sprintf("adding %d policy rules to rbac engine", len(policyRules)))
		for _, policy := range policyRules {
			err := setPolicy(enforcer, policy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// For the given accounts, assign users to the account admin role if the account does not currently have any role assignments
func setupUserAssignments(ctx context.Context, db Db, enforcer casbin.IEnforcer, accountIds []string, logger *slog.Logger) error {
	policiesByDomain, err := getGroupingPoliciesByDomain(enforcer)
	if err != nil {
		return err
	}

	groupedRules := [][]string{}
	for _, accountId := range accountIds {
		_, ok := policiesByDomain[NewAccountIdEntity(accountId).String()]
		if ok {
			continue
		}

		// get users in account
		// assign them all account admin role for the account
		users, err := db.GetAccountUsers(ctx, accountId)
		if err != nil && !neosyncdb.IsNoRows(err) {
			return err
		} else if err != nil && neosyncdb.IsNoRows(err) {
			logger.Debug(fmt.Sprintf("no users found for account %s, skipping", accountId))
			continue
		}
		logger.Debug(fmt.Sprintf("found %d users for account %s, assigning all account admin role", len(users), accountId))
		for _, user := range users {
			groupedRules = append(groupedRules, []string{
				NewUserIdEntity(user).String(),
				Role_AccountAdmin.String(),
				NewAccountIdEntity(accountId).String(),
			})
		}
	}
	if len(groupedRules) > 0 {
		logger.Debug(fmt.Sprintf("adding %d grouping policies to rbac engine", len(groupedRules)))
		_, err := enforcer.AddNamedGroupingPolicies("g", groupedRules)
		if err != nil {
			return err
		}
	}
	return nil
}

func getGroupingPoliciesByDomain(enforcer casbin.IEnforcer) (map[string][][]string, error) {
	// Get all grouping policies
	allPolicies, err := enforcer.GetNamedGroupingPolicy("g")
	if err != nil {
		return nil, fmt.Errorf("unable to get grouping policies: %w", err)
	}

	// Create a map to store policies by domain
	policiesByDomain := make(map[string][][]string)

	// Group policies by domain (domain is the third element, index 2)
	for _, policy := range allPolicies {
		domain := policy[2]
		policiesByDomain[domain] = append(policiesByDomain[domain], policy)
	}

	return policiesByDomain, nil
}

func (r *Rbac) SetupNewAccount(
	ctx context.Context,
	accountId string,
	logger *slog.Logger,
) error {
	accountRules := getAccountPolicyRules(accountId)
	if len(accountRules) > 0 {
		logger.Debug(fmt.Sprintf("adding %d policy rules to rbac engine for account %s", len(accountRules), accountId))
		for _, policy := range accountRules {
			err := setPolicy(r.e, policy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getAccountPolicyRules(accountId string) [][]string {
	accountKey := NewAccountIdEntity(accountId).String()
	return [][]string{
		{
			Role_AccountAdmin.String(),
			accountKey,
			Wildcard, // any resource in the account
			Wildcard, // all actions in the account
		},
		{
			Role_JobDeveloper.String(),
			accountKey,
			JobWildcard.String(), // all jobs in the account
			Wildcard,             // all job actions
		},
		{
			Role_JobDeveloper.String(),
			accountKey,
			ConnectionWildcard.String(), // all connections in the account
			Wildcard,                    // all connection actions
		},
		{
			Role_JobDeveloper.String(),
			accountKey,
			accountKey,
			AccountAction_View.String(), // job developer can view the account
		},
		{
			Role_JobViewer.String(),
			accountKey,
			JobWildcard.String(),
			JobAction_View.String(), // job viewer can view all jobs in the account
		},
		{
			Role_JobViewer.String(),
			accountKey,
			JobWildcard.String(),
			JobAction_Execute.String(), // job viewer can execute all jobs in the account
		},
		{
			Role_JobViewer.String(),
			accountKey,
			ConnectionWildcard.String(),
			ConnectionAction_View.String(), // job viewer can view all connections in the account
		},
		{
			Role_JobViewer.String(),
			accountKey,
			accountKey,
			AccountAction_View.String(), // job viewer can view the account
		},
	}
}

// For the given user and account, removes all existing roles and replaces them with the new role
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

// For the given user and account, removes the given role
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

// For the given user and account, removes all roles for the user
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

func (r *Rbac) EnforceJob(
	ctx context.Context,
	user EntityString,
	account EntityString,
	job EntityString,
	action JobAction,
) error {
	ok, err := r.Job(ctx, user, account, job, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s job", action))
	}
	return nil
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

func (r *Rbac) EnforceConnection(
	ctx context.Context,
	user EntityString,
	account EntityString,
	connection EntityString,
	action ConnectionAction,
) error {
	ok, err := r.Connection(ctx, user, account, connection, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s connection", action))
	}
	return nil
}

func (r *Rbac) Account(
	ctx context.Context,
	user EntityString,
	account EntityString,
	action AccountAction,
) (bool, error) {
	return r.e.Enforce(user.String(), account.String(), account.String(), action.String())
}

func (r *Rbac) EnforceAccount(
	ctx context.Context,
	user EntityString,
	account EntityString,
	action AccountAction,
) error {
	ok, err := r.Account(ctx, user, account, action)
	if err != nil {
		return err
	}
	if !ok {
		return nucleuserrors.NewForbidden(fmt.Sprintf("user does not have permission to %s account", action))
	}
	return nil
}

func toRoleName(role mgmtv1alpha1.AccountRole) (string, error) {
	switch role {
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_ADMIN:
		return Role_AccountAdmin.String(), nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_DEVELOPER:
		return Role_JobDeveloper.String(), nil
	case mgmtv1alpha1.AccountRole_ACCOUNT_ROLE_JOB_VIEWER:
		return Role_JobViewer.String(), nil
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
