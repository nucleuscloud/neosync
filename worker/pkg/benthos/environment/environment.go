package benthos_environment

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	continuation_token "github.com/nucleuscloud/neosync/internal/continuation-token"
	neosync_benthos_defaulttransform "github.com/nucleuscloud/neosync/worker/pkg/benthos/default_transform"
	neosync_benthos_dynamodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/dynamodb"
	neosync_benthos_error "github.com/nucleuscloud/neosync/worker/pkg/benthos/error"
	javascript_processor "github.com/nucleuscloud/neosync/worker/pkg/benthos/javascript"
	neosync_benthos_json "github.com/nucleuscloud/neosync/worker/pkg/benthos/json"
	benthos_metrics "github.com/nucleuscloud/neosync/worker/pkg/benthos/metrics"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	neosync_benthos_connectiondata "github.com/nucleuscloud/neosync/worker/pkg/benthos/neosync_connection_data"
	openaigenerate "github.com/nucleuscloud/neosync/worker/pkg/benthos/openai_generate"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/redpanda-data/benthos/v4/public/bloblang"
	"github.com/redpanda-data/benthos/v4/public/service"
	"go.opentelemetry.io/otel/metric"
)

type RegisterConfig struct {
	meter metric.Meter // nil to disable

	sqlConfig *SqlConfig // nil to disable

	mongoConfig *MongoConfig // nil to disable

	connectionDataConfig *ConnectionDataConfig // nil to diable

	stopChannel chan<- error

	blobEnv *bloblang.Environment
}

type Option func(cfg *RegisterConfig)

func WithMeter(meter metric.Meter) Option {
	return func(cfg *RegisterConfig) {
		cfg.meter = meter
	}
}

func WithSqlConfig(sqlcfg *SqlConfig) Option {
	return func(cfg *RegisterConfig) {
		cfg.sqlConfig = sqlcfg
	}
}
func WithStopChannel(c chan<- error) Option {
	return func(cfg *RegisterConfig) {
		cfg.stopChannel = c
	}
}
func WithMongoConfig(mongocfg *MongoConfig) Option {
	return func(cfg *RegisterConfig) {
		cfg.mongoConfig = mongocfg
	}
}
func WithConnectionDataConfig(connectionDataCfg *ConnectionDataConfig) Option {
	return func(cfg *RegisterConfig) {
		cfg.connectionDataConfig = connectionDataCfg
	}
}
func WithBlobEnv(b *bloblang.Environment) Option {
	return func(cfg *RegisterConfig) {
		cfg.blobEnv = b
	}
}

type SqlConfig struct {
	Provider               neosync_benthos_sql.ConnectionProvider
	IsRetry                bool
	InputHasMorePages      neosync_benthos_sql.OnHasMorePagesFn
	InputContinuationToken *continuation_token.ContinuationToken
}

type MongoConfig struct {
	Provider neosync_benthos_mongodb.MongoPoolProvider
}

type ConnectionDataConfig struct {
	NeosyncConnectionDataApi mgmtv1alpha1connect.ConnectionDataServiceClient
}

func NewEnvironment(logger *slog.Logger, opts ...Option) (*service.Environment, error) {
	return NewWithEnvironment(service.NewEnvironment(), logger, opts...)
}

func NewWithEnvironment(env *service.Environment, logger *slog.Logger, opts ...Option) (*service.Environment, error) {
	if env == nil {
		env = service.NewEnvironment()
	}
	config := &RegisterConfig{}

	for _, opt := range opts {
		opt(config)
	}

	if config.stopChannel == nil {
		return nil, errors.New("must provide non-nil StopChannel")
	}

	if config.meter != nil {
		err := benthos_metrics.RegisterOtelMetricsExporter(env, config.meter)
		if err != nil {
			return nil, fmt.Errorf("unable to register otel_collector for benthos metering: %w", err)
		}
	}

	if config.sqlConfig != nil {
		err := neosync_benthos_sql.RegisterPooledSqlInsertOutput(env, config.sqlConfig.Provider, config.sqlConfig.IsRetry, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_insert output to benthos instance: %w", err)
		}
		err = neosync_benthos_sql.RegisterPooledSqlUpdateOutput(env, config.sqlConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_update output to benthos instance: %w", err)
		}
		err = neosync_benthos_sql.RegisterPooledSqlRawInput(env, config.sqlConfig.Provider, config.stopChannel, config.sqlConfig.InputHasMorePages, config.sqlConfig.InputContinuationToken)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_raw input to benthos instance: %w", err)
		}
	}

	if config.mongoConfig != nil {
		err := neosync_benthos_mongodb.RegisterPooledMongoDbInput(env, config.mongoConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_mongodb input to benthos instance: %w", err)
		}
		err = neosync_benthos_mongodb.RegisterPooledMongoDbOutput(env, config.mongoConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_mongodb output to benthos instance: %w", err)
		}
	}

	if config.connectionDataConfig != nil {
		err := neosync_benthos_connectiondata.RegisterNeosyncConnectionDataInput(env, config.connectionDataConfig.NeosyncConnectionDataApi, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to register neosync_connection_data input: %w", err)
		}
	}

	err := openaigenerate.RegisterOpenaiGenerate(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register openai_generate input to benthos instance: %w", err)
	}

	err = neosync_benthos_error.RegisterErrorProcessor(env, config.stopChannel)
	if err != nil {
		return nil, fmt.Errorf("unable to register error processor to benthos instance: %w", err)
	}

	err = neosync_benthos_error.RegisterErrorOutput(env, config.stopChannel)
	if err != nil {
		return nil, fmt.Errorf("unable to register error output to benthos instance: %w", err)
	}

	err = neosync_benthos_dynamodb.RegisterDynamoDbInput(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register dynamodb input to benthos instance: %w", err)
	}

	err = neosync_benthos_dynamodb.RegisterDynamoDbOutput(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register dynamodb output to benthos instance: %w", err)
	}

	err = neosync_benthos_defaulttransform.ReisterDefaultTransformerProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register default mapping processor to benthos instance: %w", err)
	}

	err = neosync_benthos_json.RegisterNeosyncToJsonProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register Neosync to JSON processor to benthos instance: %w", err)
	}

	err = neosync_benthos_sql.RegisterNeosyncToPgxProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register Neosync to PGX processor to benthos instance: %w", err)
	}

	err = neosync_benthos_sql.RegisterNeosyncToMysqlProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register Neosync to MYSQL processor to benthos instance: %w", err)
	}

	err = neosync_benthos_sql.RegisterNeosyncToMssqlProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register Neosync to MSSQL processor to benthos instance: %w", err)
	}

	err = javascript_processor.RegisterNeosyncJavascriptProcessor(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register javascript processor to benthos instance: %w", err)
	}

	if config.blobEnv != nil {
		env.UseBloblangEnvironment(config.blobEnv)
	}

	return env, nil
}
