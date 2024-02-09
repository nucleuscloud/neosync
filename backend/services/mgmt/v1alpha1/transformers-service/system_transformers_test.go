package v1alpha1_transformersservice

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func Test_GetSystemTransformers(t *testing.T) {
	m := createServiceMock(t)

	resp, err := m.Service.GetSystemTransformers(context.Background(), &connect.Request[mgmtv1alpha1.GetSystemTransformersRequest]{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Msg.GetTransformers())
}

func Test_GetSystemTransformerBySource_NotFound(t *testing.T) {
	m := createServiceMock(t)

	resp, err := m.Service.GetSystemTransformerBySource(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
		Source: "i-do-not-exist",
	}))

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func Test_GetSystemTransformerBySource_Found(t *testing.T) {
	m := createServiceMock(t)

	resp, err := m.Service.GetSystemTransformerBySource(context.Background(), connect.NewRequest(&mgmtv1alpha1.GetSystemTransformerBySourceRequest{
		Source: string(Null),
	}))

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, resp.Msg.Transformer.Source, string(Null))
}
