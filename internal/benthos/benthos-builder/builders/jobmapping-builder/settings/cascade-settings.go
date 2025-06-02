package cascade_settings

import (
	"iter"
	"log/slog"
	"maps"
	"slices"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type CascadeSchemaSettings struct {
	config *mgmtv1alpha1.JobTypeConfig_JobTypeSync
	logger *slog.Logger
}

func NewCascadeSchemaSettings(config *mgmtv1alpha1.JobTypeConfig_JobTypeSync, logger *slog.Logger) *CascadeSchemaSettings {
	return &CascadeSchemaSettings{
		config: config,
		logger: logger,
	}
}

func (c *CascadeSchemaSettings) GetSchemaStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	return getSchemaStrategyOrDefault(
		c.logger,
		c.getGlobalSchemaStrategy(),
	)
}

func (c *CascadeSchemaSettings) getGlobalSchemaStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	return c.config.GetSchemaChange().GetSchemaStrategy()
}

func getSchemaStrategyOrDefault(
	logger *slog.Logger,
	strategies ...*mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy,
) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	var strategy *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy
	for _, s := range strategies {
		if s != nil {
			strategy = s
			break
		}
	}
	if strategy == nil {
		return getDefaultSchemaStrategy(logger)
	}

	switch s := strategy.GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_:
		if s.MapAllSchemas == nil {
			strategy.Strategy = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_{
				MapAllSchemas: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas{},
			}
		}
		return strategy
	default:
		return getDefaultSchemaStrategy(logger)
	}
}

func getDefaultSchemaStrategy(logger *slog.Logger) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	logger.Debug("no schema strategy defined, using default MapAllSchemas strategy")
	return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy{
		Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas_{
			MapAllSchemas: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy_MapAllSchemas{},
		},
	}
}

func (c *CascadeSchemaSettings) GetTableStrategy(schemaName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	return getTableStrategyOrDefault(
		c.logger,
		c.getSchemaLevelTableStrategy(schemaName),
		c.getGlobalTableStrategy(),
	)
}

func (c *CascadeSchemaSettings) getGlobalTableStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	return c.config.GetSchemaChange().GetTableStrategy()
}

func getTableStrategyOrDefault(
	logger *slog.Logger,
	strategies ...*mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy,
) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	var strategy *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy
	for _, s := range strategies {
		if s != nil {
			strategy = s
			break
		}
	}
	if strategy == nil {
		return getDefaultTableStrategy(logger)
	}

	switch s := strategy.GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_:
		if s.MapAllTables == nil {
			strategy.Strategy = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_{
				MapAllTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables{},
			}
		}
		return strategy
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_:
		if s.MapDefinedTables == nil {
			strategy.Strategy = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables_{
				MapDefinedTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapDefinedTables{},
			}
		}
		return strategy
	default:
		return getDefaultTableStrategy(logger)
	}
}

func getDefaultTableStrategy(logger *slog.Logger) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	logger.Debug("no table strategy defined, using default MapAllTables strategy")
	return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy{
		Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables_{
			MapAllTables: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy_MapAllTables{},
		},
	}
}

func (c *CascadeSchemaSettings) getSchemaLevelTableStrategy(schemaName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			return schemaMapping.GetTableStrategy()
		}
	}
	return nil
}

func (c *CascadeSchemaSettings) GetColumnStrategy(schemaName, tableName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	return getColumnStrategyOrDefault(
		c.getTablelevelColumnStrategy(schemaName, tableName),
		c.getSchemaLevelColumnStrategy(schemaName),
		c.getGlobalColumnStrategy(),
	)
}

func getColumnStrategyOrDefault(
	strategies ...*mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy,
) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	var strategy *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy
	for _, s := range strategies {
		if s != nil {
			strategy = s
			break
		}
	}
	if strategy == nil {
		return getDefaultColumnStrategy()
	}

	switch s := strategy.GetStrategy().(type) {
	case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_:
		mapAll := s.MapAllColumns
		if mapAll == nil {
			mapAll = getDefaultMapAllColumnsStrategy()
		}

		// Get default settings to merge with
		defaultStrat := getDefaultMapAllColumnsStrategy()

		// Merge with defaults for any undefined settings
		if mapAll.ColumnInSourceNotMapped == nil {
			mapAll.ColumnInSourceNotMapped = defaultStrat.ColumnInSourceNotMapped
		} else {
			switch start := mapAll.GetColumnInSourceNotMapped().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_AutoMap_:
				if start.AutoMap == nil {
					start.AutoMap = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_AutoMap{}
				}
				mapAll.ColumnInSourceNotMapped.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Drop_:
				if start.Drop == nil {
					start.Drop = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Drop{}
				}
				mapAll.ColumnInSourceNotMapped.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough_:
				if start.Passthrough == nil {
					start.Passthrough = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough{}
				}
				mapAll.ColumnInSourceNotMapped.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Halt_:
				if start.Halt == nil {
					start.Halt = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Halt{}
				}
				mapAll.ColumnInSourceNotMapped.Strategy = start
			default:
				mapAll.ColumnInSourceNotMapped = defaultStrat.ColumnInSourceNotMapped
			}
		}

		if mapAll.ColumnMappedNotInSource == nil {
			mapAll.ColumnMappedNotInSource = defaultStrat.ColumnMappedNotInSource
		} else {
			switch start := mapAll.GetColumnMappedNotInSource().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Continue_:
				if start.Continue == nil {
					start.Continue = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Continue{}
				}
				mapAll.ColumnMappedNotInSource.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Halt_:
				if start.Halt == nil {
					start.Halt = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Halt{}
				}
				mapAll.ColumnMappedNotInSource.Strategy = start
			default:
				mapAll.ColumnMappedNotInSource = defaultStrat.ColumnMappedNotInSource
			}
		}

		if mapAll.ColumnInSourceMappedNotInDestination == nil {
			mapAll.ColumnInSourceMappedNotInDestination = defaultStrat.ColumnInSourceMappedNotInDestination
		} else {
			switch start := mapAll.GetColumnInSourceMappedNotInDestination().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Drop_:
				if start.Drop == nil {
					start.Drop = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Drop{}
				}
				mapAll.ColumnInSourceMappedNotInDestination.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Halt_:
				if start.Halt == nil {
					start.Halt = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Halt{}
				}
				mapAll.ColumnInSourceMappedNotInDestination.Strategy = start
			default:
				mapAll.ColumnInSourceMappedNotInDestination = defaultStrat.ColumnInSourceMappedNotInDestination
			}
		}

		if mapAll.ColumnInDestinationNoLongerInSource == nil {
			mapAll.ColumnInDestinationNoLongerInSource = defaultStrat.ColumnInDestinationNoLongerInSource
		} else {
			switch start := mapAll.GetColumnInDestinationNoLongerInSource().GetStrategy().(type) {
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_AutoMap_:
				if start.AutoMap == nil {
					start.AutoMap = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_AutoMap{}
				}
				mapAll.ColumnInDestinationNoLongerInSource.Strategy = start
			case *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_Continue_:
				if start.Continue == nil {
					start.Continue = &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_Continue{}
				}
				mapAll.ColumnInDestinationNoLongerInSource.Strategy = start
			default:
				mapAll.ColumnInDestinationNoLongerInSource = defaultStrat.ColumnInDestinationNoLongerInSource
			}
		}

		return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
			Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
				MapAllColumns: mapAll,
			},
		}
	default:
		return getDefaultColumnStrategy()
	}
}

func (c *CascadeSchemaSettings) getTablelevelColumnStrategy(
	schemaName, tableName string,
) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			for _, tableMapping := range schemaMapping.GetTableMappings() {
				if tableMapping.GetTable() == tableName {
					return tableMapping.GetColumnStrategy()
				}
			}
			break
		}
	}
	return nil
}

func (c *CascadeSchemaSettings) getSchemaLevelColumnStrategy(schemaName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			if cs := schemaMapping.GetColumnStrategy(); cs != nil {
				return cs
			}
		}
	}
	return nil
}

func (c *CascadeSchemaSettings) getGlobalColumnStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	return c.config.GetSchemaChange().GetColumnStrategy()
}

func getDefaultColumnStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
		Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
			MapAllColumns: getDefaultMapAllColumnsStrategy(),
		},
	}
}

func getDefaultMapAllColumnsStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns {
	return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{
		ColumnInSourceNotMapped: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy{
			Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough_{
				Passthrough: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceNotMappedStrategy_Passthrough{},
			},
		},
		ColumnMappedNotInSource: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy{
			Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnMappedNotInSourceStrategy_Continue_{},
		},
		ColumnInSourceMappedNotInDestination: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy{
			Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInSourceMappedNotInDestinationStrategy_Drop_{},
		},
		ColumnInDestinationNoLongerInSource: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy{
			Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_ColumnInDestinationNotInSourceStrategy_AutoMap_{},
		},
	}
}

func (c *CascadeSchemaSettings) GetColumnTransformerConfigByTable(
	schemaName, tableName string,
) iter.Seq2[string, *mgmtv1alpha1.TransformerConfig] {
	return func(yield func(string, *mgmtv1alpha1.TransformerConfig) bool) {
		for _, schemaMapping := range c.config.GetSchemaMappings() {
			if schemaMapping.GetSchema() == schemaName {
				for _, tableMapping := range schemaMapping.GetTableMappings() {
					if tableMapping.GetTable() == tableName {
						for _, columnMapping := range tableMapping.GetColumnMappings() {
							if !yield(columnMapping.GetColumn(), columnMapping.GetTransformer()) {
								return
							}
						}
					}
				}
			}
		}
	}
}

func (c *CascadeSchemaSettings) GetDefinedSchemas() []string {
	output := map[string]bool{}
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		output[schemaMapping.GetSchema()] = true
	}
	return slices.Collect(maps.Keys(output))
}

func (c *CascadeSchemaSettings) GetDefinedTables(schemaName string) []string {
	output := map[string]bool{}
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			for _, tableMapping := range schemaMapping.GetTableMappings() {
				output[tableMapping.GetTable()] = true
			}
		}
	}
	return slices.Collect(maps.Keys(output))
}
