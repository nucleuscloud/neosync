package sync_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	syncmap "sync"
	"time"

	"connectrpc.com/connect"
	charmlog "github.com/charmbracelet/log"
	"github.com/google/uuid"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	mssql_queries "github.com/nucleuscloud/neosync/backend/pkg/mssql-querier"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	sqlmanager_mysql "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/mysql"
	sqlmanager_postgres "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/postgres"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/internal/connection-tunnel-manager"
	pool_sql_provider "github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/pool/providers/sql"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers/sqlprovider"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	benthos_environment "github.com/nucleuscloud/neosync/worker/pkg/benthos/environment"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
	"github.com/warpstreamlabs/bento/public/bloblang"
	_ "github.com/warpstreamlabs/bento/public/components/aws"
	_ "github.com/warpstreamlabs/bento/public/components/io"
	_ "github.com/warpstreamlabs/bento/public/components/pure"
	_ "github.com/warpstreamlabs/bento/public/components/pure/extended"

	"github.com/warpstreamlabs/bento/public/service"
)

type ConnectionType string
type DriverType string

const (
	postgresDriver DriverType = "postgres"
	mysqlDriver    DriverType = "mysql"
	mssqlDriver    DriverType = "mssql"

	awsS3Connection           ConnectionType = "awsS3"
	gcpCloudStorageConnection ConnectionType = "gcpCloudStorage"
	postgresConnection        ConnectionType = "postgres"
	mysqlConnection           ConnectionType = "mysql"
	awsDynamoDBConnection     ConnectionType = "awsDynamoDB"

	batchSize = 20
)

var (
	driverMap = map[string]DriverType{
		string(postgresDriver): postgresDriver,
		string(mysqlDriver):    mysqlDriver,
	}
)

type cmdConfig struct {
	Source                 *sourceConfig              `yaml:"source"`
	Destination            *sqlDestinationConfig      `yaml:"destination"`
	AwsDynamoDbDestination *dynamoDbDestinationConfig `yaml:"aws-dynamodb-destination,omitempty"`
	Debug                  bool
	OutputType             *output.OutputType `yaml:"output-type,omitempty"`
	AccountId              *string            `yaml:"account-id,omitempty"`
}

type sourceConfig struct {
	ConnectionId   string          `yaml:"connection-id"`
	ConnectionOpts *connectionOpts `yaml:"connection-opts,omitempty"`
}

type connectionOpts struct {
	JobId    *string `yaml:"job-id,omitempty"`
	JobRunId *string `yaml:"job-run-id,omitempty"`
}

type onConflictConfig struct {
	DoNothing bool `yaml:"do-nothing"`
}

type dynamoDbDestinationConfig struct {
	AwsCredConfig *AwsCredConfig `yaml:"aws-cred-config"`
}

type sqlDestinationConfig struct {
	ConnectionUrl        string               `yaml:"connection-url"`
	Driver               DriverType           `yaml:"driver"`
	InitSchema           bool                 `yaml:"init-schema,omitempty"`
	TruncateBeforeInsert bool                 `yaml:"truncate-before-insert,omitempty"`
	TruncateCascade      bool                 `yaml:"truncate-cascade,omitempty"`
	OnConflict           onConflictConfig     `yaml:"on-conflict,omitempty"`
	ConnectionOpts       sqlConnectionOptions `yaml:"connection-opts,omitempty"`
}
type sqlConnectionOptions struct {
	OpenLimit    *int32  `yaml:"open-limit,omitempty"`
	IdleLimit    *int32  `yaml:"idle-limit,omitempty"`
	IdleDuration *string `yaml:"idle-duration,omitempty"`
	OpenDuration *string `yaml:"open-duration,omitempty"`
}

type AwsCredConfig struct {
	Region          string  `yaml:"region"`
	AccessKeyID     *string `yaml:"access-key-id,omitempty"`
	SecretAccessKey *string `yaml:"secret-access-key,omitempty"`
	SessionToken    *string `yaml:"session-token,omitempty"`
	RoleARN         *string `yaml:"role-arn,omitempty"`
	RoleExternalID  *string `yaml:"role-external-id,omitempty"`
	Endpoint        *string `yaml:"endpoint,omitempty"`
	Profile         *string `yaml:"profile,omitempty"`
}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "One off sync job to local resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKeyStr, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			var apiKey *string
			if apiKeyStr != "" {
				apiKey = &apiKeyStr
			}

			config, err := buildCmdConfig(cmd)
			if err != nil {
				return err
			}

			return sync(cmd.Context(), apiKey, config)
		},
	}

	cmd.Flags().String("connection-id", "", "Connection id for sync source")
	cmd.Flags().String("job-id", "", "Id of Job to sync data from. Only used with [AWS S3, GCP Cloud Storage] connections. Can use job-run-id instead.")
	cmd.Flags().String("job-run-id", "", "Id of Job run to sync data from. Only used with [AWS S3, GCP Cloud Storage] connections. Can use job-id instead.")
	cmd.Flags().String("destination-connection-url", "", "Connection url for sync output")
	cmd.Flags().String("destination-driver", "", "Connection driver for sync output")
	cmd.Flags().String("account-id", "", "Account source connection is in. Defaults to account id in cli context")
	cmd.Flags().String("config", "", "Location of config file")
	cmd.Flags().Bool("init-schema", false, "Create table schema and its constraints")
	cmd.Flags().Bool("truncate-before-insert", false, "Truncate table before insert")
	cmd.Flags().Bool("truncate-cascade", false, "Truncate cascade table before insert (postgres only)")
	cmd.Flags().Bool("on-conflict-do-nothing", false, "If there is a conflict when inserting data do not insert")

	cmd.Flags().Int32("destination-open-limit", 0, "Maximum number of open connections")
	cmd.Flags().Int32("destination-idle-limit", 0, "Maximum number of idle connections")
	cmd.Flags().String("destination-idle-duration", "", "Maximum amount of time a connection may be idle (e.g. '5m')")
	cmd.Flags().String("destination-open-duration", "", "Maximum amount of time a connection may be open (e.g. '30s')")

	// dynamo flags
	cmd.Flags().String("aws-access-key-id", "", "AWS Access Key ID for DynamoDB")
	cmd.Flags().String("aws-secret-access-key", "", "AWS Secret Access Key for DynamoDB")
	cmd.Flags().String("aws-session-token", "", "AWS Session Token for DynamoDB")
	cmd.Flags().String("aws-role-arn", "", "AWS Role ARN for DynamoDB")
	cmd.Flags().String("aws-role-external-id", "", "AWS Role External ID for DynamoDB")
	cmd.Flags().String("aws-profile", "", "AWS Profile for DynamoDB")
	cmd.Flags().String("aws-endpoint", "", "Custom endpoint for DynamoDB")
	cmd.Flags().String("aws-region", "", "AWS Region for DynamoDB")
	output.AttachOutputFlag(cmd)

	return cmd
}

type clisync struct {
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient
	connectionclient     mgmtv1alpha1connect.ConnectionServiceClient
	sqlmanagerclient     *sqlmanager.SqlManager
	sqlconnector         *sqlconnect.SqlOpenConnector
	benv                 *service.Environment
	cmd                  *cmdConfig
	logger               *slog.Logger
	ctx                  context.Context
}

func sync(
	ctx context.Context,
	apiKey *string,
	cmd *cmdConfig,
) error {
	logLevel := charmlog.InfoLevel
	if cmd.Debug {
		logLevel = charmlog.DebugLevel
	}
	charmlogger := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		ReportTimestamp: true,
		Level:           logLevel,
	})
	logger := slog.New(charmlogger)

	logger.Info("Starting sync")

	connectInterceptors := []connect.Interceptor{}
	neosyncurl := auth.GetNeosyncUrl()
	httpclient, err := auth.GetNeosyncHttpClient(ctx, apiKey, logger)
	if err != nil {
		return err
	}
	connectInterceptorOption := connect.WithInterceptors(connectInterceptors...)
	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(httpclient, neosyncurl, connectInterceptorOption)
	connectiondataclient := mgmtv1alpha1connect.NewConnectionDataServiceClient(httpclient, neosyncurl, connectInterceptorOption)

	pgpoolmap := &syncmap.Map{}
	mysqlpoolmap := &syncmap.Map{}
	mssqlpoolmap := &syncmap.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()
	sqlConnector := &sqlconnect.SqlOpenConnector{}
	sqlmanagerclient := sqlmanager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlConnector)

	sync := &clisync{
		connectiondataclient: connectiondataclient,
		connectionclient:     connectionclient,
		sqlmanagerclient:     sqlmanagerclient,
		sqlconnector:         sqlConnector,
		cmd:                  cmd,
		logger:               logger,
		ctx:                  ctx,
	}

	return sync.configureAndRunSync()
}

func (c *clisync) configureAndRunSync() error {
	connectionprovider := providers.NewProvider(
		mongoprovider.NewProvider(),
		sqlprovider.NewProvider(c.sqlconnector),
	)
	tunnelmanager := connectiontunnelmanager.NewConnectionTunnelManager(connectionprovider)
	session := uuid.NewString()
	// might not need this in cli context
	defer func() {
		tunnelmanager.ReleaseSession(session)
	}()

	destConnection := cmdConfigToDestinationConnection(c.cmd)
	dsnToConnIdMap := &syncmap.Map{}
	var sqlDsn string
	if c.cmd.Destination != nil {
		sqlDsn = c.cmd.Destination.ConnectionUrl
	}
	dsnToConnIdMap.Store(sqlDsn, destConnection.Id)
	stopChan := make(chan error, 3)
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-stopChan:
				cancel()
				return
			}
		}
	}()
	benthosEnv, err := benthos_environment.NewEnvironment(
		c.logger,
		benthos_environment.WithSqlConfig(&benthos_environment.SqlConfig{
			Provider: pool_sql_provider.NewProvider(pool_sql_provider.GetSqlPoolProviderGetter(
				tunnelmanager,
				dsnToConnIdMap,
				map[string]*mgmtv1alpha1.Connection{
					destConnection.Id: destConnection,
				},
				session,
				c.logger,
			)),
			IsRetry: false,
		}),
		benthos_environment.WithConnectionDataConfig(&benthos_environment.ConnectionDataConfig{
			NeosyncConnectionDataApi: c.connectiondataclient,
		}),
		benthos_environment.WithStopChannel(stopChan),
		benthos_environment.WithBlobEnv(bloblang.NewEnvironment()),
	)
	if err != nil {
		return err
	}
	c.benv = benthosEnv

	groupedConfigs, err := c.configureSync()
	if err != nil {
		return err
	}
	if groupedConfigs == nil {
		return nil
	}

	return runSync(c.ctx, *c.cmd.OutputType, c.benv, groupedConfigs, c.logger)
}

func (c *clisync) configureSync() ([][]*benthosConfigResponse, error) {
	c.logger.Debug("Retrieving neosync connection")
	connResp, err := c.connectionclient.GetConnection(c.ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: c.cmd.Source.ConnectionId,
	}))
	if err != nil {
		return nil, err
	}
	sourceConnection := connResp.Msg.GetConnection()
	sourceConnectionType, err := getConnectionType(sourceConnection)
	if err != nil {
		return nil, err
	}
	c.logger.Debug(fmt.Sprintf("Source connection type: %s", sourceConnectionType))

	err = isConfigValid(c.cmd, c.logger, sourceConnection, sourceConnectionType)
	if err != nil {
		return nil, err
	}
	c.logger.Debug("Validated config")

	c.logger.Info("Retrieving connection schema...")
	var schemaConfig *schemaConfig
	switch sourceConnectionType {
	case awsS3Connection:
		c.logger.Info("Building schema and table constraints...")
		var cfg *mgmtv1alpha1.AwsS3SchemaConfig
		if c.cmd.Source.ConnectionOpts.JobRunId != nil && *c.cmd.Source.ConnectionOpts.JobRunId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobRunId{JobRunId: *c.cmd.Source.ConnectionOpts.JobRunId}}
		} else if c.cmd.Source.ConnectionOpts.JobId != nil && *c.cmd.Source.ConnectionOpts.JobId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobId{JobId: *c.cmd.Source.ConnectionOpts.JobId}}
		}
		s3Config := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config{
				AwsS3Config: cfg,
			},
		}

		schemaCfg, err := c.getDestinationSchemaConfig(sourceConnection, s3Config)
		if err != nil {
			return nil, err
		}
		if len(schemaCfg.Schemas) == 0 {
			c.logger.Warn("No tables found when building destination schema from s3.")
			return nil, nil
		}
		schemaConfig = schemaCfg
	case gcpCloudStorageConnection:
		var cfg *mgmtv1alpha1.GcpCloudStorageSchemaConfig
		if c.cmd.Source.ConnectionOpts.JobRunId != nil && *c.cmd.Source.ConnectionOpts.JobRunId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobRunId{JobRunId: *c.cmd.Source.ConnectionOpts.JobRunId}}
		} else if c.cmd.Source.ConnectionOpts.JobId != nil && *c.cmd.Source.ConnectionOpts.JobId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobId{JobId: *c.cmd.Source.ConnectionOpts.JobId}}
		}

		gcpConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_GcpCloudstorageConfig{
				GcpCloudstorageConfig: cfg,
			},
		}

		schemaCfg, err := c.getDestinationSchemaConfig(sourceConnection, gcpConfig)
		if err != nil {
			return nil, err
		}
		if len(schemaCfg.Schemas) == 0 {
			c.logger.Warn("No tables found when building destination schema from gcp cloud storage.")
			return nil, nil
		}
		schemaConfig = schemaCfg
	case mysqlConnection:
		c.logger.Info("Building schema and table constraints...")
		mysqlCfg := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlSchemaConfig{},
			},
		}
		schemaCfg, err := c.getConnectionSchemaConfig(sourceConnection, mysqlCfg)
		if err != nil {
			return nil, err
		}
		if len(schemaCfg.Schemas) == 0 {
			c.logger.Warn("No tables found when building destination schema from mysql.")
			return nil, nil
		}
		schemaConfig = schemaCfg
	case postgresConnection:
		c.logger.Info("Building schema and table constraints...")
		postgresConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresSchemaConfig{},
			},
		}
		schemaCfg, err := c.getConnectionSchemaConfig(sourceConnection, postgresConfig)
		if err != nil {
			return nil, err
		}
		if len(schemaCfg.Schemas) == 0 {
			c.logger.Warn("No tables found when building destination schema from postgres.")
			return nil, nil
		}
		schemaConfig = schemaCfg
	case awsDynamoDBConnection:
		dynamoConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_DynamodbConfig{
				DynamodbConfig: &mgmtv1alpha1.DynamoDBSchemaConfig{},
			},
		}
		schemaCfg, err := c.getConnectionSchemaConfig(sourceConnection, dynamoConfig)
		if err != nil {
			return nil, err
		}
		if len(schemaCfg.Schemas) == 0 {
			c.logger.Warn("No tables found when building destination schema from dynamodb.")
			return nil, nil
		}
		tableMap := map[string]struct{}{}
		for _, s := range schemaCfg.Schemas {
			tableMap[s.Table] = struct{}{}
		}
		configs := []*benthosConfigResponse{}
		for t := range tableMap {
			benthosConfig := generateDynamoDbBenthosConfig(c.cmd, t)
			configs = append(configs, benthosConfig)
		}
		return [][]*benthosConfigResponse{configs}, nil
	default:
		return nil, fmt.Errorf("this connection type is not currently supported: %T", sourceConnectionType)
	}

	c.logger.Debug("Building sync configs")
	syncConfigs := buildSyncConfigs(schemaConfig, c.logger)
	if syncConfigs == nil {
		return nil, nil
	}
	c.logger.Info("Running table init statements...")
	err = c.runDestinationInitStatements(syncConfigs, schemaConfig)
	if err != nil {
		return nil, err
	}

	syncConfigCount := len(syncConfigs)
	c.logger.Info(fmt.Sprintf("Generating %d sync configs...", syncConfigCount))
	configs := []*benthosConfigResponse{}
	for _, cfg := range syncConfigs {
		benthosConfig := generateBenthosConfig(c.cmd, sourceConnectionType, cfg)
		configs = append(configs, benthosConfig)
	}

	// order configs in run order by dependency
	c.logger.Debug("Ordering configs by dependency")
	groupedConfigs := groupConfigsByDependency(configs, c.logger)

	return groupedConfigs, nil
}

var (
	// Hack that locks the instanced bento stream builder build step that causes data races if done in parallel
	streamBuilderMu syncmap.Mutex
)

func syncData(ctx context.Context, benv *service.Environment, cfg *benthosConfigResponse, logger *slog.Logger, outputType output.OutputType) error {
	configbits, err := yaml.Marshal(cfg.Config)
	if err != nil {
		return err
	}

	benthosStreamMutex := syncmap.Mutex{}
	var benthosStream *service.Stream
	go func() {
		for { //nolint
			select {
			case <-ctx.Done():
				benthosStreamMutex.Lock()
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.StopWithin(1 * time.Millisecond)
					if err != nil {
						logger.Error(err.Error())
					}
				}
				benthosStreamMutex.Unlock()
				return
			}
		}
	}()

	split := strings.Split(cfg.Name, ".")
	var runType string
	if len(split) != 0 {
		runType = split[len(split)-1]
	}
	streamBuilderMu.Lock()
	streambldr := benv.NewStreamBuilder()
	if streambldr == nil {
		return fmt.Errorf("failed to create StreamBuilder")
	}
	if outputType == output.PlainOutput {
		streambldr.SetLogger(logger.With("benthos", "true", "table", cfg.Table, "runType", runType))
	}
	if benv == nil {
		return fmt.Errorf("benthos env is nil")
	}

	err = streambldr.SetYAML(string(configbits))
	if err != nil {
		return fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	streamBuilderMu.Unlock()
	if err != nil {
		return err
	}
	benthosStreamMutex.Lock()
	benthosStream = stream
	benthosStreamMutex.Unlock()

	err = stream.Run(ctx)
	if err != nil {
		return fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStreamMutex.Lock()
	benthosStream = nil
	benthosStreamMutex.Unlock()
	return nil
}

func toSqlConnectionOptions(cfg sqlConnectionOptions) *mgmtv1alpha1.SqlConnectionOptions {
	outputOptions := &mgmtv1alpha1.SqlConnectionOptions{
		MaxConnectionLimit: shared.Ptr(int32(25)),
	}

	if cfg.OpenLimit != nil {
		outputOptions.MaxConnectionLimit = cfg.OpenLimit
	}
	if cfg.IdleLimit != nil {
		outputOptions.MaxIdleConnections = cfg.IdleLimit
	}
	if cfg.OpenDuration != nil {
		outputOptions.MaxOpenDuration = cfg.OpenDuration
	}
	if cfg.IdleDuration != nil {
		outputOptions.MaxIdleDuration = cfg.IdleDuration
	}

	return outputOptions
}

func cmdConfigToDestinationConnection(cmd *cmdConfig) *mgmtv1alpha1.Connection {
	destId := uuid.NewString()
	if cmd.Destination != nil {
		switch cmd.Destination.Driver {
		case postgresDriver:
			return &mgmtv1alpha1.Connection{
				Id:   destId,
				Name: destId,
				ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
					Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
						PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
							ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
								Url: cmd.Destination.ConnectionUrl,
							},
							ConnectionOptions: toSqlConnectionOptions(cmd.Destination.ConnectionOpts),
						},
					},
				},
			}
		case mysqlDriver:
			return &mgmtv1alpha1.Connection{
				Id:   destId,
				Name: destId,
				ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
					Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
						MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
							ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
								Url: cmd.Destination.ConnectionUrl,
							},
							ConnectionOptions: toSqlConnectionOptions(cmd.Destination.ConnectionOpts),
						},
					},
				},
			}
		case mssqlDriver:
			return &mgmtv1alpha1.Connection{
				Id:   destId,
				Name: destId,
				ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
					Config: &mgmtv1alpha1.ConnectionConfig_MssqlConfig{
						MssqlConfig: &mgmtv1alpha1.MssqlConnectionConfig{
							ConnectionConfig: &mgmtv1alpha1.MssqlConnectionConfig_Url{
								Url: cmd.Destination.ConnectionUrl,
							},
							ConnectionOptions: toSqlConnectionOptions(cmd.Destination.ConnectionOpts),
						},
					},
				},
			}
		}
	} else if cmd.AwsDynamoDbDestination != nil {
		creds := &mgmtv1alpha1.AwsS3Credentials{}
		if cmd.AwsDynamoDbDestination.AwsCredConfig != nil {
			cfg := cmd.AwsDynamoDbDestination.AwsCredConfig
			creds.Profile = cfg.Profile
			creds.AccessKeyId = cfg.AccessKeyID
			creds.SecretAccessKey = cfg.SecretAccessKey
			creds.SessionToken = cfg.SessionToken
			creds.RoleArn = cfg.RoleARN
			creds.RoleExternalId = cfg.RoleExternalID
		}
		return &mgmtv1alpha1.Connection{
			Id:   destId,
			Name: destId,
			ConnectionConfig: &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_DynamodbConfig{
					DynamodbConfig: &mgmtv1alpha1.DynamoDBConnectionConfig{
						Credentials: creds,
					},
				},
			},
		}
	}
	return &mgmtv1alpha1.Connection{}
}

func (c *clisync) runDestinationInitStatements(
	syncConfigs []*tabledependency.RunConfig,
	schemaConfig *schemaConfig,
) error {
	dependencyMap := buildDependencyMap(syncConfigs)
	db, err := c.sqlmanagerclient.NewSqlDbFromUrl(c.ctx, string(c.cmd.Destination.Driver), c.cmd.Destination.ConnectionUrl)
	if err != nil {
		return err
	}
	defer db.Db.Close()
	if c.cmd.Destination.InitSchema {
		if len(schemaConfig.InitSchemaStatements) != 0 {
			for _, block := range schemaConfig.InitSchemaStatements {
				c.logger.Info(fmt.Sprintf("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements)))
				if len(block.Statements) == 0 {
					continue
				}
				err = db.Db.BatchExec(c.ctx, batchSize, block.Statements, &sql_manager.BatchExecOpts{})
				if err != nil {
					c.logger.Error(fmt.Sprintf("Error creating tables: %v", err))
					return fmt.Errorf("unable to exec pg %s statements: %w", block.Label, err)
				}
			}
		} else if len(schemaConfig.InitTableStatementsMap) != 0 {
			// @deprecated mysql init table statements
			orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(dependencyMap)
			if err != nil {
				return err
			}
			if orderedTablesResp.HasCycles {
				return errors.New("init schema: unable to handle circular dependencies")
			}
			orderedInitStatements := []string{}
			for _, t := range orderedTablesResp.OrderedTables {
				orderedInitStatements = append(orderedInitStatements, schemaConfig.InitTableStatementsMap[t.String()])
			}

			err = db.Db.BatchExec(c.ctx, batchSize, orderedInitStatements, &sql_manager.BatchExecOpts{})
			if err != nil {
				c.logger.Error(fmt.Sprintf("Error creating tables: %v", err))
				return err
			}
		}
	}
	if c.cmd.Destination.Driver == postgresDriver {
		if c.cmd.Destination.TruncateCascade {
			truncateCascadeStmts := []string{}
			for _, syncCfg := range syncConfigs {
				stmt, ok := schemaConfig.TruncateTableStatementsMap[syncCfg.Table()]
				if ok {
					truncateCascadeStmts = append(truncateCascadeStmts, stmt)
				}
			}
			err = db.Db.BatchExec(c.ctx, batchSize, truncateCascadeStmts, &sql_manager.BatchExecOpts{})
			if err != nil {
				c.logger.Error(fmt.Sprintf("Error truncate cascade tables: %v", err))
				return err
			}
		} else if c.cmd.Destination.TruncateBeforeInsert {
			orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(dependencyMap)
			if err != nil {
				return err
			}
			orderedTruncateStatement, err := sqlmanager_postgres.BuildPgTruncateStatement(orderedTablesResp.OrderedTables)
			if err != nil {
				return err
			}
			err = db.Db.Exec(c.ctx, orderedTruncateStatement)
			if err != nil {
				c.logger.Error(fmt.Sprintf("Error truncating tables: %v", err))
				return err
			}
		}
	} else if c.cmd.Destination.Driver == mysqlDriver {
		orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(dependencyMap)
		if err != nil {
			return err
		}
		orderedTableTruncateStatements := []string{}
		for _, t := range orderedTablesResp.OrderedTables {
			orderedTableTruncateStatements = append(orderedTableTruncateStatements, schemaConfig.TruncateTableStatementsMap[t.String()])
		}
		disableFkChecks := sql_manager.DisableForeignKeyChecks
		err = db.Db.BatchExec(c.ctx, batchSize, orderedTableTruncateStatements, &sql_manager.BatchExecOpts{Prefix: &disableFkChecks})
		if err != nil {
			c.logger.Error(fmt.Sprintf("Error truncating tables: %v", err))
			return err
		}
	}
	return nil
}

func buildSyncConfigs(
	schemaConfig *schemaConfig,
	logger *slog.Logger,
) []*tabledependency.RunConfig {
	tableColMap := getTableColMap(schemaConfig.Schemas)
	if len(tableColMap) == 0 {
		return nil
	}
	primaryKeysMap := map[string][]string{}
	for table, constraints := range schemaConfig.TablePrimaryKeys {
		primaryKeysMap[table] = constraints.GetColumns()
	}

	runConfigs, err := tabledependency.GetRunConfigs(schemaConfig.TableConstraints, map[string]string{}, primaryKeysMap, tableColMap)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	return runConfigs
}

func getTableInitStatementMap(
	ctx context.Context,
	logger *slog.Logger,
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	connectionId string,
	opts *sqlDestinationConfig,
) (*mgmtv1alpha1.GetConnectionInitStatementsResponse, error) {
	if opts.InitSchema || opts.TruncateBeforeInsert || opts.TruncateCascade {
		logger.Info("Creating init statements...")
		truncateBeforeInsert := opts.TruncateBeforeInsert
		if opts.Driver == postgresDriver && truncateBeforeInsert {
			// postgres truncate must be ordered properly
			// handled in runDestinationInitStatements function
			truncateBeforeInsert = false
		}
		initStatementResp, err := connectiondataclient.GetConnectionInitStatements(ctx,
			connect.NewRequest(&mgmtv1alpha1.GetConnectionInitStatementsRequest{
				ConnectionId: connectionId,
				Options: &mgmtv1alpha1.InitStatementOptions{
					InitSchema:           opts.InitSchema,
					TruncateBeforeInsert: truncateBeforeInsert,
					TruncateCascade:      opts.TruncateCascade,
				},
			},
			))
		if err != nil {
			return nil, err
		}
		return initStatementResp.Msg, nil
	}
	return nil, nil
}

type benthosConfigResponse struct {
	Name      string
	DependsOn []*tabledependency.DependsOn
	Config    *neosync_benthos.BenthosConfig
	Table     string
	Columns   []string
}

func generateBenthosConfig(
	cmd *cmdConfig,
	connectionType ConnectionType,
	syncConfig *tabledependency.RunConfig,
) *benthosConfigResponse {
	schema, table := sqlmanager_shared.SplitTableKey(syncConfig.Table())

	var jobId, jobRunId *string
	if cmd.Source.ConnectionOpts != nil {
		jobRunId = cmd.Source.ConnectionOpts.JobRunId
		jobId = cmd.Source.ConnectionOpts.JobId
	}

	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Logger: &neosync_benthos.LoggerConfig{
				Level:        "ERROR",
				AddTimestamp: true,
			},
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
						ConnectionId:   cmd.Source.ConnectionId,
						ConnectionType: string(connectionType),
						JobId:          jobId,
						JobRunId:       jobRunId,
						Schema:         schema,
						Table:          table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output:   &neosync_benthos.OutputConfig{},
		},
	}

	if syncConfig.RunType() == tabledependency.RunTypeUpdate {
		args := syncConfig.InsertColumns()
		args = append(args, syncConfig.PrimaryKeys()...)
		bc.Output = &neosync_benthos.OutputConfig{
			Outputs: neosync_benthos.Outputs{
				PooledSqlUpdate: &neosync_benthos.PooledSqlUpdate{
					Driver: string(cmd.Destination.Driver),
					Dsn:    cmd.Destination.ConnectionUrl,

					Schema:       schema,
					Table:        table,
					Columns:      syncConfig.InsertColumns(),
					WhereColumns: syncConfig.PrimaryKeys(),
					ArgsMapping:  buildPlainInsertArgs(args),

					Batching: &neosync_benthos.Batching{
						Period: "5s",
						Count:  100,
					},
				},
			},
		}
	} else {
		bc.Output = &neosync_benthos.OutputConfig{
			Outputs: neosync_benthos.Outputs{
				PooledSqlInsert: &neosync_benthos.PooledSqlInsert{
					Driver: string(cmd.Destination.Driver),
					Dsn:    cmd.Destination.ConnectionUrl,

					Schema:              schema,
					Table:               table,
					Columns:             syncConfig.SelectColumns(),
					OnConflictDoNothing: cmd.Destination.OnConflict.DoNothing,
					ArgsMapping:         buildPlainInsertArgs(syncConfig.SelectColumns()),

					Batching: &neosync_benthos.Batching{
						Period: "5s",
						Count:  100,
					},
				},
			},
		}
	}

	return &benthosConfigResponse{
		Name:      fmt.Sprintf("%s.%s", syncConfig.Table(), syncConfig.RunType()),
		Config:    bc,
		DependsOn: syncConfig.DependsOn(),
		Table:     syncConfig.Table(),
		Columns:   syncConfig.InsertColumns(),
	}
}

type schemaConfig struct {
	Schemas                    []*mgmtv1alpha1.DatabaseColumn
	TableConstraints           map[string][]*sql_manager.ForeignConstraint
	TablePrimaryKeys           map[string]*mgmtv1alpha1.PrimaryConstraint
	InitTableStatementsMap     map[string]string
	TruncateTableStatementsMap map[string]string
	InitSchemaStatements       []*mgmtv1alpha1.SchemaInitStatements
}

func (c *clisync) getConnectionSchemaConfig(
	connection *mgmtv1alpha1.Connection,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
) (*schemaConfig, error) {
	var schemas []*mgmtv1alpha1.DatabaseColumn
	var tableConstraints map[string]*mgmtv1alpha1.ForeignConstraintTables
	var tablePrimaryKeys map[string]*mgmtv1alpha1.PrimaryConstraint
	var initTableStatementsMap map[string]string
	var truncateTableStatementsMap map[string]string
	var initSchemaStatements []*mgmtv1alpha1.SchemaInitStatements
	errgrp, errctx := errgroup.WithContext(c.ctx)
	errgrp.Go(func() error {
		schemaResp, err := c.connectiondataclient.GetConnectionSchema(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: connection.Id,
			SchemaConfig: sc,
		}))
		if err != nil {
			return err
		}
		schemas = schemaResp.Msg.GetSchemas()
		return nil
	})

	errgrp.Go(func() error {
		constraintConnectionResp, err := c.connectiondataclient.GetConnectionTableConstraints(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionTableConstraintsRequest{ConnectionId: c.cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tableConstraints = constraintConnectionResp.Msg.GetForeignKeyConstraints()
		tablePrimaryKeys = constraintConnectionResp.Msg.GetPrimaryKeyConstraints()
		return nil
	})

	errgrp.Go(func() error {
		initStatementsResp, err := getTableInitStatementMap(errctx, c.logger, c.connectiondataclient, c.cmd.Source.ConnectionId, c.cmd.Destination)
		if err != nil {
			return err
		}
		initTableStatementsMap = initStatementsResp.GetTableInitStatements()
		truncateTableStatementsMap = initStatementsResp.GetTableTruncateStatements()
		initSchemaStatements = initStatementsResp.GetSchemaInitStatements()
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}
	tc := map[string][]*sql_manager.ForeignConstraint{}
	for table, constraints := range tableConstraints {
		fkConstraints := []*sql_manager.ForeignConstraint{}
		for _, fk := range constraints.GetConstraints() {
			var foreignKey *sql_manager.ForeignKey
			if fk.ForeignKey != nil {
				foreignKey = &sql_manager.ForeignKey{
					Table:   fk.GetForeignKey().GetTable(),
					Columns: fk.GetForeignKey().GetColumns(),
				}
			}
			fkConstraints = append(fkConstraints, &sql_manager.ForeignConstraint{
				Columns:     fk.GetColumns(),
				NotNullable: fk.GetNotNullable(),
				ForeignKey:  foreignKey,
			})
		}
		tc[table] = fkConstraints
	}

	return &schemaConfig{
		Schemas:                    schemas,
		TableConstraints:           tc,
		TablePrimaryKeys:           tablePrimaryKeys,
		InitTableStatementsMap:     initTableStatementsMap,
		TruncateTableStatementsMap: truncateTableStatementsMap,
		InitSchemaStatements:       initSchemaStatements,
	}, nil
}

func (c *clisync) getDestinationSchemaConfig(
	connection *mgmtv1alpha1.Connection,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
) (*schemaConfig, error) {
	schemaResp, err := c.connectiondataclient.GetConnectionSchema(c.ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		ConnectionId: connection.Id,
		SchemaConfig: sc,
	}))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve connection schema for connection: %w", err)
	}

	tableColMap := getTableColMap(schemaResp.Msg.GetSchemas())
	if len(tableColMap) == 0 {
		c.logger.Warn("no tables found after retrieving connection schema.")
		return &schemaConfig{}, nil
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp.Msg.GetSchemas() {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	c.logger.Info(fmt.Sprintf("Building table constraints for %d schemas...", len(schemas)))
	tableConstraints, err := c.getDestinationTableConstraints(schemas)
	if err != nil {
		return nil, fmt.Errorf("unable to build destination tablle constraints: %w", err)
	}

	primaryKeys := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for tableName, cols := range tableConstraints.PrimaryKeyConstraints {
		primaryKeys[tableName] = &mgmtv1alpha1.PrimaryConstraint{
			Columns: cols,
		}
	}

	truncateTableStatementsMap := map[string]string{}
	if c.cmd.Destination.Driver == postgresDriver {
		if c.cmd.Destination.TruncateCascade {
			for t := range tableColMap {
				schema, table := sqlmanager_shared.SplitTableKey(t)
				stmt, err := sqlmanager_postgres.BuildPgTruncateCascadeStatement(schema, table)
				if err != nil {
					return nil, err
				}
				truncateTableStatementsMap[t] = stmt
			}
		}
		// truncate before insert handled in runDestinationInitStatements
	} else {
		if c.cmd.Destination.TruncateBeforeInsert {
			for t := range tableColMap {
				schema, table := sqlmanager_shared.SplitTableKey(t)
				stmt, err := sqlmanager_mysql.BuildMysqlTruncateStatement(schema, table)
				if err != nil {
					return nil, err
				}
				truncateTableStatementsMap[t] = stmt
			}
		}
	}

	return &schemaConfig{
		Schemas:                    schemaResp.Msg.GetSchemas(),
		TableConstraints:           tableConstraints.ForeignKeyConstraints,
		TablePrimaryKeys:           primaryKeys,
		TruncateTableStatementsMap: truncateTableStatementsMap,
	}, nil
}

func (c *clisync) getDestinationTableConstraints(schemas []string) (*sql_manager.TableConstraints, error) {
	cctx, cancel := context.WithDeadline(c.ctx, time.Now().Add(5*time.Second))
	defer cancel()
	db, err := c.sqlmanagerclient.NewSqlDbFromUrl(cctx, string(c.cmd.Destination.Driver), c.cmd.Destination.ConnectionUrl)
	if err != nil {
		return nil, err
	}
	defer db.Db.Close()

	constraints, err := db.Db.GetTableConstraintsBySchema(cctx, schemas)
	if err != nil {
		return nil, err
	}

	return constraints, nil
}
