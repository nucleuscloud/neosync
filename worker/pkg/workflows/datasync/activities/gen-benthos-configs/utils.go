package genbenthosconfigs_activity

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

func getMapValuesCount[K comparable, V any](m map[K][]V) int {
	count := 0
	for _, v := range m {
		count += len(v)
	}
	return count
}

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx := range cols {
		pieces[idx] = fmt.Sprintf("this.%q", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := make([]string, len(mappings))
	for idx := range mappings {
		columns[idx] = mappings[idx].Column
	}
	return columns
}

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	if t == nil {
		return false
	}

	source :=
		t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
			t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH
	config := t.GetConfig().GetConfig()
	passthrough := t.GetConfig().GetPassthroughConfig()
	return source || (config != nil && passthrough == nil)
}

func shouldProcessStrict(t *mgmtv1alpha1.JobMappingTransformer) bool {
	if t == nil {
		return false
	}

	source :=
		t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_UNSPECIFIED &&
			t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL &&
			t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_PASSTHROUGH &&
			t.GetSource() != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_DEFAULT
	config := t.GetConfig().GetConfig()
	genNull := t.GetConfig().GetNullconfig()
	passthrough := t.GetConfig().GetPassthroughConfig()
	genDefault := t.GetConfig().GetGenerateDefaultConfig()
	return source || (config != nil && genNull == nil && passthrough == nil && genDefault == nil)
}
