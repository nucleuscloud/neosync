package workflow_testdata

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

type ExpectedOutput struct {
	RowCount int
}

type TestJobOptions struct {
	SubsetByForeignKeyConstraints bool
}
type IntegrationTest struct {
	Name               string
	SourceFilePaths    []string
	TargetFilePaths    []string
	Folder             string
	SubsetMap          map[string]string                                         // schema.table -> where clause
	TransformerMap     map[string]map[string]*mgmtv1alpha1.JobMappingTransformer // schema.table.column -> transformer config
	JobMappings        []*mgmtv1alpha1.JobMapping
	JobOptions         *TestJobOptions
	VirtualForeignKeys []*mgmtv1alpha1.VirtualForeignConstraint
	ExpectError        *bool
	Expected           map[string]*ExpectedOutput // schema.table -> expected output
}
