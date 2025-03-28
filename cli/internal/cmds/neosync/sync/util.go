package sync_cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	benthosbuilder "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder"
	"github.com/nucleuscloud/neosync/internal/runconfigs"
)

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseDriverString(str string) (DriverType, bool) {
	p, ok := driverMap[strings.ToLower(str)]
	return p, ok
}

func isConfigReady(config *benthosbuilder.BenthosConfigResponse, queuedMap map[string][]string) bool {
	for _, dep := range config.DependsOn {
		if cols, ok := queuedMap[dep.Table]; ok {
			for _, dc := range dep.Columns {
				if !slices.Contains(cols, dc) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

func groupConfigsByDependency(configs []*benthosbuilder.BenthosConfigResponse, logger *slog.Logger) [][]*benthosbuilder.BenthosConfigResponse {
	groupedConfigs := [][]*benthosbuilder.BenthosConfigResponse{}
	configMap := map[string]*benthosbuilder.BenthosConfigResponse{}
	queuedMap := map[string][]string{} // map -> table to cols

	// get root configs
	rootConfigs := []*benthosbuilder.BenthosConfigResponse{}
	for _, c := range configs {
		if len(c.DependsOn) == 0 {
			table := fmt.Sprintf("%s.%s", c.TableSchema, c.TableName)
			rootConfigs = append(rootConfigs, c)
			queuedMap[table] = c.Columns
		} else {
			configMap[c.Name] = c
		}
	}
	if len(rootConfigs) == 0 {
		logger.Info("No root configs found. There must be one config with no dependencies.")
		return nil
	}
	groupedConfigs = append(groupedConfigs, rootConfigs)

	prevTableLen := 0
	for len(configMap) > 0 {
		// prevents looping forever
		if prevTableLen == len(configMap) {
			logger.Error("Unable to order configs by dependency. No path found.")
			return nil
		}
		prevTableLen = len(configMap)
		dependentConfigs := []*benthosbuilder.BenthosConfigResponse{}
		for _, c := range configMap {
			if isConfigReady(c, queuedMap) {
				dependentConfigs = append(dependentConfigs, c)
				delete(configMap, c.Name)
			}
		}
		if len(dependentConfigs) > 0 {
			groupedConfigs = append(groupedConfigs, dependentConfigs)
			for _, c := range dependentConfigs {
				table := fmt.Sprintf("%s.%s", c.TableSchema, c.TableName)
				queuedMap[table] = append(queuedMap[table], c.Columns...)
			}
		}
	}

	return groupedConfigs
}

func getTableColMap(schemas []*mgmtv1alpha1.DatabaseColumn) map[string][]string {
	tableColMap := map[string][]string{}
	for _, record := range schemas {
		table := sql_manager.BuildTable(record.Schema, record.Table)
		_, ok := tableColMap[table]
		if ok {
			tableColMap[table] = append(tableColMap[table], record.Column)
		} else {
			tableColMap[table] = []string{record.Column}
		}
	}

	return tableColMap
}

func buildDependencyMap(syncConfigs []*runconfigs.RunConfig) map[string][]string {
	dependencyMap := map[string][]string{}
	for _, cfg := range syncConfigs {
		_, dpOk := dependencyMap[cfg.Table()]
		if !dpOk {
			dependencyMap[cfg.Table()] = []string{}
		}

		for _, dep := range cfg.DependsOn() {
			dependencyMap[cfg.Table()] = append(dependencyMap[cfg.Table()], dep.Table)
		}
	}
	return dependencyMap
}

func areSourceAndDestCompatible(connection *mgmtv1alpha1.Connection, destinationDriver *DriverType) error {
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if destinationDriver != nil && *destinationDriver != postgresDriver {
			return fmt.Errorf("connection and destination types are incompatible [postgres, %s]", *destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if destinationDriver != nil && *destinationDriver != mysqlDriver {
			return fmt.Errorf("connection and destination types are incompatible [mysql, %s]", *destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config, *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig, *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
	default:
		return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
	}
	return nil
}
