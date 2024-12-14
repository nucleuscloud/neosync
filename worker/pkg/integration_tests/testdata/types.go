package integration_tests

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

type ExpectedOutput struct {
	RowCount int
}

type DefaultTransformers struct {
	Boolean *mgmtv1alpha1.JobMappingTransformer
	String  *mgmtv1alpha1.JobMappingTransformer
	Number  *mgmtv1alpha1.JobMappingTransformer
	Byte    *mgmtv1alpha1.JobMappingTransformer
}

type TestJobOptions struct {
	SubsetByForeignKeyConstraints bool
	InitSchema                    bool
	Truncate                      bool
	TruncateCascade               bool
	DefaultTransformers           *DefaultTransformers
	SkipForeignKeyViolations      bool
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
	ExpectError        bool
	Expected           map[string]*ExpectedOutput // schema.table -> expected output
}
