package postgres_humanresources

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

func GetVirtualForeignKeys(schema string) []*mgmtv1alpha1.VirtualForeignConstraint {
	return []*mgmtv1alpha1.VirtualForeignConstraint{
		{
			Schema:  schema,
			Table:   "countries",
			Columns: []string{"region_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "regions",
				Columns: []string{"region_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "departments",
			Columns: []string{"location_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "locations",
				Columns: []string{"location_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "dependents",
			Columns: []string{"employee_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "employees",
				Columns: []string{"employee_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "employees",
			Columns: []string{"manager_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "employees",
				Columns: []string{"employee_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "employees",
			Columns: []string{"department_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "departments",
				Columns: []string{"department_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "employees",
			Columns: []string{"job_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "jobs",
				Columns: []string{"job_id"},
			},
		},
		{
			Schema:  schema,
			Table:   "locations",
			Columns: []string{"country_id"},
			ForeignKey: &mgmtv1alpha1.VirtualForeignKey{
				Schema:  schema,
				Table:   "countries",
				Columns: []string{"country_id"},
			},
		},
	}
}
