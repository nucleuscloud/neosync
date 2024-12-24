package mssql_datatypes

import (
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Passthrough",
			Folder:          "mssql/data-types",
			SourceFilePaths: []string{"create-schema.sql", "create-table.sql", "insert.sql"},
			TargetFilePaths: []string{"create-schema.sql", "create-table.sql"},
			JobMappings:     GetDefaultSyncJobMappings("other"),
			JobOptions:      &workflow_testdata.TestJobOptions{},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"alltypes.alldatatypes": &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}
