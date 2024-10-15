package ee_transformers

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

var (
	TransformPiiText = &mgmtv1alpha1.SystemTransformer{
		Name:        "Transform PII Text",
		Description: "Transforms free-form text using PII analyzers",
		DataTypes: []mgmtv1alpha1.TransformerDataType{
			mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_STRING,
			mgmtv1alpha1.TransformerDataType_TRANSFORMER_DATA_TYPE_NULL,
		},
		SupportedJobTypes: []mgmtv1alpha1.SupportedJobType{mgmtv1alpha1.SupportedJobType_SUPPORTED_JOB_TYPE_SYNC},
		Source:            mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_PII_TEXT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformPiiTextConfig{
				TransformPiiTextConfig: &mgmtv1alpha1.TransformPiiText{
					ScoreThreshold: 0.5,
					DefaultAnonymizer: &mgmtv1alpha1.PiiAnonymizer{
						Config: &mgmtv1alpha1.PiiAnonymizer_Replace_{
							Replace: &mgmtv1alpha1.PiiAnonymizer_Replace{},
						},
					},
				},
			},
		},
	}

	Transformers = []*mgmtv1alpha1.SystemTransformer{
		TransformPiiText,
	}
)
