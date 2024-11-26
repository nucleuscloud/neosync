package postgres_virtualforeignkeys

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Virtual Foreign Keys sync",
			Folder:          "testdata/postgres/virtual-foreign-keys",
			SourceFilePaths: []string{"source-setup.sql"},
			TargetFilePaths: []string{"target-setup.sql"},
			JobOptions: &workflow_testdata.TestJobOptions{
				Truncate:        false,
				TruncateCascade: false,
			},
			JobMappings:        GetDefaultSyncJobMappings(),
			VirtualForeignKeys: GetVirtualForeignKeys(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"vfk_hr.regions":     &workflow_testdata.ExpectedOutput{RowCount: 4},
				"vfk_hr.countries":   &workflow_testdata.ExpectedOutput{RowCount: 25},
				"vfk_hr.locations":   &workflow_testdata.ExpectedOutput{RowCount: 7},
				"vfk_hr.departments": &workflow_testdata.ExpectedOutput{RowCount: 11},
				"vfk_hr.dependents":  &workflow_testdata.ExpectedOutput{RowCount: 30},
				"vfk_hr.employees":   &workflow_testdata.ExpectedOutput{RowCount: 40},
				"vfk_hr.jobs":        &workflow_testdata.ExpectedOutput{RowCount: 19},
			},
		},
		{
			Name:            "Virtual Foreign Keys subset",
			Folder:          "testdata/postgres/virtual-foreign-keys",
			SourceFilePaths: []string{"source-setup.sql"},
			TargetFilePaths: []string{"target-setup.sql"},
			SubsetMap: map[string]string{
				"vfk_hr.employees": "first_name = 'Alexander'",
			},
			JobOptions: &workflow_testdata.TestJobOptions{
				SubsetByForeignKeyConstraints: true,
			},
			JobMappings:        GetDefaultSyncJobMappings(),
			VirtualForeignKeys: GetVirtualForeignKeys(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"vfk_hr.regions":     &workflow_testdata.ExpectedOutput{RowCount: 4},
				"vfk_hr.countries":   &workflow_testdata.ExpectedOutput{RowCount: 25},
				"vfk_hr.locations":   &workflow_testdata.ExpectedOutput{RowCount: 7},
				"vfk_hr.departments": &workflow_testdata.ExpectedOutput{RowCount: 11},
				"vfk_hr.dependents":  &workflow_testdata.ExpectedOutput{RowCount: 2},
				"vfk_hr.employees":   &workflow_testdata.ExpectedOutput{RowCount: 2},
				"vfk_hr.jobs":        &workflow_testdata.ExpectedOutput{RowCount: 19},
			},
		},
	}
}

func GetVirtualForeignKeys() []*mgmtv1alpha1.VirtualForeignConstraint {
	return []*mgmtv1alpha1.VirtualForeignConstraint{
		{
			Schema:  "vfk_hr",
			Table:   "countries",
			Columns: []string{"region_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "regions",
				Columns: []string{"region_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "departments",
			Columns: []string{"location_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "locations",
				Columns: []string{"location_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "dependents",
			Columns: []string{"employee_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "employees",
				Columns: []string{"employee_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "employees",
			Columns: []string{"manager_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "employees",
				Columns: []string{"employee_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "employees",
			Columns: []string{"department_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "departments",
				Columns: []string{"department_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "employees",
			Columns: []string{"job_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "jobs",
				Columns: []string{"job_id"},
			},
		},
		{
			Schema:  "vfk_hr",
			Table:   "locations",
			Columns: []string{"country_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  "vfk_hr",
				Table:   "countries",
				Columns: []string{"country_id"},
			},
		},
	}

}
