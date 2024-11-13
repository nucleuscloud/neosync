package rbac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/casbin/casbin/v2"
)

type Rbac struct {
	e casbin.IEnforcer
}

func New(
	e casbin.IEnforcer,
) *Rbac {
	return &Rbac{e: e}
}

type Db interface {
	GetAccountIds(ctx context.Context) ([]string, error)
}

func (r *Rbac) InitPolicies(
	ctx context.Context,
	db Db,
	logger *slog.Logger,
) error {
	accountIds, err := db.GetAccountIds(ctx)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("found %d account ids to associate with rbac policies", len(accountIds)))

	policyRules := [][]string{}
	for _, accountId := range accountIds {
		accountKey := fmt.Sprintf("account/%s", accountId)

		policyRules = append(
			policyRules,
			[]string{
				"account_admin",
				accountKey,
				"*", // any resource in the account
				"*", // all actions in the account
			},
			[]string{
				"job_developer",
				accountKey,
				"jobs/*", // all jobs in the account
				"*",      // all job actions
			},
			[]string{
				"job_viewer",
				accountKey,
				"jobs/*",
				"view",
			},
			[]string{
				"job_viewer",
				accountKey,
				"jobs/*",
				"execute",
			},
		)
	}
	if len(policyRules) > 0 {
		logger.Debug(fmt.Sprintf("adding %d policy rules to rbac enginee", len(policyRules)))
		for _, policy := range policyRules {
			err := setPolicy(r.e, policy)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func setPolicy(e casbin.IEnforcer, policy []string) error {
	// AddPoliciesEx is what should be uesd here but is resulting in duplicates
	// This logic here seems to handle this instead
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
