package sync_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
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
	cli_neosync_benthos "github.com/nucleuscloud/neosync/cli/internal/benthos"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/nucleuscloud/neosync/cli/internal/version"
	connectiontunnelmanager "github.com/nucleuscloud/neosync/internal/connection-tunnel-manager"
	pool_sql_provider "github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/pool/providers/sql"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers/mongoprovider"
	"github.com/nucleuscloud/neosync/internal/connection-tunnel-manager/providers/sqlprovider"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	_ "github.com/nucleuscloud/neosync/cli/internal/benthos/inputs"
	benthos_environment "github.com/nucleuscloud/neosync/worker/pkg/benthos/environment"
	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/sql"
	"github.com/warpstreamlabs/bento/public/bloblang"
	_ "github.com/warpstreamlabs/bento/public/components/aws"
	_ "github.com/warpstreamlabs/bento/public/components/io"
	_ "github.com/warpstreamlabs/bento/public/components/pure"
	_ "github.com/warpstreamlabs/bento/public/components/pure/extended"

	http_client "github.com/nucleuscloud/neosync/worker/pkg/http/client"

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
	ConnectionUrl        string           `yaml:"connection-url"`
	Driver               DriverType       `yaml:"driver"`
	InitSchema           bool             `yaml:"init-schema,omitempty"`
	TruncateBeforeInsert bool             `yaml:"truncate-before-insert,omitempty"`
	TruncateCascade      bool             `yaml:"truncate-cascade,omitempty"`
	OnConflict           onConflictConfig `yaml:"on-conflict,omitempty"`
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

			config := &cmdConfig{
				Source: &sourceConfig{
					ConnectionOpts: &connectionOpts{},
				},
				Destination:            &sqlDestinationConfig{},
				AwsDynamoDbDestination: &dynamoDbDestinationConfig{},
			}
			configPath, err := cmd.Flags().GetString("config")
			if err != nil {
				return err
			}

			if configPath != "" {
				fileBytes, err := os.ReadFile(configPath)
				if err != nil {
					return fmt.Errorf("error reading config file: %w", err)
				}
				err = yaml.Unmarshal(fileBytes, &config)
				if err != nil {
					return fmt.Errorf("error parsing config file: %w", err)
				}
			}

			connectionId, err := cmd.Flags().GetString("connection-id")
			if err != nil {
				return err
			}
			if connectionId != "" {
				config.Source.ConnectionId = connectionId
			}

			destConnUrl, err := cmd.Flags().GetString("destination-connection-url")
			if err != nil {
				return err
			}
			if destConnUrl != "" {
				config.Destination.ConnectionUrl = destConnUrl
			}

			driver, err := cmd.Flags().GetString("destination-driver")
			if err != nil {
				return err
			}
			pDriver, ok := parseDriverString(driver)
			if ok {
				config.Destination.Driver = pDriver
			}

			initSchema, err := cmd.Flags().GetBool("init-schema")
			if err != nil {
				return err
			}
			if initSchema {
				config.Destination.InitSchema = initSchema
			}

			truncateBeforeInsert, err := cmd.Flags().GetBool("truncate-before-insert")
			if err != nil {
				return err
			}
			if truncateBeforeInsert {
				config.Destination.TruncateBeforeInsert = truncateBeforeInsert
			}

			truncateCascade, err := cmd.Flags().GetBool("truncate-cascade")
			if err != nil {
				return err
			}
			if truncateCascade {
				config.Destination.TruncateCascade = truncateCascade
			}

			onConflictDoNothing, err := cmd.Flags().GetBool("on-conflict-do-nothing")
			if err != nil {
				return err
			}
			if onConflictDoNothing {
				config.Destination.OnConflict.DoNothing = onConflictDoNothing
			}

			jobId, err := cmd.Flags().GetString("job-id")
			if err != nil {
				return err
			}
			if jobId != "" {
				config.Source.ConnectionOpts.JobId = &jobId
			}

			jobRunId, err := cmd.Flags().GetString("job-run-id")
			if err != nil {
				return err
			}
			if jobRunId != "" {
				config.Source.ConnectionOpts.JobRunId = &jobRunId
			}

			config, err = buildAwsCredConfig(cmd, config)
			if err != nil {
				return err
			}

			if config.Source.ConnectionId == "" {
				return fmt.Errorf("must provide connection-id")
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}

			outputType, err := output.ValidateAndRetrieveOutputFlag(cmd)
			if err != nil {
				return err
			}

			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				return err
			}
			config.Debug = debug

			return sync(cmd.Context(), outputType, apiKey, &accountId, config)
		},
	}

	cmd.Flags().Bool("debug", false, "Run in debug mode")

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

func sync(
	ctx context.Context,
	outputType output.OutputType,
	apiKey, accountIdFlag *string,
	cmd *cmdConfig,
) error {
	logLevel := charmlog.InfoLevel
	if cmd.Debug {
		logLevel = charmlog.DebugLevel
	}
	logger := charmlog.NewWithOptions(os.Stderr, charmlog.Options{
		ReportTimestamp: true,
		Level:           logLevel,
	})
	logger.Info("Starting sync")
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	httpclient := http_client.NewWithHeaders(version.Get().Headers())
	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpclient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	connectiondataclient := mgmtv1alpha1connect.NewConnectionDataServiceClient(
		httpclient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	pgpoolmap := &syncmap.Map{}
	mysqlpoolmap := &syncmap.Map{}
	mssqlpoolmap := &syncmap.Map{}
	pgquerier := pg_queries.New()
	mysqlquerier := mysql_queries.New()
	mssqlquerier := mssql_queries.New()
	sqlConnector := &sqlconnect.SqlOpenConnector{}
	sqlmanagerclient := sqlmanager.NewSqlManager(pgpoolmap, pgquerier, mysqlpoolmap, mysqlquerier, mssqlpoolmap, mssqlquerier, sqlConnector)

	logger.Debug("Retrieving neosync source connection")
	connResp, err := connectionclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: cmd.Source.ConnectionId,
	}))
	if err != nil {
		return err
	}
	sourceConnection := connResp.Msg.GetConnection()
	sourceConnectionType, err := getConnectionType(sourceConnection)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Source connection type: %s", sourceConnectionType))

	if sourceConnectionType == awsS3Connection && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("S3 source connection type requires job-id or job-run-id.")
	}
	if sourceConnectionType == gcpCloudStorageConnection && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("GCP Cloud Storage source connection type requires job-id or job-run-id")
	}

	if cmd.Destination.TruncateCascade && cmd.Destination.Driver == mysqlDriver {
		return fmt.Errorf("truncate cascade is only supported in postgres")
	}

	if sourceConnectionType == mysqlConnection || sourceConnectionType == postgresConnection {
		if cmd.Destination.Driver == "" {
			return fmt.Errorf("must provide destination-driver")
		}
		if cmd.Destination.ConnectionUrl == "" {
			return fmt.Errorf("must provide destination-connection-url")
		}

		if cmd.Destination.Driver != mysqlDriver && cmd.Destination.Driver != postgresDriver {
			return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
		}
	}

	if sourceConnectionType == awsDynamoDBConnection {
		if cmd.AwsDynamoDbDestination == nil {
			return fmt.Errorf("must provide destination aws credentials")
		}

		if cmd.AwsDynamoDbDestination.AwsCredConfig.Region == "" {
			return fmt.Errorf("must provide destination aws region")
		}
	}
	logger.Debug("Validated config")

	var token *string
	if isAuthEnabled {
		logger.Debug("Auth Enabled")
		if apiKey != nil && *apiKey != "" {
			logger.Debug("found API Key")
			token = apiKey
		} else {
			logger.Debug("Retrieving Access Token")
			accessToken, err := userconfig.GetAccessToken()
			if err != nil {
				logger.Error("Unable to retrieve access token. Please use neosync login command and try again.")
				return err
			}
			token = &accessToken
			logger.Debug("Setting account id")
			var accountId = accountIdFlag
			if accountId == nil || *accountId == "" {
				aId, err := userconfig.GetAccountId()
				if err != nil {
					logger.Error("Unable to retrieve account id. Please use account switch command to set account.")
					return err
				}
				accountId = &aId
			}

			if accountId == nil || *accountId == "" {
				return errors.New("Account Id not found. Please use account switch command to set account.")
			}

			if sourceConnection.AccountId != *accountId {
				return fmt.Errorf("Connection not found. AccountId: %s", *accountId)
			}
		}
	}

	logger.Debug("Checking if source and destination are compatible")
	err = areSourceAndDestCompatible(sourceConnection, cmd.Destination.Driver)
	if err != nil {
		return err
	}

	connectionprovider := providers.NewProvider(
		mongoprovider.NewProvider(),
		sqlprovider.NewProvider(sqlConnector),
	)
	tunnelmanager := connectiontunnelmanager.NewConnectionTunnelManager(connectionprovider)
	session := uuid.NewString()
	// might not need this in cli context
	defer func() {
		tunnelmanager.ReleaseSession(session)
	}()

	destConnection := cmdConfigToDestinationConnection(cmd)
	dsnToConnIdMap := &syncmap.Map{}
	var sqlDsn string
	if cmd.Destination != nil {
		sqlDsn = cmd.Destination.ConnectionUrl
	}
	dsnToConnIdMap.Store(sqlDsn, destConnection.Id)
	stopChan := make(chan error, 3)
	ctx, cancel := context.WithCancel(ctx)
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
	benv, err := benthos_environment.NewEnvironment(
		slog.Default(),
		benthos_environment.WithSqlConfig(&benthos_environment.SqlConfig{
			Provider: pool_sql_provider.NewProvider(pool_sql_provider.GetSqlPoolProviderGetter(
				tunnelmanager,
				dsnToConnIdMap,
				map[string]*mgmtv1alpha1.Connection{
					destConnection.Id: destConnection,
				},
				session,
				slog.Default(),
			)),
			IsRetry: false,
		}),
		benthos_environment.WithStopChannel(stopChan),
		benthos_environment.WithBlobEnv(bloblang.NewEnvironment()),
	)
	if err != nil {
		return err
	}

	logger.Info("Retrieving connection schema...")
	var schemaConfig *schemaConfig
	switch sourceConnectionType {
	case awsS3Connection:
		var cfg *mgmtv1alpha1.AwsS3SchemaConfig
		if cmd.Source.ConnectionOpts.JobRunId != nil && *cmd.Source.ConnectionOpts.JobRunId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobRunId{JobRunId: *cmd.Source.ConnectionOpts.JobRunId}}
		} else if cmd.Source.ConnectionOpts.JobId != nil && *cmd.Source.ConnectionOpts.JobId != "" {
			cfg = &mgmtv1alpha1.AwsS3SchemaConfig{Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobId{JobId: *cmd.Source.ConnectionOpts.JobId}}
		}
		s3Config := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config{
				AwsS3Config: cfg,
			},
		}

		schemaCfg, err := getDestinationSchemaConfig(ctx, connectiondataclient, sqlmanagerclient, sourceConnection, cmd, s3Config, logger)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			logger.Warn("No tables found.")
			return nil
		}
		schemaConfig = schemaCfg
	case gcpCloudStorageConnection:
		var cfg *mgmtv1alpha1.GcpCloudStorageSchemaConfig
		if cmd.Source.ConnectionOpts.JobRunId != nil && *cmd.Source.ConnectionOpts.JobRunId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobRunId{JobRunId: *cmd.Source.ConnectionOpts.JobRunId}}
		} else if cmd.Source.ConnectionOpts.JobId != nil && *cmd.Source.ConnectionOpts.JobId != "" {
			cfg = &mgmtv1alpha1.GcpCloudStorageSchemaConfig{Id: &mgmtv1alpha1.GcpCloudStorageSchemaConfig_JobId{JobId: *cmd.Source.ConnectionOpts.JobId}}
		}

		gcpConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_GcpCloudstorageConfig{
				GcpCloudstorageConfig: cfg,
			},
		}

		schemaCfg, err := getDestinationSchemaConfig(ctx, connectiondataclient, sqlmanagerclient, sourceConnection, cmd, gcpConfig, logger)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			logger.Warn("No tables found.")
			return nil
		}
		schemaConfig = schemaCfg
	case mysqlConnection:
		logger.Info("Building schema and table constraints...")
		mysqlCfg := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlSchemaConfig{},
			},
		}
		schemaCfg, err := getConnectionSchemaConfig(ctx, logger, connectiondataclient, sourceConnection, cmd, mysqlCfg)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			logger.Warn("No tables found.")
			return nil
		}
		schemaConfig = schemaCfg
	case postgresConnection:
		logger.Info("Building schema and table constraints...")
		postgresConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresSchemaConfig{},
			},
		}
		schemaCfg, err := getConnectionSchemaConfig(ctx, logger, connectiondataclient, sourceConnection, cmd, postgresConfig)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			logger.Warn("No tables found.")
			return nil
		}
		schemaConfig = schemaCfg
	case awsDynamoDBConnection:
		dynamoConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_DynamodbConfig{
				DynamodbConfig: &mgmtv1alpha1.DynamoDBSchemaConfig{},
			},
		}
		schemaCfg, err := getConnectionSchemaConfig(ctx, logger, connectiondataclient, sourceConnection, cmd, dynamoConfig)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			logger.Warn("No tables found.")
			return nil
		}
		tableMap := map[string]struct{}{}
		for _, s := range schemaCfg.Schemas {
			tableMap[s.Table] = struct{}{}
		}
		configs := []*benthosConfigResponse{}
		for t := range tableMap {
			benthosConfig := generateDynamoDbBenthosConfig(cmd, serverconfig.GetApiBaseUrl(), token, t)
			configs = append(configs, benthosConfig)
		}

		return runSync(ctx, outputType, benv, [][]*benthosConfigResponse{configs}, logger)
	default:
		return fmt.Errorf("this connection type is not currently supported")
	}

	logger.Debug("Building sync configs")
	syncConfigs := buildSyncConfigs(schemaConfig, logger)
	if syncConfigs == nil {
		return nil
	}
	logger.Info("Running table init statements...")
	err = runDestinationInitStatements(ctx, logger, sqlmanagerclient, cmd, syncConfigs, schemaConfig)
	if err != nil {
		return err
	}

	syncConfigCount := len(syncConfigs)
	logger.Infof("Generating %d sync configs...", syncConfigCount)
	configs := []*benthosConfigResponse{}
	for _, cfg := range syncConfigs {
		benthosConfig := generateBenthosConfig(cmd, sourceConnectionType, serverconfig.GetApiBaseUrl(), cfg, token)
		configs = append(configs, benthosConfig)
	}

	// order configs in run order by dependency
	groupedConfigs := groupConfigsByDependency(configs, logger)
	if groupedConfigs == nil {
		return nil
	}

	return runSync(ctx, outputType, benv, groupedConfigs, logger)
}

func areSourceAndDestCompatible(connection *mgmtv1alpha1.Connection, destinationDriver DriverType) error {
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if destinationDriver != postgresDriver {
			return fmt.Errorf("Connection and destination types are incompatible [postgres, %s]", destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if destinationDriver != mysqlDriver {
			return fmt.Errorf("Connection and destination types are incompatible [mysql, %s]", destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config, *mgmtv1alpha1.ConnectionConfig_GcpCloudstorageConfig, *mgmtv1alpha1.ConnectionConfig_DynamodbConfig:
	default:
		return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
	}
	return nil
}

var (
	// Hack that locks the instanced bento stream builder build step that causes data races if done in parallel
	streamBuilderMu syncmap.Mutex
)

func syncData(ctx context.Context, benv *service.Environment, cfg *benthosConfigResponse, logger *charmlog.Logger, outputType output.OutputType) error {
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
	if outputType == output.PlainOutput {
		streambldr.SetPrintLogger(logger.With("benthos", "true", "table", cfg.Table, "runType", runType).StandardLog())
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

func runDestinationInitStatements(
	ctx context.Context,
	logger *charmlog.Logger,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	cmd *cmdConfig,
	syncConfigs []*tabledependency.RunConfig,
	schemaConfig *schemaConfig,
) error {
	dependencyMap := buildDependencyMap(syncConfigs)
	db, err := sqlmanagerclient.NewSqlDbFromUrl(ctx, string(cmd.Destination.Driver), cmd.Destination.ConnectionUrl)
	if err != nil {
		return err
	}
	defer db.Db.Close()
	if cmd.Destination.InitSchema {
		if len(schemaConfig.InitSchemaStatements) != 0 {
			for _, block := range schemaConfig.InitSchemaStatements {
				logger.Infof("[%s] found %d statements to execute during schema initialization", block.Label, len(block.Statements))
				if len(block.Statements) == 0 {
					continue
				}
				err = db.Db.BatchExec(ctx, batchSize, block.Statements, &sql_manager.BatchExecOpts{})
				if err != nil {
					logger.Error("Error creating tables:", err)
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

			err = db.Db.BatchExec(ctx, batchSize, orderedInitStatements, &sql_manager.BatchExecOpts{})
			if err != nil {
				logger.Error("Error creating tables:", err)
				return err
			}
		}
	}
	if cmd.Destination.Driver == postgresDriver {
		if cmd.Destination.TruncateCascade {
			truncateCascadeStmts := []string{}
			for _, syncCfg := range syncConfigs {
				stmt, ok := schemaConfig.TruncateTableStatementsMap[syncCfg.Table()]
				if ok {
					truncateCascadeStmts = append(truncateCascadeStmts, stmt)
				}
			}
			err = db.Db.BatchExec(ctx, batchSize, truncateCascadeStmts, &sql_manager.BatchExecOpts{})
			if err != nil {
				logger.Error("Error truncate cascade tables:", err)
				return err
			}
		} else if cmd.Destination.TruncateBeforeInsert {
			orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(dependencyMap)
			if err != nil {
				return err
			}
			orderedTruncateStatement, err := sqlmanager_postgres.BuildPgTruncateStatement(orderedTablesResp.OrderedTables)
			if err != nil {
				return err
			}
			err = db.Db.Exec(ctx, orderedTruncateStatement)
			if err != nil {
				logger.Error("Error truncating tables:", err)
				return err
			}
		}
	} else if cmd.Destination.Driver == mysqlDriver {
		orderedTablesResp, err := tabledependency.GetTablesOrderedByDependency(dependencyMap)
		if err != nil {
			return err
		}
		orderedTableTruncateStatements := []string{}
		for _, t := range orderedTablesResp.OrderedTables {
			orderedTableTruncateStatements = append(orderedTableTruncateStatements, schemaConfig.TruncateTableStatementsMap[t.String()])
		}
		disableFkChecks := sql_manager.DisableForeignKeyChecks
		err = db.Db.BatchExec(ctx, batchSize, orderedTableTruncateStatements, &sql_manager.BatchExecOpts{Prefix: &disableFkChecks})
		if err != nil {
			logger.Error("Error truncating tables:", err)
			return err
		}
	}
	return nil
}

func buildSyncConfigs(
	schemaConfig *schemaConfig,
	logger *charmlog.Logger,
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

func buildDependencyMap(syncConfigs []*tabledependency.RunConfig) map[string][]string {
	dependencyMap := map[string][]string{}
	for _, cfg := range syncConfigs {
		_, dpOk := dependencyMap[cfg.Table()]
		if !dpOk {
			dependencyMap[cfg.Table()] = []string{}
		}

		for _, dep := range cfg.DependsOn() {
			dependencyMap[cfg.Table()] = append(dependencyMap[cfg.Table()], dep.Table)
		}
	}
	return dependencyMap
}

func getTableInitStatementMap(
	ctx context.Context,
	logger *charmlog.Logger,
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

type SqlTable struct {
	Schema  string
	Table   string
	Columns []string
}

func getTableColMap(schemas []*mgmtv1alpha1.DatabaseColumn) map[string][]string {
	tableColMap := map[string][]string{}
	for _, record := range schemas {
		table := sql_manager.BuildTable(record.Schema, record.Table)
		_, ok := tableColMap[table]
		if ok {
			tableColMap[table] = append(tableColMap[table], record.Column)
		} else {
			tableColMap[table] = []string{record.Column}
		}
	}

	return tableColMap
}

type benthosConfigResponse struct {
	Name      string
	DependsOn []*tabledependency.DependsOn
	Config    *cli_neosync_benthos.BenthosConfig
	Table     string
	Columns   []string
}

func generateBenthosConfig(
	cmd *cmdConfig,
	connectionType ConnectionType,
	apiUrl string,
	syncConfig *tabledependency.RunConfig,
	authToken *string,
) *benthosConfigResponse {
	schema, table := sqlmanager_shared.SplitTableKey(syncConfig.Table())

	var jobId, jobRunId *string
	if cmd.Source.ConnectionOpts != nil {
		jobRunId = cmd.Source.ConnectionOpts.JobRunId
		jobId = cmd.Source.ConnectionOpts.JobId
	}

	bc := &cli_neosync_benthos.BenthosConfig{
		StreamConfig: cli_neosync_benthos.StreamConfig{
			Logger: &cli_neosync_benthos.LoggerConfig{
				Level:        "ERROR",
				AddTimestamp: true,
			},
			Input: &cli_neosync_benthos.InputConfig{
				Inputs: cli_neosync_benthos.Inputs{
					NeosyncConnectionData: &cli_neosync_benthos.NeosyncConnectionData{
						ApiKey:         authToken,
						ApiUrl:         apiUrl,
						ConnectionId:   cmd.Source.ConnectionId,
						ConnectionType: string(connectionType),
						JobId:          jobId,
						JobRunId:       jobRunId,
						Schema:         schema,
						Table:          table,
					},
				},
			},
			Pipeline: &cli_neosync_benthos.PipelineConfig{},
			Output:   &cli_neosync_benthos.OutputConfig{},
		},
	}

	if syncConfig.RunType() == tabledependency.RunTypeUpdate {
		args := syncConfig.InsertColumns()
		args = append(args, syncConfig.PrimaryKeys()...)
		bc.Output = &cli_neosync_benthos.OutputConfig{
			Outputs: cli_neosync_benthos.Outputs{
				PooledSqlUpdate: &cli_neosync_benthos.PooledSqlUpdate{
					Driver: string(cmd.Destination.Driver),
					Dsn:    cmd.Destination.ConnectionUrl,

					Schema:       schema,
					Table:        table,
					Columns:      syncConfig.InsertColumns(),
					WhereColumns: syncConfig.PrimaryKeys(),
					ArgsMapping:  buildPlainInsertArgs(args),

					Batching: &cli_neosync_benthos.Batching{
						Period: "5s",
						Count:  100,
					},
				},
			},
		}
	} else {
		bc.Output = &cli_neosync_benthos.OutputConfig{
			Outputs: cli_neosync_benthos.Outputs{
				PooledSqlInsert: &cli_neosync_benthos.PooledSqlInsert{
					Driver: string(cmd.Destination.Driver),
					Dsn:    cmd.Destination.ConnectionUrl,

					Schema:              schema,
					Table:               table,
					Columns:             syncConfig.SelectColumns(),
					OnConflictDoNothing: cmd.Destination.OnConflict.DoNothing,
					ArgsMapping:         buildPlainInsertArgs(syncConfig.SelectColumns()),

					Batching: &cli_neosync_benthos.Batching{
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
func groupConfigsByDependency(configs []*benthosConfigResponse, logger *charmlog.Logger) [][]*benthosConfigResponse {
	groupedConfigs := [][]*benthosConfigResponse{}
	configMap := map[string]*benthosConfigResponse{}
	queuedMap := map[string][]string{} // map -> table to cols

	// get root configs
	rootConfigs := []*benthosConfigResponse{}
	for _, c := range configs {
		if len(c.DependsOn) == 0 {
			rootConfigs = append(rootConfigs, c)
			queuedMap[c.Table] = c.Columns
		} else {
			configMap[c.Name] = c
		}
	}
	if len(rootConfigs) == 0 {
		logger.Info("No root configs found. There must be one config with no dependencies.")
		return nil
	}
	groupedConfigs = append(groupedConfigs, rootConfigs)

	prevTableLen := 0
	for len(configMap) > 0 {
		// prevents looping forever
		if prevTableLen == len(configMap) {
			logger.Info("Unable to order configs by dependency. No path found.")
			return nil
		}
		prevTableLen = len(configMap)
		dependentConfigs := []*benthosConfigResponse{}
		for _, c := range configMap {
			if isConfigReady(c, queuedMap) {
				dependentConfigs = append(dependentConfigs, c)
				delete(configMap, c.Name)
			}
		}
		if len(dependentConfigs) > 0 {
			groupedConfigs = append(groupedConfigs, dependentConfigs)
			for _, c := range dependentConfigs {
				queuedMap[c.Table] = append(queuedMap[c.Table], c.Columns...)
			}
		}
	}

	return groupedConfigs
}
func isConfigReady(config *benthosConfigResponse, queuedMap map[string][]string) bool {
	for _, dep := range config.DependsOn {
		if cols, ok := queuedMap[dep.Table]; ok {
			for _, dc := range dep.Columns {
				if !slices.Contains(cols, dc) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

type schemaConfig struct {
	Schemas                    []*mgmtv1alpha1.DatabaseColumn
	TableConstraints           map[string][]*sql_manager.ForeignConstraint
	TablePrimaryKeys           map[string]*mgmtv1alpha1.PrimaryConstraint
	InitTableStatementsMap     map[string]string
	TruncateTableStatementsMap map[string]string
	InitSchemaStatements       []*mgmtv1alpha1.SchemaInitStatements
}

func getConnectionSchemaConfig(
	ctx context.Context,
	logger *charmlog.Logger,
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	connection *mgmtv1alpha1.Connection,
	cmd *cmdConfig,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
) (*schemaConfig, error) {
	var schemas []*mgmtv1alpha1.DatabaseColumn
	var tableConstraints map[string]*mgmtv1alpha1.ForeignConstraintTables
	var tablePrimaryKeys map[string]*mgmtv1alpha1.PrimaryConstraint
	var initTableStatementsMap map[string]string
	var truncateTableStatementsMap map[string]string
	var initSchemaStatements []*mgmtv1alpha1.SchemaInitStatements
	errgrp, errctx := errgroup.WithContext(ctx)
	errgrp.Go(func() error {
		schemaResp, err := connectiondataclient.GetConnectionSchema(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
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
		constraintConnectionResp, err := connectiondataclient.GetConnectionTableConstraints(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionTableConstraintsRequest{ConnectionId: cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tableConstraints = constraintConnectionResp.Msg.GetForeignKeyConstraints()
		tablePrimaryKeys = constraintConnectionResp.Msg.GetPrimaryKeyConstraints()
		return nil
	})

	errgrp.Go(func() error {
		initStatementsResp, err := getTableInitStatementMap(errctx, logger, connectiondataclient, cmd.Source.ConnectionId, cmd.Destination)
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

func getDestinationSchemaConfig(
	ctx context.Context,
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	sqlmanagerclient sqlmanager.SqlManagerClient,
	connection *mgmtv1alpha1.Connection,
	cmd *cmdConfig,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
	logger *charmlog.Logger,
) (*schemaConfig, error) {
	schemaResp, err := connectiondataclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		ConnectionId: connection.Id,
		SchemaConfig: sc,
	}))
	if err != nil {
		return nil, err
	}

	tableColMap := getTableColMap(schemaResp.Msg.GetSchemas())
	if len(tableColMap) == 0 {
		logger.Info("No tables found.")
		return nil, nil
	}

	schemaMap := map[string]struct{}{}
	for _, s := range schemaResp.Msg.GetSchemas() {
		schemaMap[s.Schema] = struct{}{}
	}
	schemas := []string{}
	for s := range schemaMap {
		schemas = append(schemas, s)
	}

	logger.Info("Building table constraints...")
	tableConstraints, err := getDestinationTableConstraints(ctx, sqlmanagerclient, cmd.Destination.Driver, cmd.Destination.ConnectionUrl, schemas)
	if err != nil {
		return nil, err
	}

	primaryKeys := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for tableName, cols := range tableConstraints.PrimaryKeyConstraints {
		primaryKeys[tableName] = &mgmtv1alpha1.PrimaryConstraint{
			Columns: cols,
		}
	}

	truncateTableStatementsMap := map[string]string{}
	if cmd.Destination.Driver == postgresDriver {
		if cmd.Destination.TruncateCascade {
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
		if cmd.Destination.TruncateBeforeInsert {
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

func getDestinationTableConstraints(ctx context.Context, sqlmanagerclient sqlmanager.SqlManagerClient, connectionDriver DriverType, connectionUrl string, schemas []string) (*sql_manager.TableConstraints, error) {
	cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
	defer cancel()
	db, err := sqlmanagerclient.NewSqlDbFromUrl(cctx, string(connectionDriver), connectionUrl)
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

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx := range cols {
		pieces[idx] = fmt.Sprintf("this.%q", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parseDriverString(str string) (DriverType, bool) {
	p, ok := driverMap[strings.ToLower(str)]
	return p, ok
}

func getConnectionType(connection *mgmtv1alpha1.Connection) (ConnectionType, error) {
	if connection.ConnectionConfig.GetAwsS3Config() != nil {
		return awsS3Connection, nil
	}
	if connection.GetConnectionConfig().GetGcpCloudstorageConfig() != nil {
		return gcpCloudStorageConnection, nil
	}
	if connection.ConnectionConfig.GetMysqlConfig() != nil {
		return mysqlConnection, nil
	}
	if connection.ConnectionConfig.GetPgConfig() != nil {
		return postgresConnection, nil
	}
	if connection.ConnectionConfig.GetDynamodbConfig() != nil {
		return awsDynamoDBConnection, nil
	}
	return "", errors.New("unsupported connection type")
}
