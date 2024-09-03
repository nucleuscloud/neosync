package mssql_simple

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		// 	{
		// 		Name:            "Simple",
		// 		Folder:          "mssql/simple",
		// 		SourceFilePaths: []string{"create-schema-sales.sql", "create-schema-production.sql", "create-table.sql", "insert.sql"},
		// 		TargetFilePaths: []string{"create-schema-sales.sql", "create-schema-production.sql", "create-table.sql"},
		// 		JobMappings:     GetDefaultSyncJobMappings(),
		// 		JobOptions: &workflow_testdata.TestJobOptions{
		// 			InitSchema: true,
		// 		},
		// 		Expected: map[string]*workflow_testdata.ExpectedOutput{
		// 			"production.categories": &workflow_testdata.ExpectedOutput{RowCount: 7},
		// 			"production.brands":     &workflow_testdata.ExpectedOutput{RowCount: 9},
		// 			"production.products":   &workflow_testdata.ExpectedOutput{RowCount: 18},
		// 			"production.stocks":     &workflow_testdata.ExpectedOutput{RowCount: 32},
		// 			"sales.customers":       &workflow_testdata.ExpectedOutput{RowCount: 15},
		// 			"sales.stores":          &workflow_testdata.ExpectedOutput{RowCount: 3},
		// 			"sales.staffs":          &workflow_testdata.ExpectedOutput{RowCount: 10},
		// 			"sales.orders":          &workflow_testdata.ExpectedOutput{RowCount: 13},
		// 			"sales.order_items":     &workflow_testdata.ExpectedOutput{RowCount: 26},
		// 		},
		// 	},
		{
			Name:            "Identity Columns set to default",
			Folder:          "mssql/simple",
			SourceFilePaths: []string{"create-schema-sales.sql", "create-schema-production.sql", "create-table.sql", "insert.sql"},
			TargetFilePaths: []string{"create-schema-sales.sql", "create-schema-production.sql", "create-table.sql"},
			JobMappings:     getJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"production.categories": &workflow_testdata.ExpectedOutput{RowCount: 7},
				"production.brands":     &workflow_testdata.ExpectedOutput{RowCount: 9},
				"production.products":   &workflow_testdata.ExpectedOutput{RowCount: 18},
				"production.stocks":     &workflow_testdata.ExpectedOutput{RowCount: 32},
				"sales.customers":       &workflow_testdata.ExpectedOutput{RowCount: 15},
				"sales.stores":          &workflow_testdata.ExpectedOutput{RowCount: 3},
				"sales.staffs":          &workflow_testdata.ExpectedOutput{RowCount: 10},
				"sales.orders":          &workflow_testdata.ExpectedOutput{RowCount: 13},
				"sales.order_items":     &workflow_testdata.ExpectedOutput{RowCount: 26},
			},
		},
	}
}

func getJobmappings() []*mgmtv1alpha1.JobMapping {
	tableColTypeMap := GetTableColumnTypeMap()
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		colTypeMap, ok := tableColTypeMap[fmt.Sprintf("%s.%s", jm.Schema, jm.Table)]
		if ok {
			t, ok := colTypeMap[jm.Column]
			if ok && strings.HasPrefix(t, "INTIDENTITY") {
				updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
					Schema:      jm.Schema,
					Table:       jm.Table,
					Column:      jm.Column,
					Transformer: getDefaultTransformerConfig(),
				})
				continue
			}
		}
		updatedJobmappings = append(updatedJobmappings, jm)
	}
	return updatedJobmappings
}

func getDefaultTransformerConfig() *mgmtv1alpha1.JobMappingTransformer {
	return &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
				GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
			},
		},
	}
}
