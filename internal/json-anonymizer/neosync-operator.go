package jsonanonymizer

import (
	"context"
	"fmt"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	transformer "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type neosyncOperatorApi struct {
	logger *slog.Logger
}

func newNeosyncOperatorApi(logger *slog.Logger) *neosyncOperatorApi {
	return &neosyncOperatorApi{logger: logger}
}

func (n *neosyncOperatorApi) Transform(ctx context.Context, config *mgmtv1alpha1.TransformerConfig, value string) (string, error) {
	executor, err := transformer.InitializeTransformerByConfigType(config, transformer.WithLogger(n.logger))
	if err != nil {
		return "", err
	}
	result, err := executor.Mutate(value, executor.Opts)
	if err != nil {
		return "", err
	}
	switch result := result.(type) {
	case string:
		return result, nil
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%v", derefPointer(result)), nil
	}
}
