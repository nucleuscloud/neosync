package transformer_executor

import (
	"context"
	"log/slog"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	ee_transformer_fns "github.com/nucleuscloud/neosync/internal/ee/transformers/functions"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)

type piiTextApi struct {
	execConfig         *transformPiiTextConfig
	neosyncOperatorApi ee_transformer_fns.NeosyncOperatorApi
	logger             *slog.Logger
}

func newFromExecConfig(
	execConfig *transformPiiTextConfig,
	neosyncOperatorApi ee_transformer_fns.NeosyncOperatorApi,
	logger *slog.Logger,
) (transformers.TransformPiiTextApi, error) {
	return &piiTextApi{
		execConfig:         execConfig,
		neosyncOperatorApi: neosyncOperatorApi,
		logger:             logger,
	}, nil
}

func (p *piiTextApi) Transform(ctx context.Context, config *mgmtv1alpha1.TransformPiiText, value string) (string, error) {
	return ee_transformer_fns.TransformPiiText(
		ctx,
		p.execConfig.analyze,
		p.execConfig.anonymize,
		p.neosyncOperatorApi,
		config,
		value,
		p.logger,
	)
}
