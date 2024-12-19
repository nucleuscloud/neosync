package neosync_benthos_connectiondata

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log/slog"
	"sync"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	neosync_dynamodb "github.com/nucleuscloud/neosync/internal/dynamodb"
	neosync_metadata "github.com/nucleuscloud/neosync/worker/pkg/benthos/metadata"

	neosyncgob "github.com/nucleuscloud/neosync/internal/gob"
	neosynctypes "github.com/nucleuscloud/neosync/internal/neosync-types"
	"github.com/warpstreamlabs/bento/public/service"
)

var _ = func() any { //nolint:unparam
	neosyncgob.RegisterGobTypes()
	return nil
}()

var neosyncConnectionDataConfigSpec = service.NewConfigSpec().
	Summary("Streams Neosync connection data").
	Field(service.NewStringField("connection_id")).
	Field(service.NewStringField("connection_type")).
	Field(service.NewStringField("schema")).
	Field(service.NewStringField("table")).
	Field(service.NewStringField("job_id").Optional()).
	Field(service.NewStringField("job_run_id").Optional())

func newNeosyncConnectionDataInput(
	conf *service.ParsedConfig,
	neosyncConnectApi mgmtv1alpha1connect.ConnectionDataServiceClient,
	logger *slog.Logger,
) (service.Input, error) {
	connectionId, err := conf.FieldString("connection_id")
	if err != nil {
		return nil, err
	}

	connectionType, err := conf.FieldString("connection_type")
	if err != nil {
		return nil, err
	}

	schema, err := conf.FieldString("schema")
	if err != nil {
		return nil, err
	}
	table, err := conf.FieldString("table")
	if err != nil {
		return nil, err
	}

	var jobId *string
	if conf.Contains("job_id") {
		jobIdStr, err := conf.FieldString("job_id")
		if err != nil {
			return nil, err
		}
		jobId = &jobIdStr
	}
	var jobRunId *string
	if conf.Contains("job_run_id") {
		jobRunIdStr, err := conf.FieldString("job_run_id")
		if err != nil {
			return nil, err
		}
		jobRunId = &jobRunIdStr
	}

	registry := neosynctypes.NewTypeRegistry(logger)

	return service.AutoRetryNacks(&neosyncInput{
		connectionId:   connectionId,
		connectionType: connectionType,
		schema:         schema,
		table:          table,
		connectionOpts: &connOpts{
			jobId:    jobId,
			jobRunId: jobRunId,
		},
		neosyncConnectApi:   neosyncConnectApi,
		neosyncTypeRegistry: registry,
		logger:              logger,
	}), nil
}

func RegisterNeosyncConnectionDataInput(
	env *service.Environment,
	neosyncConnectApi mgmtv1alpha1connect.ConnectionDataServiceClient,
	logger *slog.Logger,
) error {
	return env.RegisterInput(
		"neosync_connection_data", neosyncConnectionDataConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			return newNeosyncConnectionDataInput(conf, neosyncConnectApi, logger)
		},
	)
}

//------------------------------------------------------------------------------

type connOpts struct {
	jobId    *string
	jobRunId *string
}

type neosyncInput struct {
	connectionId   string
	connectionType string
	connectionOpts *connOpts
	schema         string
	table          string

	logger              *slog.Logger
	neosyncConnectApi   mgmtv1alpha1connect.ConnectionDataServiceClient
	neosyncTypeRegistry *neosynctypes.TypeRegistry

	recvMut sync.Mutex

	resp *connect.ServerStreamForClient[mgmtv1alpha1.GetConnectionDataStreamResponse]
}

func (g *neosyncInput) Connect(ctx context.Context) error {
	var streamCfg *mgmtv1alpha1.ConnectionStreamConfig

	if g.connectionType == string(benthosbuilder_shared.ConnectionTypeAwsS3) {
		awsS3Cfg := &mgmtv1alpha1.AwsS3StreamConfig{}
		if g.connectionOpts != nil {
			if g.connectionOpts.jobRunId != nil && *g.connectionOpts.jobRunId != "" {
				awsS3Cfg.Id = &mgmtv1alpha1.AwsS3StreamConfig_JobRunId{JobRunId: *g.connectionOpts.jobRunId}
			} else if g.connectionOpts.jobId != nil && *g.connectionOpts.jobId != "" {
				awsS3Cfg.Id = &mgmtv1alpha1.AwsS3StreamConfig_JobId{JobId: *g.connectionOpts.jobId}
			}
		}
		streamCfg = &mgmtv1alpha1.ConnectionStreamConfig{
			Config: &mgmtv1alpha1.ConnectionStreamConfig_AwsS3Config{
				AwsS3Config: awsS3Cfg,
			},
		}
	} else if g.connectionType == string(benthosbuilder_shared.ConnectionTypeGCP) {
		if g.connectionOpts != nil {
			gcpCfg := &mgmtv1alpha1.GcpCloudStorageStreamConfig{}
			if g.connectionOpts != nil {
				if g.connectionOpts.jobRunId != nil && *g.connectionOpts.jobRunId != "" {
					gcpCfg.Id = &mgmtv1alpha1.GcpCloudStorageStreamConfig_JobRunId{JobRunId: *g.connectionOpts.jobRunId}
				} else if g.connectionOpts.jobId != nil && *g.connectionOpts.jobId != "" {
					gcpCfg.Id = &mgmtv1alpha1.GcpCloudStorageStreamConfig_JobId{JobId: *g.connectionOpts.jobId}
				}
			}
			streamCfg = &mgmtv1alpha1.ConnectionStreamConfig{
				Config: &mgmtv1alpha1.ConnectionStreamConfig_GcpCloudstorageConfig{
					GcpCloudstorageConfig: gcpCfg,
				},
			}
		}
	}

	resp, err := g.neosyncConnectApi.GetConnectionDataStream(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionDataStreamRequest{
		ConnectionId: g.connectionId,
		Schema:       g.schema,
		Table:        g.table,
		StreamConfig: streamCfg,
	}))
	if err != nil {
		return err
	}
	g.resp = resp
	return nil
}

func (g *neosyncInput) Read(ctx context.Context) (*service.Message, service.AckFunc, error) {
	g.recvMut.Lock()
	defer g.recvMut.Unlock()

	if g.neosyncConnectApi == nil && g.resp == nil {
		return nil, nil, service.ErrNotConnected
	}
	if g.resp == nil {
		return nil, nil, service.ErrEndOfInput
	}

	ok := g.resp.Receive()
	if !ok {
		err := g.resp.Err()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, service.ErrEndOfInput
	}
	rowBytes := g.resp.Msg().RowBytes

	if g.connectionType == string(benthosbuilder_shared.ConnectionTypeDynamodb) {
		var dynamoDBItem map[string]any
		decoder := gob.NewDecoder(bytes.NewReader(rowBytes))
		err := decoder.Decode(&dynamoDBItem)
		if err != nil {
			return nil, nil, fmt.Errorf("error decoding data connection stream response with gob decoder: %w", err)
		}

		resMap, keyTypeMap := neosync_dynamodb.UnmarshalDynamoDBItem(dynamoDBItem)
		msg := service.NewMessage(nil)
		msg.MetaSetMut(neosync_metadata.MetaTypeMapStr, keyTypeMap)
		msg.SetStructuredMut(resMap)
		return msg, func(ctx context.Context, err error) error {
			// Nacks are retried automatically when we use service.AutoRetryNacks
			return nil
		}, nil
	}

	valuesMap := map[string]any{}
	decoder := gob.NewDecoder(bytes.NewReader(rowBytes))
	err := decoder.Decode(&valuesMap)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding data connection stream response with gob decoder: %w", err)
	}
	msg := service.NewMessage(nil)
	msg.SetStructuredMut(valuesMap)
	return msg, func(ctx context.Context, err error) error {
		// Nacks are retried automatically when we use service.AutoRetryNacks
		return nil
	}, nil
}

func (g *neosyncInput) Close(ctx context.Context) error {
	// close client
	// todo: prob need mutex
	if g.resp != nil {
		err := g.resp.Close()
		if err != nil {
			return err
		}
		g.resp = nil
	}

	g.neosyncConnectApi = nil // idk if this really matters
	return nil
}
