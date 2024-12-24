package mysql_alltypes

import (
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "All datatypes passthrough",
			Folder:          "testdata/mysql/all-types",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create-dbs.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
				Truncate:   true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"all_types.all_data_types": &workflow_testdata.ExpectedOutput{RowCount: 2},
				"all_types.json_data":      &workflow_testdata.ExpectedOutput{RowCount: 12},
			},
		},
	}
}
