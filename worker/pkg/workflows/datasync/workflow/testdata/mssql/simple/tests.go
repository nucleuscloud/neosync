package mssql_simple

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Simple",
			Folder:          "mssql/simple",
			SourceFilePaths: []string{"create-schema.sql", "create-table.sql", "insert.sql"},
			TargetFilePaths: []string{"create-schema.sql", "create-table.sql"},
			JobMappings:     GetDefaultSyncJobMappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"simpledb.all_types": &workflow_testdata.ExpectedOutput{RowCount: 1},
			},
		},
	}
}

func GetDefaultSyncJobMappings() []*mgmtv1alpha1.JobMapping {
	return []*mgmtv1alpha1.JobMapping{
		{
			Schema: "simpledb",
			Table:  "all_types",
			Column: "intcol",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
			},
		},
	}
}
