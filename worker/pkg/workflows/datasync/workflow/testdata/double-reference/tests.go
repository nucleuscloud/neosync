package testdata_doublereference

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

type ExpectedOutput struct {
	RowCount int
}
type IntegrationTest struct {
	Name            string
	SourceFilePaths []string
	TargetFilePaths []string
	SubsetMap       map[string]string                                         // schema.table -> where clause
	TransformerMap  map[string]map[string]*mgmtv1alpha1.JobMappingTransformer // schema.table.column -> transformer config
	OutputPath      string
	ExpectError     *bool
	Expected        map[string]*ExpectedOutput // schema.table -> expected output
}

var Tests = []*IntegrationTest{
	{
		Name:            "Double reference sync",
		SourceFilePaths: []string{"double-reference/source-create.sql", "double-reference/insert.sql"},
		TargetFilePaths: []string{"double-reference/source-create.sql"},
		OutputPath:      "double-reference/mocks.go",
		Expected: map[string]*ExpectedOutput{
			"double_reference.company":        &ExpectedOutput{RowCount: 3},
			"double_reference.department":     &ExpectedOutput{RowCount: 4},
			"double_reference.expense_report": &ExpectedOutput{RowCount: 2},
			"double_reference.transaction":    &ExpectedOutput{RowCount: 3},
		},
	},
	{
		Name:            "Double reference subset",
		SourceFilePaths: []string{"double-reference/source-create.sql", "double-reference/insert.sql"},
		TargetFilePaths: []string{"double-reference/source-create.sql"},
		SubsetMap: map[string]string{
			"double_reference.company": "id in (1)",
		},
		OutputPath: "double-reference/mocks.go",
		Expected: map[string]*ExpectedOutput{
			"double_reference.company":        &ExpectedOutput{RowCount: 1},
			"double_reference.department":     &ExpectedOutput{RowCount: 2},
			"double_reference.expense_report": &ExpectedOutput{RowCount: 1},
			"double_reference.transaction":    &ExpectedOutput{RowCount: 2},
		},
	},
}
