
package main

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

var JobMappings = []*mgmtv1alpha1.JobMapping{
	{
		Schema: "TODO",
		Table:  "regions",
		Column: "region_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "regions",
		Column: "region_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "countries",
		Column: "country_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "countries",
		Column: "country_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "countries",
		Column: "region_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "location_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "street_address",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "postal_code",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "city",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "state_province",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "locations",
		Column: "country_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "departments",
		Column: "department_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "departments",
		Column: "department_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "departments",
		Column: "location_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "jobs",
		Column: "job_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "jobs",
		Column: "job_title",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "jobs",
		Column: "min_salary",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "jobs",
		Column: "max_salary",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "employee_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "first_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "last_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "email",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "phone_number",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "hire_date",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "job_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "salary",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "NOT",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "manager_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "employees",
		Column: "department_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "dependents",
		Column: "dependent_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "dependents",
		Column: "first_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "dependents",
		Column: "last_name",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "dependents",
		Column: "relationship",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
	{
		Schema: "TODO",
		Table:  "dependents",
		Column: "employee_id",
		Transformer: &mgmtv1alpha1.JobMappingTransformer{
			Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH,
		},
	},
}
