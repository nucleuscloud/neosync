package cascade_settings

import (
	"iter"
	"maps"
	"slices"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type CascadeSchemaSettings struct {
	config *mgmtv1alpha1.JobTypeConfig_JobTypeSync
}

func NewCascadeSchemaSettings(config *mgmtv1alpha1.JobTypeConfig_JobTypeSync) *CascadeSchemaSettings {
	return &CascadeSchemaSettings{
		config: config,
	}
}

func (c *CascadeSchemaSettings) GetSchemaStrategy() *mgmtv1alpha1.JobTypeConfig_JobTypeSync_SchemaStrategy {
	return c.config.GetSchemaChange().GetSchemaStrategy()
}

func (c *CascadeSchemaSettings) GetTableStrategy(schemaName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_TableStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			ts := schemaMapping.GetTableStrategy()
			if ts != nil {
				return ts
			} else {
				break // fall back to global table strategy
			}
		}
	}
	return c.config.GetSchemaChange().GetTableStrategy()
}

func (c *CascadeSchemaSettings) GetColumnStrategy(schemaName, tableName string) *mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy {
	for _, schemaMapping := range c.config.GetSchemaMappings() {
		if schemaMapping.GetSchema() == schemaName {
			for _, tableMapping := range schemaMapping.GetTableMappings() {
				if tableMapping.GetTable() == tableName {
					tableLevelColumnStrategy := tableMapping.GetColumnStrategy()
					if tableLevelColumnStrategy != nil {
						return tableLevelColumnStrategy
					}
					break // fall back to schema level column strategy
				}
			}
			schemaLevelColumnStrategy := schemaMapping.GetColumnStrategy()
			if schemaLevelColumnStrategy != nil {
				return schemaLevelColumnStrategy
			}
			break // fall back to global column strategy
		}
	}
	globalStrat := c.config.GetSchemaChange().GetColumnStrategy()
	if globalStrat != nil {
		return globalStrat
	}
	return &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy{
		Strategy: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns_{
			MapAllColumns: &mgmtv1alpha1.JobTypeConfig_JobTypeSync_ColumnStrategy_MapAllColumns{
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
			},
		},
	}
}

func (c *CascadeSchemaSettings) GetColumnTransformerConfigByTable(schemaName, tableName string) iter.Seq2[string, *mgmtv1alpha1.TransformerConfig] {
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
