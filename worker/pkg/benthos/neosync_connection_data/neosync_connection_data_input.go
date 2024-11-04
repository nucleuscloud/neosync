package neosync_benthos_connectiondata

import (
	"context"
	"encoding/json"
	"sync"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosync_dynamodb "github.com/nucleuscloud/neosync/internal/dynamodb"
	neosync_metadata "github.com/nucleuscloud/neosync/worker/pkg/benthos/metadata"
	"github.com/warpstreamlabs/bento/public/service"
)

var neosyncConnectionDataConfigSpec = service.NewConfigSpec().
	Summary("Streams Neosync connection data").
	Field(service.NewStringField("connection_id")).
	Field(service.NewStringField("connection_type")).
	Field(service.NewStringField("schema")).
	Field(service.NewStringField("table")).
	Field(service.NewStringField("job_id").Optional()).
	Field(service.NewStringField("job_run_id").Optional())

func newNeosyncConnectionDataInput(conf *service.ParsedConfig, neosyncConnectApi mgmtv1alpha1connect.ConnectionDataServiceClient) (service.Input, error) {
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

	return service.AutoRetryNacks(&neosyncInput{
		connectionId:   connectionId,
		connectionType: connectionType,
		schema:         schema,
		table:          table,
		connectionOpts: &connOpts{
			jobId:    jobId,
			jobRunId: jobRunId,
		},
		neosyncConnectApi: neosyncConnectApi,
	}), nil
}

func RegisterNeosyncConnectionDataInput(env *service.Environment, neosyncConnectApi mgmtv1alpha1connect.ConnectionDataServiceClient) error {
	return env.RegisterInput(
		"neosync_connection_data", neosyncConnectionDataConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Input, error) {
			return newNeosyncConnectionDataInput(conf, neosyncConnectApi)
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

	neosyncConnectApi mgmtv1alpha1connect.ConnectionDataServiceClient

	recvMut sync.Mutex

	resp *connect.ServerStreamForClient[mgmtv1alpha1.GetConnectionDataStreamResponse]
}

func (g *neosyncInput) Connect(ctx context.Context) error {
	var streamCfg *mgmtv1alpha1.ConnectionStreamConfig

	if g.connectionType == "aws-s3" {
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
	} else if g.connectionType == "gcp-cloud-storage" {
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
	row := g.resp.Msg().Row

	if g.connectionType == "awsDynamoDB" {
		for _, val := range row {
			var dynamoDBItem map[string]any
			err := json.Unmarshal(val, &dynamoDBItem)
			if err != nil {
				return nil, nil, err
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
	}
	valuesMap := map[string]any{}
	for col, val := range row {
		if len(val) == 0 {
			valuesMap[col] = nil
		} else {
			valuesMap[col] = val
		}
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
