package selectquerybuilder

import (
	"testing"

	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
	"github.com/stretchr/testify/assert"
)

func Test_BuildQuery_MySQLColumnQualification(t *testing.T) {
	t.Run("mysql column on left", func(t *testing.T) {
		assert.Equal(t,
			"SELECT `orders`.`order_id`, `orders`.`user_id` FROM `public`.`orders` AS `orders` INNER JOIN `public`.`users` AS `t_09a0eed1cbbe07ca` ON (`t_09a0eed1cbbe07ca`.`user_id` = `orders`.`user_id`) WHERE t_09a0eed1cbbe07ca.user_id < 100 ORDER BY `orders`.`order_id` ASC",
			buildOrdersUsersSubsettingQuery(t, "user_id < 100", sqlmanager_shared.MysqlDriver),
		)
	})

	t.Run("mysql column on right", func(t *testing.T) {
		assert.Equal(t,
			"SELECT `orders`.`order_id`, `orders`.`user_id` FROM `public`.`orders` AS `orders` INNER JOIN `public`.`users` AS `t_09a0eed1cbbe07ca` ON (`t_09a0eed1cbbe07ca`.`user_id` = `orders`.`user_id`) WHERE 100 > t_09a0eed1cbbe07ca.user_id ORDER BY `orders`.`order_id` ASC",
			buildOrdersUsersSubsettingQuery(t, "100 > user_id", sqlmanager_shared.MysqlDriver),
		)
	})
}

func buildOrdersUsersSubsettingQuery(t *testing.T, whereClause, driver string) string {
	t.Helper()

	runConfigs, err := runconfigs.BuildRunConfigs(
		map[string][]*sqlmanager_shared.ForeignConstraint{
			"public.orders": {{
				Columns:     []string{"user_id"},
				NotNullable: []bool{true},
				ForeignKey: &sqlmanager_shared.ForeignKey{
					Table:   "public.users",
					Columns: []string{"user_id"},
				},
			}},
		},
		map[string]string{"public.users": whereClause},
		map[string][]string{
			"public.orders": {"order_id"},
			"public.users":  {"user_id"},
		},
		map[string][]string{
			"public.orders": {"order_id", "user_id"},
			"public.users":  {"user_id"},
		},
		map[string][][]string{},
		map[string][][]string{},
	)
	assert.NoError(t, err)

	var ordersConfig *runconfigs.RunConfig

	for _, rc := range runConfigs {
		if rc.Table() == "public.orders" {
			ordersConfig = rc

			break
		}
	}

	assert.NotNil(t, ordersConfig)

	query, _, _, _, err := NewSelectQueryBuilder("public", driver, true, 0).BuildQuery(ordersConfig)
	assert.NoError(t, err)

	return query
}
