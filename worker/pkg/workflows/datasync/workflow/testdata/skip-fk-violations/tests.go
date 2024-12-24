package testdata_skipfkviolations

import (
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Skip Foreign Key Violations",
			Folder:          "testdata/skip-fk-violations",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				SkipForeignKeyViolations: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"fk_violations.jobs":        &workflow_testdata.ExpectedOutput{RowCount: 19},
				"fk_violations.regions":     &workflow_testdata.ExpectedOutput{RowCount: 4},
				"fk_violations.countries":   &workflow_testdata.ExpectedOutput{RowCount: 24},
				"fk_violations.dependents":  &workflow_testdata.ExpectedOutput{RowCount: 7},
				"fk_violations.employees":   &workflow_testdata.ExpectedOutput{RowCount: 10},
				"fk_violations.locations":   &workflow_testdata.ExpectedOutput{RowCount: 4},
				"fk_violations.departments": &workflow_testdata.ExpectedOutput{RowCount: 4},
			},
		},
		{
			Name:            "Foreign Key Violations Error",
			Folder:          "testdata/skip-fk-violations",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				SubsetByForeignKeyConstraints: false,
			},
			ExpectError: true,
			Expected:    map[string]*workflow_testdata.ExpectedOutput{},
		},
	}
}
