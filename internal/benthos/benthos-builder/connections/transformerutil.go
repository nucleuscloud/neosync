package benthosbuilder_connections

import mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH
}

func shouldProcessStrict(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH &&
		t.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT
}
