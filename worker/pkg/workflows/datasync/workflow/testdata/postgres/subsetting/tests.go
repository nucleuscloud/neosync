package postgres_subsetting

import workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Complex subsetting",
			Folder:          "testdata/postgres/subsetting",
			SourceFilePaths: []string{"setup.sql"},
			TargetFilePaths: []string{"schema-create.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				SubsetByForeignKeyConstraints: true,
				InitSchema:                    true,
			},
			SubsetMap: map[string]string{
				"subsetting.users":     "user_id in (1,2,5,6,7,8)",
				"subsetting.test_2_x":  "created > '2023-06-03'",
				"subsetting.test_2_b":  "created > '2023-06-03'",
				"subsetting.addresses": "id in (1,5)",
				"subsetting.division":  "id in (3,5)",
				"subsetting.bosses":    "id in (3,5)",
				"subsetting.accounts":  "id = 1",
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"subsetting.attachments": &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.comments":    &workflow_testdata.ExpectedOutput{RowCount: 4},
				"subsetting.initiatives": &workflow_testdata.ExpectedOutput{RowCount: 4},
				"subsetting.skills":      &workflow_testdata.ExpectedOutput{RowCount: 10},
				"subsetting.tasks":       &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.user_skills": &workflow_testdata.ExpectedOutput{RowCount: 6},
				"subsetting.users":       &workflow_testdata.ExpectedOutput{RowCount: 6},
				"subsetting.test_2_x":    &workflow_testdata.ExpectedOutput{RowCount: 3},
				"subsetting.test_2_b":    &workflow_testdata.ExpectedOutput{RowCount: 3},
				"subsetting.test_2_a":    &workflow_testdata.ExpectedOutput{RowCount: 4},
				"subsetting.test_2_c":    &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.test_2_d":    &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.test_2_e":    &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.orders":      &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.addresses":   &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.customers":   &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.payments":    &workflow_testdata.ExpectedOutput{RowCount: 1},
				"subsetting.division":    &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.employees":   &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.projects":    &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.bosses":      &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.minions":     &workflow_testdata.ExpectedOutput{RowCount: 2},
				"subsetting.accounts":    &workflow_testdata.ExpectedOutput{RowCount: 1},
				"subsetting.blueprints":  &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}
