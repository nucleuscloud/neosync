package ctganactivities

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	http_client "github.com/nucleuscloud/neosync/worker/internal/http/client"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/activity"
)

type GetTrainModelInputRequest struct {
	JobId string
}

type GetTrainModelInputResponse struct {
	Epochs          uint32
	DiscreteColumns []string
	ModelPath       string
	SourceDsn       *string
	Schema          *string
	Table           *string
	Columns         []string
}

func GetTrainModelInput(
	ctx context.Context,
	req *GetTrainModelInputRequest,
) (*GetTrainModelInputResponse, error) {
	logger := activity.GetLogger(ctx)
	_ = logger
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		neosyncUrl = "http://localhost:8080"
	}

	neosyncApiKey := viper.GetString("NEOSYNC_API_KEY")
	httpClient := http.DefaultClient
	if neosyncApiKey != "" {
		httpClient = http_client.NewWithHeaders(
			map[string]string{"Authorization": fmt.Sprintf("Bearer %s", neosyncApiKey)},
		)
	}
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)

	jobResp, err := jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: req.JobId,
	}))
	if err != nil {
		return nil, err
	}

	job := jobResp.Msg.Job

	trainSource := job.Source.Options.GetSingleTableCtganTrain()
	if trainSource == nil {
		return nil, errors.New("job source is not a single table train source type")
	}
	inputResponse := &GetTrainModelInputResponse{
		Epochs:          trainSource.Epochs,
		DiscreteColumns: trainSource.DiscreteColumns,
		Columns:         trainSource.Columns,
	}

	switch config := trainSource.ConnectionConfig.(type) {
	case *mgmtv1alpha1.TrainSingleTableCtganSourceOptions_Postgres:
		connection, err := getConnectionById(ctx, connclient, config.Postgres.ConnectionId)
		if err != nil {
			return nil, err
		}
		pgconfig := connection.ConnectionConfig.GetPgConfig()
		if pgconfig == nil {
			return nil, errors.New("source connection is not a postgres config")
		}
		dsn, err := getPgDsn(pgconfig)
		if err != nil {
			return nil, err
		}
		inputResponse.SourceDsn = &dsn
		inputResponse.Schema = &config.Postgres.Schema
		inputResponse.Table = &config.Postgres.Table
	default:
		return nil, errors.New("unsupported source option type")
	}

	if len(job.Destinations) != 1 {
		return nil, errors.New("train workflow only currently supports syncing to a single model location")
	}
	destId := job.Destinations[0].ConnectionId
	destConnection, err := getConnectionById(ctx, connclient, destId)
	if err != nil {
		return nil, err
	}
	localDirConfig := destConnection.ConnectionConfig.GetLocalDirConfig()
	if localDirConfig == nil {
		return nil, errors.New("destination connection is not a local directory config")
	}
	dstLocalDirConfig := job.Destinations[0].Options.GetLocalDirectoryOptions()
	if dstLocalDirConfig == nil {
		return nil, errors.New("job destination has no local directory options configured")
	}
	inputResponse.ModelPath = fmt.Sprintf("%s/%s", localDirConfig.Path, dstLocalDirConfig.FileName)

	return inputResponse, nil
}

func getConnectionById(
	ctx context.Context,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	connectionId string,
) (*mgmtv1alpha1.Connection, error) {
	getConnResp, err := connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}

func getPgDsn(
	config *mgmtv1alpha1.PostgresConnectionConfig,
) (string, error) {
	if config == nil {
		return "", errors.New("must provide non-nil config")
	}
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		if cfg.Connection == nil {
			return "", errors.New("must provide non-nil connection config")
		}
		dburl := fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s",
			cfg.Connection.User,
			cfg.Connection.Pass,
			cfg.Connection.Host,
			cfg.Connection.Port,
			cfg.Connection.Name,
		)
		if cfg.Connection.SslMode != nil && *cfg.Connection.SslMode != "" {
			dburl = fmt.Sprintf("%s?sslmode=%s", dburl, *cfg.Connection.SslMode)
		}
		return dburl, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		return cfg.Url, nil
	default:
		return "", fmt.Errorf("unsupported postgres connection config type")
	}
}
