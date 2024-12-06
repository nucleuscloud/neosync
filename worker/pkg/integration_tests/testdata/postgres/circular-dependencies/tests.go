package postgres_circulardependencies

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Circular Dependency sync + init schema",
			Folder:          "testdata/postgres/circular-dependencies",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"circular_dependencies.addresses": &workflow_testdata.ExpectedOutput{RowCount: 8},
				"circular_dependencies.customers": &workflow_testdata.ExpectedOutput{RowCount: 10},
				"circular_dependencies.orders":    &workflow_testdata.ExpectedOutput{RowCount: 10},
			},
		},
	}
}
