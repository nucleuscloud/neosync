package sync_activity

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/javascript"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/redis"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/redis"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"

	"connectrpc.com/connect"
	"github.com/benthosdev/benthos/v4/public/service"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"golang.org/x/sync/errgroup"
)

type SyncMetadata struct {
	Schema string
	Table  string
}

type SyncRequest struct {
	BenthosConfig string
	BenthosDsns   []*shared.BenthosDsn
}
type SyncResponse struct{}

func New(
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	tunnelmanager *ConnectionTunnelManager,
) *Activity {
	return &Activity{connclient: connclient, tunnelmanager: tunnelmanager}
}

type Activity struct {
	connclient    mgmtv1alpha1connect.ConnectionServiceClient
	tunnelmanager *ConnectionTunnelManager
}

// Temporal activity that runs benthos and syncs a source connection to one or more destination connections
func (a *Activity) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata, workflowMetadata *shared.WorkflowMetadata) (*SyncResponse, error) {
	loggerKeyVals := []any{
		"metadata", metadata,
		"WorkflowID", workflowMetadata.WorkflowId,
		"RunID", workflowMetadata.RunId,
	}
	logger := log.With(activity.GetLogger(ctx), loggerKeyVals...)
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With(loggerKeyVals...)
	var benthosStream *service.Stream
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.Stop(ctx)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				return
			}
		}
	}()

	connections := make([]*mgmtv1alpha1.Connection, len(req.BenthosDsns))

	errgrp, errctx := errgroup.WithContext(ctx)
	for idx, bdns := range req.BenthosDsns {
		idx := idx
		bdns := bdns
		errgrp.Go(func() error {
			resp, err := a.connclient.GetConnection(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{Id: bdns.ConnectionId}))
			if err != nil {
				return fmt.Errorf("failed to retrieve connection: %w", err)
			}
			connections[idx] = resp.Msg.Connection
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("unable to retrieve all or some connections: %w", err)
	}

	envKeyDsnSyncMap := sync.Map{}
	// tunnelMap := sync.Map{}
	defer func() {
		// tunnelMap.Range(func(key, value any) bool {
		// 	tunnel, ok := value.(*sshtunnel.Sshtunnel)
		// 	if !ok {
		// 		logger.Warn("unable to convert value to Sshtunnel for key", "key", key)
		// 		return true
		// 	}
		// 	tunnel.Close()
		// 	return true
		// })
		for _, conn := range connections {
			a.tunnelmanager.Release(conn.Id)
		}
	}()
	errgrp, errctx = errgroup.WithContext(ctx)
	for idx, bdns := range req.BenthosDsns {
		idx := idx
		bdns := bdns
		errgrp.Go(func() error {
			connection := connections[idx]

			localConnStr, err := a.tunnelmanager.GetConnectionString(connection, slogger)
			if err != nil {
				return err
			}
			logger.Info("***LOCAL CONN STR", "conn", localConnStr)

			// details, err := sqlconnect.GetConnectionDetails(connection.ConnectionConfig, shared.Ptr(uint32(5)), slogger)
			// if err != nil {
			// 	return fmt.Errorf("unable to get connection details for the given connection id: %s: %w", connection.Id, err)
			// }
			// if details.Tunnel != nil {
			// 	ready, err := details.Tunnel.Start()
			// 	if err != nil {
			// 		return fmt.Errorf("unable to start ssh tunnel: %w", err)
			// 	}
			// 	tunnelMap.Store(bdns.EnvVarKey, details.Tunnel)
			// 	<-ready
			// 	details.GeneralDbConnectConfig.Host = details.Tunnel.Local.Host
			// 	details.GeneralDbConnectConfig.Port = int32(details.Tunnel.Local.Port)
			// 	logger.Debug(
			// 		"ssh tunnel is ready, updated configuration host and port",
			// 		"host", details.Tunnel.Local.Host,
			// 		"port", details.Tunnel.Local.Port,
			// 	)
			// }
			// envKeyDsnSyncMap.Store(bdns.EnvVarKey, details.GeneralDbConnectConfig.String())
			envKeyDsnSyncMap.Store(bdns.EnvVarKey, localConnStr)
			return nil
		})
	}
	if err := errgrp.Wait(); err != nil {
		return nil, fmt.Errorf("was unable to build connection details for some or all connections: %w", err)
	}

	envKeyDnsMap := syncMapToStringMap(&envKeyDsnSyncMap)

	streambldr := service.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	streambldr.SetLogger(slogger.With(
		"benthos", "true",
	))

	// This must come before SetYaml as otherwise it will not be invoked
	streambldr.SetEnvVarLookupFunc(getEnvVarLookupFn(envKeyDnsMap))

	err := streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos config: %w", err)
	}
	benthosStream = stream
	err = stream.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStream = nil
	logger.Info("sync complete")
	return &SyncResponse{}, nil
}

func getEnvVarLookupFn(input map[string]string) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		if input == nil {
			return "", false
		}
		output, ok := input[key]
		return output, ok
	}
}

func syncMapToStringMap(incoming *sync.Map) map[string]string {
	out := map[string]string{}
	if incoming == nil {
		return out
	}

	incoming.Range(func(key, value any) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}
		valStr, ok := value.(string)
		if !ok {
			return true
		}
		out[keyStr] = valStr
		return true
	})
	return out
}
