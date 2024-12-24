package mysql_multipledbs

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "multiple databases sync + init schema",
			Folder:          "testdata/mysql/multiple-dbs",
			SourceFilePaths: []string{"create-dbs.sql", "create.sql", "insert.sql"},
			TargetFilePaths: []string{"create-dbs.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"m_db_1.container":        &workflow_testdata.ExpectedOutput{RowCount: 10},
				"m_db_1.container_status": &workflow_testdata.ExpectedOutput{RowCount: 10},
				"m_db_2.container":        &workflow_testdata.ExpectedOutput{RowCount: 8},
				"m_db_2.container_status": &workflow_testdata.ExpectedOutput{RowCount: 8},
			},
		},
	}
}
