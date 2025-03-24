package postgres_schemainit

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func GetAlteredSyncJobMappings(schema string) []*mgmtv1alpha1.JobMapping {
	return []*mgmtv1alpha1.JobMapping{
		{
			Schema: schema,
			Table:  "regions",
			Column: "is_active",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "regions",
			Column: "region_number",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenerateDefaultConfig{
						GenerateDefaultConfig: &mgmtv1alpha1.GenerateDefault{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "countries",
			Column: "last_update",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "locations",
			Column: "phone_numbers",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "jobs",
			Column: "job_type",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "employees",
			Column: "profile",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "employees",
			Column: "employee_code",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
		{
			Schema: schema,
			Table:  "jobs",
			Column: "job_code",
			Transformer: &mgmtv1alpha1.JobMappingTransformer{
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
						PassthroughConfig: &mgmtv1alpha1.Passthrough{},
					},
				},
			},
		},
	}
}
