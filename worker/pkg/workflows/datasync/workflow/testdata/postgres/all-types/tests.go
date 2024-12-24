package postgres_alltypes

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "All Postgres types",
			Folder:          "testdata/postgres/all-types",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql", "setup.sql"},
			JobMappings:     getJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				Truncate:        true,
				TruncateCascade: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"alltypes.all_postgres_types": &workflow_testdata.ExpectedOutput{RowCount: 2},
				"alltypes.array_types":        &workflow_testdata.ExpectedOutput{RowCount: 1},
				"alltypes.time_time":          &workflow_testdata.ExpectedOutput{RowCount: 2},
				"alltypes.json_data":          &workflow_testdata.ExpectedOutput{RowCount: 12},
			},
		},
		{
			Name:            "All Postgres types + init schema",
			Folder:          "testdata/postgres/all-types",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql"},
			JobMappings:     getJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"alltypes.all_postgres_types": &workflow_testdata.ExpectedOutput{RowCount: 2},
				"alltypes.array_types":        &workflow_testdata.ExpectedOutput{RowCount: 1},
				"alltypes.time_time":          &workflow_testdata.ExpectedOutput{RowCount: 2},
				"alltypes.json_data":          &workflow_testdata.ExpectedOutput{RowCount: 12},
			},
		},
	}
}

func getJobmappings() []*mgmtv1alpha1.JobMapping {
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		if jm.Column == "tax_amount" || jm.Column == "total_price" {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema: jm.Schema,
				Table:  jm.Table,
				Column: jm.Column,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{},
					},
				},
			})
		} else {
			updatedJobmappings = append(updatedJobmappings, jm)
		}
	}
	return updatedJobmappings
}
