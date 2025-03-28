package jsonanonymizer

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/benthos/transformer_executor"
)

type neosyncOperatorApi struct {
	opts []transformer_executor.TransformerExecutorOption
}

func newNeosyncOperatorApi(
	executorOpts []transformer_executor.TransformerExecutorOption,
) *neosyncOperatorApi {
	return &neosyncOperatorApi{opts: executorOpts}
}

func (n *neosyncOperatorApi) Transform(
	ctx context.Context,
	config *mgmtv1alpha1.TransformerConfig,
	value string,
) (string, error) {
	executor, err := transformer_executor.InitializeTransformerByConfigType(config, n.opts...)
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

type udtResolver struct {
	transformerClient mgmtv1alpha1connect.TransformersServiceClient
}

func newUdtResolver(transformerClient mgmtv1alpha1connect.TransformersServiceClient) *udtResolver {
	return &udtResolver{transformerClient: transformerClient}
}

func (u *udtResolver) GetUserDefinedTransformer(
	ctx context.Context,
	id string,
) (*mgmtv1alpha1.TransformerConfig, error) {
	resp, err := u.transformerClient.GetUserDefinedTransformerById(
		ctx,
		connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{
			TransformerId: id,
		}),
	)
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetTransformer().GetConfig(), nil
}
