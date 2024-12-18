package mysql_compositekeys

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/internal/gotypeutil"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Composite key transformation + truncate",
			Folder:          "testdata/mysql/composite-keys",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql", "insert.sql"},
			JobMappings:     getPkTransformerJobmappings(),
			JobOptions: &workflow_testdata.TestJobOptions{
				InitSchema: false,
				Truncate:   true,
			},
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"composite.order_details":   &workflow_testdata.ExpectedOutput{RowCount: 10},
				"composite.orders":          &workflow_testdata.ExpectedOutput{RowCount: 10},
				"composite.order_shipping":  &workflow_testdata.ExpectedOutput{RowCount: 10},
				"composite.shipping_status": &workflow_testdata.ExpectedOutput{RowCount: 10},
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
