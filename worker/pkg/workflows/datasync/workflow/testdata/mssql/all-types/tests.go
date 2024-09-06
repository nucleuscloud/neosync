package mssql_alltypes

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Passthrough",
			Folder:          "mssql/simple",
			SourceFilePaths: []string{"create-schema", "create-table.sql", "insert.sql"},
			TargetFilePaths: []string{"create-schema.sql", "create-table.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions:      &workflow_testdata.TestJobOptions{},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"alltypes.alldatatypes": &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}

ddd
