package benthos_environment

import (
	"errors"
	"fmt"
	"log/slog"

	neosync_benthos_defaulttransform "github.com/nucleuscloud/neosync/worker/pkg/benthos/default_transform"
	neosync_benthos_dynamodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/dynamodb"
	neosync_benthos_error "github.com/nucleuscloud/neosync/worker/pkg/benthos/error"
	benthos_metrics "github.com/nucleuscloud/neosync/worker/pkg/benthos/metrics"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	openaigenerate "github.com/nucleuscloud/neosync/worker/pkg/benthos/openai_generate"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/warpstreamlabs/bento/public/bloblang"
	"github.com/warpstreamlabs/bento/public/service"
	"go.opentelemetry.io/otel/metric"
)

type RegisterConfig struct {
	meter metric.Meter // nil to disable

	sqlConfig *SqlConfig // nil to disable

	mongoConfig *MongoConfig // nil to disable

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
func WithBlobEnv(b *bloblang.Environment) Option {
	return func(cfg *RegisterConfig) {
		cfg.blobEnv = b
	}
}

type SqlConfig struct {
	Provider neosync_benthos_sql.DbPoolProvider
	IsRetry  bool
}

type MongoConfig struct {
	Provider neosync_benthos_mongodb.MongoPoolProvider
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
		err = neosync_benthos_sql.RegisterPooledSqlRawInput(env, config.sqlConfig.Provider, config.stopChannel)
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

	if config.blobEnv != nil {
		env.UseBloblangEnvironment(config.blobEnv)
	}

	return env, nil
}