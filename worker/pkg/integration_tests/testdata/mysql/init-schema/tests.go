package mysql_initschema

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Init Schema",
			Folder:          "testdata/mysql/init-schema",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create-dbs.sql"},
			JobMappings:     getJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: true,
				Truncate:   true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"init_schema.container_status":             &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema.container":                    &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema2.container_status":            &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema2.container":                   &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.users":                       &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.unique_emails":               &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.unique_emails_and_usernames": &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.t1":                          &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.t2":                          &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.t3":                          &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.parent1":                     &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.child1":                      &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.t4":                          &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.t5":                          &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.employee_log":                &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.custom_table":                &workflow_testdata.ExpectedOutput{RowCount: 5},
				"init_schema3.tablewithcount":              &workflow_testdata.ExpectedOutput{RowCount: 5},
			},
		},
	}
}

func getJobmappings() []*mgmtv1alpha1.JobMapping {
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		if jm.Column != "fullname" {
			updatedJobmappings = append(updatedJobmappings, jm)
		} else {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema: jm.Schema,
				Table:  jm.Table,
				Column: jm.Column,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT,
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{},
					},
				},
			})
		}
	}
	return updatedJobmappings
}
