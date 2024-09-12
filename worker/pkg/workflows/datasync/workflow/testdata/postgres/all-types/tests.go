package postgres_alltypes

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "All Postgres types",
			Folder:          "postgres/all-types",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql", "setup.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				Truncate:        true,
				TruncateCascade: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"ptypes.all_postgres_types": &workflow_testdata.ExpectedOutput{RowCount: 1},
				"ptypes.array_types":        &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
		{
			Name:            "All Postgres types + init schema",
			Folder:          "postgres/all-types",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"alltypes.all_postgres_types": &workflow_testdata.ExpectedOutput{RowCount: 1},
				"alltypes.array_types":        &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}
