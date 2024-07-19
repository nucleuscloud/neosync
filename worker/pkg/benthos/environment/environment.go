package benthos_environment

import (
	"errors"
	"fmt"

	neosync_benthos_error "github.com/nucleuscloud/neosync/worker/pkg/benthos/error"
	benthos_metrics "github.com/nucleuscloud/neosync/worker/pkg/benthos/metrics"
	neosync_benthos_mongodb "github.com/nucleuscloud/neosync/worker/pkg/benthos/mongodb"
	openaigenerate "github.com/nucleuscloud/neosync/worker/pkg/benthos/openai_generate"
	neosync_benthos_sql "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/warpstreamlabs/bento/public/service"
	"go.opentelemetry.io/otel/metric"
)

type RegisterConfig struct {
	Meter metric.Meter // nil to disable

	SqlConfig *SqlConfig // nil to disable

	MongoConfig *MongoConfig // nil to disable

	StopChannel chan<- error
}

type SqlConfig struct {
	Provider neosync_benthos_sql.DbPoolProvider
	IsRetry  bool
}

type MongoConfig struct {
	Provider neosync_benthos_mongodb.MongoPoolProvider
}

func New(config *RegisterConfig) (*service.Environment, error) {
	return NewWithEnvironment(service.NewEnvironment(), config)
}

func NewWithEnvironment(env *service.Environment, config *RegisterConfig) (*service.Environment, error) {
	if env == nil {
		env = service.NewEnvironment()
	}
	if config == nil {
		config = &RegisterConfig{}
	}
	if config.StopChannel == nil {
		return nil, errors.New("must provide non-nil StopChannel")
	}

	if config.Meter != nil {
		err := benthos_metrics.RegisterOtelMetricsExporter(env, config.Meter)
		if err != nil {
			return nil, fmt.Errorf("unable to register otel_collector for benthos metering: %w", err)
		}
	}

	if config.SqlConfig != nil {
		err := neosync_benthos_sql.RegisterPooledSqlInsertOutput(env, config.SqlConfig.Provider, config.SqlConfig.IsRetry)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_insert output to benthos instance: %w", err)
		}
		err = neosync_benthos_sql.RegisterPooledSqlUpdateOutput(env, config.SqlConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_update output to benthos instance: %w", err)
		}
		err = neosync_benthos_sql.RegisterPooledSqlRawInput(env, config.SqlConfig.Provider, config.StopChannel)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_sql_raw input to benthos instance: %w", err)
		}
	}

	if config.MongoConfig != nil {
		err := neosync_benthos_mongodb.RegisterPooledMongoDbInput(env, config.MongoConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_mongodb input to benthos instance: %w", err)
		}
		err = neosync_benthos_mongodb.RegisterPooledMongoDbOutput(env, config.MongoConfig.Provider)
		if err != nil {
			return nil, fmt.Errorf("unable to register pooled_mongodb output to benthos instance: %w", err)
		}
	}

	err := openaigenerate.RegisterOpenaiGenerate(env)
	if err != nil {
		return nil, fmt.Errorf("unable to register openai_generate input to benthos instance: %w", err)
	}

	err = neosync_benthos_error.RegisterErrorProcessor(env, config.StopChannel)
	if err != nil {
		return nil, fmt.Errorf("unable to register error processor to benthos instance: %w", err)
	}

	err = neosync_benthos_error.RegisterErrorOutput(env, config.StopChannel)
	if err != nil {
		return nil, fmt.Errorf("unable to register error output to benthos instance: %w", err)
	}

	return env, nil
}
