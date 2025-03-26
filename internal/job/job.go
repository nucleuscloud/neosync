package job

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type SqlJobSourceOpts struct {
	// Determines if the job should halt if a new column is detected that is not present in the job mappings
	HaltOnNewColumnAddition bool
	// Determines if the job should halt if a column is removed from the source database
	HaltOnColumnRemoval bool
	// Newly detected columns are automatically transformed
	GenerateNewColumnTransformers bool
	SubsetByForeignKeyConstraints bool
	SchemaOpt                     []*SchemaOptions
}

type SchemaOptions struct {
	Schema string
	Tables []*TableOptions
}
type TableOptions struct {
	Table       string
	WhereClause *string
}

func GetSqlJobSourceOpts(
	source *mgmtv1alpha1.JobSource,
) (*SqlJobSourceOpts, error) {
	switch jobSourceConfig := source.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		if jobSourceConfig.Postgres == nil {
			return nil, nil
		}
		schemaOpt := []*SchemaOptions{}
		for _, opt := range jobSourceConfig.Postgres.Schemas {
			tableOpts := []*TableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &TableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &SchemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		shouldHalt := false
		shouldGenerateNewColTransforms := false
		switch jobSourceConfig.Postgres.GetNewColumnAdditionStrategy().GetStrategy().(type) {
		case *mgmtv1alpha1.PostgresSourceConnectionOptions_NewColumnAdditionStrategy_HaltJob_:
			shouldHalt = true
		case *mgmtv1alpha1.PostgresSourceConnectionOptions_NewColumnAdditionStrategy_AutoMap_:
			shouldGenerateNewColTransforms = true
		}

		shouldHaltOnColumnRemoval := false
		if jobSourceConfig.Postgres.GetColumnRemovalStrategy().GetHaltJob() != nil {
			shouldHaltOnColumnRemoval = true
		}

		return &SqlJobSourceOpts{
			HaltOnNewColumnAddition:       shouldHalt,
			HaltOnColumnRemoval:           shouldHaltOnColumnRemoval,
			GenerateNewColumnTransformers: shouldGenerateNewColTransforms,
			SubsetByForeignKeyConstraints: jobSourceConfig.Postgres.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		if jobSourceConfig.Mysql == nil {
			return nil, nil
		}
		schemaOpt := []*SchemaOptions{}
		for _, opt := range jobSourceConfig.Mysql.Schemas {
			tableOpts := []*TableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &TableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &SchemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		shouldHalt := false
		shouldGenerateNewColTransforms := false
		switch jobSourceConfig.Mysql.GetNewColumnAdditionStrategy().GetStrategy().(type) {
		case *mgmtv1alpha1.MysqlSourceConnectionOptions_NewColumnAdditionStrategy_HaltJob_:
			shouldHalt = true
		case *mgmtv1alpha1.MysqlSourceConnectionOptions_NewColumnAdditionStrategy_AutoMap_:
			shouldGenerateNewColTransforms = true
		}
		if !shouldHalt && jobSourceConfig.Mysql.GetHaltOnNewColumnAddition() {
			shouldHalt = true
		}
		shouldHaltOnColumnRemoval := false
		if jobSourceConfig.Mysql.GetColumnRemovalStrategy().GetHaltJob() != nil {
			shouldHaltOnColumnRemoval = true
		}
		return &SqlJobSourceOpts{
			HaltOnNewColumnAddition:       shouldHalt,
			HaltOnColumnRemoval:           shouldHaltOnColumnRemoval,
			GenerateNewColumnTransformers: shouldGenerateNewColTransforms,
			SubsetByForeignKeyConstraints: jobSourceConfig.Mysql.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		if jobSourceConfig.Mssql == nil {
			return nil, nil
		}
		schemaOpt := []*SchemaOptions{}
		for _, opt := range jobSourceConfig.Mssql.Schemas {
			tableOpts := []*TableOptions{}
			for _, t := range opt.GetTables() {
				tableOpts = append(tableOpts, &TableOptions{
					Table:       t.Table,
					WhereClause: t.WhereClause,
				})
			}
			schemaOpt = append(schemaOpt, &SchemaOptions{
				Schema: opt.GetSchema(),
				Tables: tableOpts,
			})
		}
		shouldHaltOnColumnRemoval := false
		if jobSourceConfig.Mssql.GetColumnRemovalStrategy().GetHaltJob() != nil {
			shouldHaltOnColumnRemoval = true
		}
		return &SqlJobSourceOpts{
			HaltOnNewColumnAddition:       jobSourceConfig.Mssql.HaltOnNewColumnAddition,
			HaltOnColumnRemoval:           shouldHaltOnColumnRemoval,
			SubsetByForeignKeyConstraints: jobSourceConfig.Mssql.SubsetByForeignKeyConstraints,
			SchemaOpt:                     schemaOpt,
		}, nil
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		if jobSourceConfig.Generate == nil {
			return nil, nil
		}
		return &SqlJobSourceOpts{}, nil
	default:
		return nil, fmt.Errorf("unsupported job source options type for sql job source: %T", jobSourceConfig)
	}
}
