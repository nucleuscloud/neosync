package mysql_multipledbs

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Circular Dependency sync + init schema",
			Folder:          "circular-dependencies",
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
		{
			Name:            "Circular Dependency subset + truncate",
			Folder:          "circular-dependencies",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"setup.sql"},
			SubsetMap: map[string]string{
				"circular_dependencies.orders": "id = 'f216a6f8-3bcd-46d8-8b99-e3b31dd5e6f3'",
			},
			JobOptions: &workflow_testdata.TestJobOptions{
				SubsetByForeignKeyConstraints: true,
				Truncate:                      true,
			},
			JobMappings: GetDefaultSyncJobMappings(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"circular_dependencies.addresses": &workflow_testdata.ExpectedOutput{RowCount: 1},
				"circular_dependencies.customers": &workflow_testdata.ExpectedOutput{RowCount: 1},
				"circular_dependencies.orders":    &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}
