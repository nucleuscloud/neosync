package testdata_primarykeytransformer

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Circular Dependency primary key transformation",
			Folder:          "testdata/primary-key-transformer",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql"},
			JobMappings:     getPkTransformerJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: false,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"primary_$key.store_notifications": &workflow_testdata.ExpectedOutput{RowCount: 20},
				"primary_$key.stores":              &workflow_testdata.ExpectedOutput{RowCount: 20},
				"primary_$key.store_customers":     &workflow_testdata.ExpectedOutput{RowCount: 20},
				"primary_$key.referral_codes":      &workflow_testdata.ExpectedOutput{RowCount: 20},
			},
		},
	}
}

func getPkTransformerJobmappings() []*mgmtv1alpha1.JobMapping {
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		if jm.Column != "id" {
			updatedJobmappings = append(updatedJobmappings, jm)
		} else {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema: jm.Schema,
				Table:  jm.Table,
				Column: jm.Column,
				Transformer: &mgmtv1alpha1.JobMappingTransformer{
					Config: &mgmtv1alpha1.TransformerConfig{
						Config: &mgmtv1alpha1.TransformerConfig_GenerateUuidConfig{
							GenerateUuidConfig: &mgmtv1alpha1.GenerateUuid{
								IncludeHyphens: gotypeutil.ToPtr(true),
							},
						},
					},
				},
			})
		}
	}
	return updatedJobmappings
}
