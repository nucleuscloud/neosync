package testdata_doublereference

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Double reference sync",
			Folder:          "double-reference",
			SourceFilePaths: []string{"source-create.sql", "insert.sql"},
			TargetFilePaths: []string{"source-create.sql"},
			JobMappings:     getDefaultSyncJobMappings(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"double_reference.company":        &workflow_testdata.ExpectedOutput{RowCount: 3},
				"double_reference.department":     &workflow_testdata.ExpectedOutput{RowCount: 4},
				"double_reference.expense_report": &workflow_testdata.ExpectedOutput{RowCount: 2},
				"double_reference.transaction":    &workflow_testdata.ExpectedOutput{RowCount: 3},
			},
		},
		{
			Name:            "Double reference subset",
			Folder:          "double-reference",
			SourceFilePaths: []string{"source-create.sql", "insert.sql"},
			TargetFilePaths: []string{"source-create.sql"},
			SubsetMap: map[string]string{
				"double_reference.company": "id in (1)",
			},
			JobOptions: &workflow_testdata.TestJobOptions{
				SubsetByForeignKeyConstraints: true,
			},
			JobMappings: getDefaultSyncJobMappings(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"double_reference.company":        &workflow_testdata.ExpectedOutput{RowCount: 1},
				"double_reference.department":     &workflow_testdata.ExpectedOutput{RowCount: 2},
				"double_reference.expense_report": &workflow_testdata.ExpectedOutput{RowCount: 1},
				"double_reference.transaction":    &workflow_testdata.ExpectedOutput{RowCount: 2},
			},
		},
	}
}
