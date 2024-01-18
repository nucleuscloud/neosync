package sync_cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	neosync_benthos "github.com/nucleuscloud/neosync/cli/internal/benthos"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/output"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/nucleuscloud/neosync/cli/internal/userconfig"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	_ "github.com/nucleuscloud/neosync/cli/internal/benthos/inputs"

	"github.com/benthosdev/benthos/v4/public/service"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	maxPgParamLimit = 65535

	postgresDriver DriverType = "postgres"
	mysqlDriver    DriverType = "mysql"

	awsS3Connection    ConnectionType = "awsS3"
	postgresConnection ConnectionType = "postgres"
	mysqlConnection    ConnectionType = "mysql"
)

var (
	driverMap = map[string]DriverType{
		string(postgresDriver): postgresDriver,
		string(mysqlDriver):    mysqlDriver,
	}
)

type ConnectionType string
type DriverType string

type model struct {
	ctx            context.Context
	groupedConfigs [][]*benthosConfigResponse
	tableSynced    int
	index          int
	width          int
	height         int
	spinner        spinner.Model
	progress       progress.Model
	done           bool
}

var (
	bold                = lipgloss.NewStyle().PaddingLeft(2).Bold(true)
	header              = lipgloss.NewStyle().Faint(true).PaddingLeft(2)
	printlog            = lipgloss.NewStyle().PaddingLeft(2)
	currentPkgNameStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("211"))
	doneStyle           = lipgloss.NewStyle().Margin(1, 2)
	checkMark           = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("42")).SetString("✓")
)

type cmdConfig struct {
	Source      *sourceConfig      `yaml:"source"`
	Destination *destinationConfig `yaml:"destination"`
}

type sourceConfig struct {
	ConnectionId   string          `yaml:"connection-id"`
	ConnectionOpts *connectionOpts `yaml:"connection-opts,omitempty"`
}

type connectionOpts struct {
	JobId    *string `yaml:"job-id,omitempty"`
	JobRunId *string `yaml:"job-run-id,omitempty"`
}

type destinationConfig struct {
	ConnectionUrl        string     `yaml:"connection-url"`
	Driver               DriverType `yaml:"driver"`
	InitSchema           bool       `yaml:"init-schema,omitempty"`
	TruncateBeforeInsert bool       `yaml:"truncate-before-insert,omitempty"`
	TruncateCascade      bool       `yaml:"truncate-cascade,omitempty"`
}

type syncConfig struct {
	Query         string
	ArgsMapping   string
	InitStatement string
	DependsOn     []*tabledependency.DependsOn
	Schema        string
	Table         string
	Columns       []string
	Name          string
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
				Destination: &destinationConfig{},
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
			if truncateBeforeInsert {
				config.Destination.TruncateCascade = truncateCascade
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

			if config.Source.ConnectionId == "" {
				return fmt.Errorf("must provide connection-id")
			}
			if config.Destination.Driver == "" {
				return fmt.Errorf("must provide destination-driver")
			}
			if config.Destination.ConnectionUrl == "" {
				return fmt.Errorf("must provide destination-connection-url")
			}

			if config.Destination.TruncateCascade && config.Destination.Driver != postgresDriver {
				return fmt.Errorf("wrong driver type. truncate cascade is only supported in postgres")
			}

			if config.Destination.Driver != mysqlDriver && config.Destination.Driver != postgresDriver {
				return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}

			outputType, err := output.ValidateAndRetrieveOutputFlag(cmd)
			if err != nil {
				return err
			}

			return sync(cmd.Context(), outputType, apiKey, &accountId, config)
		},
	}

	cmd.Flags().String("connection-id", "", "Connection id for sync source")
	cmd.Flags().String("job-id", "", "Id of Job to sync data from. Only used with AWS S3 connections. Can use job-run-id instead.")
	cmd.Flags().String("job-run-id", "", "Id of Job run to sync data from. Only used with AWS S3 connections. Can use job-id instead.")
	cmd.Flags().String("destination-connection-url", "", "Connection url for sync output")
	cmd.Flags().String("destination-driver", "", "Connection driver for sync output")
	cmd.Flags().String("account-id", "", "Account source connection is in. Defaults to account id in cli context")
	cmd.Flags().String("config", "", "Location of config file")
	cmd.Flags().Bool("init-schema", false, "Create table schema and its constraints")
	cmd.Flags().Bool("truncate-before-insert", false, "Truncate table before insert")
	cmd.Flags().Bool("truncate-cascade", false, "Truncate cascade table before insert (postgres only)")
	output.AttachOutputFlag(cmd)

	return cmd
}

func sync(
	ctx context.Context,
	outputType output.OutputType,
	apiKey, accountIdFlag *string,
	cmd *cmdConfig,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	connectiondataclient := mgmtv1alpha1connect.NewConnectionDataServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	connResp, err := connectionclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: cmd.Source.ConnectionId,
	}))
	if err != nil {
		return err
	}
	connection := connResp.Msg.GetConnection()
	connectionType, err := getConnectionType(connection)
	if err != nil {
		return err
	}

	if connectionType == awsS3Connection && (cmd.Source.ConnectionOpts.JobId == nil || *cmd.Source.ConnectionOpts.JobId == "") && (cmd.Source.ConnectionOpts.JobRunId == nil || *cmd.Source.ConnectionOpts.JobRunId == "") {
		return errors.New("S3 source connection type requires job-id or job-run-id.")
	}

	var token *string
	if isAuthEnabled {
		if apiKey != nil && *apiKey != "" {
			token = apiKey
		} else {
			accessToken, err := userconfig.GetAccessToken()
			if err != nil {
				fmt.Println("Unable to retrieve access token. Please use neosync login command and try again.") // nolint
				return err
			}
			token = &accessToken
			var accountId = accountIdFlag
			if accountId == nil || *accountId == "" {
				aId, err := userconfig.GetAccountId()
				if err != nil {
					fmt.Println("Unable to retrieve account id. Please use account switch command to set account.") // nolint
					return err
				}
				accountId = &aId
			}

			if accountId == nil || *accountId == "" {
				return errors.New("Account Id not found. Please use account switch command to set account.")
			}

			if connection.AccountId != *accountId {
				return errors.New(fmt.Sprintf("Connection not found. AccountId: %s", *accountId)) // nolint
			}
		}
	}

	err = areSourceAndDestCompatible(connection, cmd.Destination.Driver)
	if err != nil {
		return err
	}

	fmt.Println(header.Render("\n── Preparing ─────────────────────────────────────")) // nolint
	fmt.Println(printlog.Render("Retrieving connection schema..."))                    // nolint

	syncConfigs := []*syncConfig{}
	switch connectionType {
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

		schemaCfg, err := getDestinationSchemaConfig(ctx, connectiondataclient, connection, cmd, s3Config)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		if cmd.Destination.Driver == postgresDriver {
			configs := buildSyncConfigs(schemaCfg, buildPostgresInsertQuery, buildPostgresUpdateQuery)
			if configs == nil {
				return nil
			}
			syncConfigs = append(syncConfigs, configs...)
		} else if cmd.Destination.Driver == mysqlDriver {
			configs := buildSyncConfigs(schemaCfg, buildMysqlInsertQuery, buildMysqlUpdateQuery)
			if configs == nil {
				return nil
			}
			syncConfigs = append(syncConfigs, configs...)
		}

	case mysqlConnection:
		fmt.Println(printlog.Render("Building schema and table constraints...")) // nolint
		mysqlCfg := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig{
				MysqlConfig: &mgmtv1alpha1.MysqlSchemaConfig{},
			},
		}
		schemaCfg, err := getConnectionSchemaConfig(ctx, connectiondataclient, connection, cmd, mysqlCfg)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		configs := buildSyncConfigs(schemaCfg, buildPostgresInsertQuery, buildPostgresUpdateQuery)
		if configs == nil {
			return nil
		}
		syncConfigs = append(syncConfigs, configs...)

	case postgresConnection:
		fmt.Println(printlog.Render("Building schema and table constraints...")) // nolint
		postgresConfig := &mgmtv1alpha1.ConnectionSchemaConfig{
			Config: &mgmtv1alpha1.ConnectionSchemaConfig_PgConfig{
				PgConfig: &mgmtv1alpha1.PostgresSchemaConfig{},
			},
		}
		schemaCfg, err := getConnectionSchemaConfig(ctx, connectiondataclient, connection, cmd, postgresConfig)
		if err != nil {
			return err
		}
		if len(schemaCfg.Schemas) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		configs := buildSyncConfigs(schemaCfg, buildPostgresInsertQuery, buildPostgresUpdateQuery)
		if configs == nil {
			return nil
		}
		syncConfigs = append(syncConfigs, configs...)

	default:
		return fmt.Errorf("this connection type is not currently supported")
	}

	fmt.Println(printlog.Render("Generating configs... \n")) // nolint
	configs := []*benthosConfigResponse{}
	for _, cfg := range syncConfigs {
		benthosConfig := generateBenthosConfig(cmd, connectionType, serverconfig.GetApiBaseUrl(), cfg, token)
		configs = append(configs, benthosConfig)
	}

	// order configs in run order by dependency
	groupedConfigs := groupConfigsByDependency(configs)
	if groupedConfigs == nil {
		return nil
	}

	var opts []tea.ProgramOption
	if outputType == output.PlainOutput {
		// Plain mode don't render the TUI
		opts = []tea.ProgramOption{tea.WithoutRenderer(), tea.WithInput(nil)}
	} else {
		// TUI mode, discard log output
		log.SetOutput(io.Discard)
	}
	fmt.Println(header.Render("── Syncing Tables ────────────────────────────────")) // nolint
	if _, err := tea.NewProgram(newModel(ctx, groupedConfigs), opts...).Run(); err != nil {
		fmt.Println("Error syncing data:", err) // nolint
		os.Exit(1)
	}

	return nil
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
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
	default:
		return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
	}
	return nil
}

func syncData(ctx context.Context, cfg *benthosConfigResponse) error {
	configbits, err := yaml.Marshal(cfg.Config)
	if err != nil {
		return err
	}

	var benthosStream *service.Stream
	go func() {
		for { // nolint
			select {
			case <-ctx.Done():
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.Stop(ctx)
					if err != nil {
						fmt.Println(err.Error()) // nolint
					}
				}
				return
			}
		}
	}()

	streambldr := service.NewStreamBuilder()
	// would ideally use the activity logger here but can't convert it into a slog.
	benthoslogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	streambldr.SetLogger(benthoslogger.With(
		"benthos", "true",
	))

	err = streambldr.SetYAML(string(configbits))
	if err != nil {
		return fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		return err
	}
	benthosStream = stream

	err = stream.Run(ctx)
	if err != nil {
		return fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStream = nil
	return nil
}

func buildSyncConfigs(
	schemaConfig *schemaConfig,
	buildInsertQueryFunc func(table string, cols []string) string,
	buildUpdateQueryFunc func(table string, col []string, primaryKeys []string) string,
) []*syncConfig {
	syncConfigs := []*syncConfig{}
	tableColMap := getTableColMap(schemaConfig.Schemas)
	if len(tableColMap) == 0 {
		return syncConfigs
	}
	dependencyMap := getDependencyConfigs(schemaConfig.TableConstraints, tableColMap)

	for table, dc := range dependencyMap {
		split := strings.Split(table, ".")
		cols := tableColMap[table]
		if len(dc) > 1 {
			for _, c := range dc {
				if c.Columns != nil && c.Columns.Exclude != nil && len(c.Columns.Exclude) > 0 {
					filteredCols := []string{}
					for _, col := range cols {
						if !slices.Contains(c.Columns.Exclude, col) {
							filteredCols = append(filteredCols, col)
						}
					}
					syncConfigs = append(syncConfigs, &syncConfig{
						Query:         buildInsertQueryFunc(table, filteredCols),
						ArgsMapping:   buildPlainInsertArgs(filteredCols),
						InitStatement: schemaConfig.InitTableStatementsMap[table],
						Schema:        split[0],
						Table:         split[1],
						Columns:       filteredCols,
						DependsOn:     c.DependsOn,
						Name:          table,
					})
				} else if c.Columns != nil && c.Columns.Include != nil && len(c.Columns.Include) > 0 {
					primaryKeyCols := schemaConfig.TablePrimaryKeys[table].GetColumns()
					if len(primaryKeyCols) == 0 {
						fmt.Println(bold.Render(fmt.Sprintf("No primary keys found for table (%s). Unable to build update query.", table))) // nolint
						return nil
					}
					argCols := []string{}
					argCols = append(argCols, c.Columns.Include...)
					argCols = append(argCols, primaryKeyCols...)
					syncConfigs = append(syncConfigs, &syncConfig{
						Query:       buildUpdateQueryFunc(table, c.Columns.Include, primaryKeyCols),
						ArgsMapping: buildPlainInsertArgs(argCols),
						Schema:      split[0],
						Table:       split[1],
						Columns:     c.Columns.Include,
						DependsOn:   c.DependsOn,
						Name:        fmt.Sprintf("%s.update", table),
					})
				}
			}
		} else {
			syncConfigs = append(syncConfigs, &syncConfig{
				Query:         buildInsertQueryFunc(table, cols),
				ArgsMapping:   buildPlainInsertArgs(cols),
				InitStatement: schemaConfig.InitTableStatementsMap[table],
				Schema:        split[0],
				Table:         split[1],
				Columns:       cols,
				DependsOn:     dc[0].DependsOn,
				Name:          table,
			})
		}
	}

	return syncConfigs
}

func getTableInitStatementMap(ctx context.Context, connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient, connectionId string, opts *destinationConfig) (map[string]string, error) {
	if opts.InitSchema || opts.TruncateBeforeInsert || opts.TruncateCascade {
		fmt.Println(printlog.Render("Creating init statements...")) // nolint
		initStatementResp, err := connectiondataclient.GetConnectionInitStatements(ctx,
			connect.NewRequest(&mgmtv1alpha1.GetConnectionInitStatementsRequest{
				ConnectionId: connectionId,
				Options: &mgmtv1alpha1.InitStatementOptions{
					InitSchema:           opts.InitSchema,
					TruncateBeforeInsert: opts.TruncateBeforeInsert,
					TruncateCascade:      opts.TruncateCascade,
				},
			},
			))
		if err != nil {
			return nil, err
		}
		return initStatementResp.Msg.TableInitStatements, nil
	}
	return map[string]string{}, nil
}

type SqlTable struct {
	Schema  string
	Table   string
	Columns []string
}

func getTableColMap(schemas []*mgmtv1alpha1.DatabaseColumn) map[string][]string {
	tableColMap := map[string][]string{}
	for _, record := range schemas {
		table := dbschemas_utils.BuildTable(record.Schema, record.Table)
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
	Config    *neosync_benthos.BenthosConfig
	Table     string
	Columns   []string
}

func generateBenthosConfig(
	cmd *cmdConfig,
	connectionType ConnectionType,
	apiUrl string,
	syncConfig *syncConfig,
	authToken *string,
) *benthosConfigResponse {
	tableName := dbschemas_utils.BuildTable(syncConfig.Schema, syncConfig.Table)

	var jobId, jobRunId *string
	if cmd.Source.ConnectionOpts != nil {
		jobRunId = cmd.Source.ConnectionOpts.JobRunId
		jobId = cmd.Source.ConnectionOpts.JobId
	}

	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
						ApiKey:         authToken,
						ApiUrl:         apiUrl,
						ConnectionId:   cmd.Source.ConnectionId,
						ConnectionType: string(connectionType),
						JobId:          jobId,
						JobRunId:       jobRunId,
						Schema:         syncConfig.Schema,
						Table:          syncConfig.Table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					SqlRaw: &neosync_benthos.SqlRaw{
						Driver: string(cmd.Destination.Driver),
						Dsn:    cmd.Destination.ConnectionUrl,

						Query:         syncConfig.Query,
						ArgsMapping:   syncConfig.ArgsMapping,
						InitStatement: syncConfig.InitStatement,

						Batching: &neosync_benthos.Batching{
							Period: "5s",
							// max allowed by postgres in a single batch
							Count: 100,
						},
					},
				},
			},
		},
	}

	return &benthosConfigResponse{
		Name:      syncConfig.Name,
		Config:    bc,
		DependsOn: syncConfig.DependsOn,
		Table:     tableName,
		Columns:   syncConfig.Columns,
	}
}
func groupConfigsByDependency(configs []*benthosConfigResponse) [][]*benthosConfigResponse {
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
		fmt.Println(bold.Render("No root configs found. There must be one config with no dependencies.")) // nolint
		return nil
	}
	groupedConfigs = append(groupedConfigs, rootConfigs)

	maxCount := len(configs)
	for len(configMap) > 0 && maxCount > 0 {
		dependentConfigs := []*benthosConfigResponse{}
		maxCount--
		for _, c := range configMap {
			if isConfigReady(c, queuedMap) {
				dependentConfigs = append(dependentConfigs, c)

				delete(configMap, c.Name)
			}
		}
		if len(dependentConfigs) > 0 {
			groupedConfigs = append(groupedConfigs, dependentConfigs)
			for _, c := range dependentConfigs {
				queuedMap[c.Table] = c.Columns
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

func newModel(ctx context.Context, groupedConfigs [][]*benthosConfigResponse) *model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	return &model{
		ctx:            ctx,
		groupedConfigs: groupedConfigs,
		tableSynced:    0,
		spinner:        s,
		progress:       p,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(syncConfigs(m.ctx, m.groupedConfigs[m.index]), m.spinner.Tick)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			return m, tea.Quit
		}
	case syncedDataMsg:
		totalConfigCount := getConfigCount(m.groupedConfigs)
		configCount := len(m.groupedConfigs)

		successStrs := []string{}
		for _, config := range m.groupedConfigs[m.index] {
			successStrs = append(successStrs, fmt.Sprintf("%s %s", checkMark, config.Name))
			m.tableSynced++
		}
		progressCmd := m.progress.SetPercent(float64(m.tableSynced) / float64(totalConfigCount))

		if m.index == configCount-1 {
			m.done = true
			log.Printf("Done! Completed %d tables.", configCount)
			return m, tea.Batch(
				progressCmd,
				tea.Println(strings.Join(successStrs, " \n")),
				tea.Quit,
			)
		}

		m.index++
		return m, tea.Batch(
			progressCmd,
			tea.Println(strings.Join(successStrs, " \n")),
			syncConfigs(m.ctx, m.groupedConfigs[m.index]),
		)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if newModel, ok := newModel.(progress.Model); ok {
			m.progress = newModel
		}
		return m, cmd
	}
	return m, nil
}

func (m *model) View() string {
	configCount := getConfigCount(m.groupedConfigs)
	w := lipgloss.Width(fmt.Sprintf("%d", configCount))

	if m.done {
		return doneStyle.Render(fmt.Sprintf("Done! Completed %d tables.\n", configCount))
	}

	pkgCount := fmt.Sprintf(" %*d/%*d", w, m.tableSynced, w, configCount)

	spin := m.spinner.View() + " "
	prog := m.progress.View()
	cellsAvail := max(0, m.width-lipgloss.Width(spin+prog+pkgCount))

	successStrs := []string{}
	for _, config := range m.groupedConfigs[m.index] {
		successStrs = append(successStrs, config.Name)
	}
	pkgName := currentPkgNameStyle.Render(successStrs...)
	info := lipgloss.NewStyle().MaxWidth(cellsAvail).Render("Syncing " + pkgName)

	cellsRemaining := max(0, m.width-lipgloss.Width(spin+info+prog+pkgCount))
	gap := strings.Repeat(" ", cellsRemaining)

	return spin + info + gap + prog + pkgCount
}

type syncedDataMsg string

func syncConfigs(ctx context.Context, configs []*benthosConfigResponse) tea.Cmd {
	return func() tea.Msg {
		errgrp, errctx := errgroup.WithContext(ctx)
		for _, cfg := range configs {
			cfg := cfg
			errgrp.Go(func() error {
				log.Printf("Syncing table %s \n", cfg.Name)
				err := syncData(errctx, cfg)
				if err != nil {
					fmt.Printf("Error syncing table: %s \n", err.Error()) // nolint
					return err
				}
				return nil
			})
		}

		if err := errgrp.Wait(); err != nil {
			tea.Printf("Error syncing data: %s \n", err.Error())
			return tea.Quit
		}

		message := ""
		for _, config := range configs {
			message = fmt.Sprintf("%s, %s", message, config.Name)
		}
		return syncedDataMsg(message)
	}
}

type schemaConfig struct {
	Schemas                []*mgmtv1alpha1.DatabaseColumn
	TableConstraints       map[string]*dbschemas_utils.TableConstraints
	TablePrimaryKeys       map[string]*mgmtv1alpha1.PrimaryConstraint
	InitTableStatementsMap map[string]string
}

func getConnectionSchemaConfig(
	ctx context.Context,
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	connection *mgmtv1alpha1.Connection,
	cmd *cmdConfig,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
) (*schemaConfig, error) {
	var schemas []*mgmtv1alpha1.DatabaseColumn
	var tableConstraints map[string]*mgmtv1alpha1.ForeignConstraintTables
	var tablePrimaryKeys map[string]*mgmtv1alpha1.PrimaryConstraint
	var initTableStatementsMap map[string]string
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
		fkConnectionResp, err := connectiondataclient.GetConnectionForeignConstraints(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tableConstraints = fkConnectionResp.Msg.GetTableConstraints()
		return nil
	})

	errgrp.Go(func() error {
		pkConnectionResp, err := connectiondataclient.GetConnectionPrimaryConstraints(errctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionPrimaryConstraintsRequest{ConnectionId: cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tablePrimaryKeys = pkConnectionResp.Msg.GetTableConstraints()
		return nil
	})

	errgrp.Go(func() error {
		initTableStatements, err := getTableInitStatementMap(errctx, connectiondataclient, cmd.Source.ConnectionId, cmd.Destination)
		if err != nil {
			return err
		}
		initTableStatementsMap = initTableStatements
		return nil
	})
	if err := errgrp.Wait(); err != nil {
		return nil, err
	}
	tc := map[string]*dbschemas_utils.TableConstraints{}
	for table, constraints := range tableConstraints {
		fkConstraints := []*dbschemas_utils.ForeignConstraint{}
		for _, fk := range constraints.GetConstraints() {
			var foreignKey *dbschemas_utils.ForeignKey
			if fk.ForeignKey != nil {
				foreignKey = &dbschemas_utils.ForeignKey{
					Table:  fk.ForeignKey.Table,
					Column: fk.ForeignKey.Column,
				}
			}
			fkConstraints = append(fkConstraints, &dbschemas_utils.ForeignConstraint{
				Column:     fk.Column,
				IsNullable: fk.IsNullable,
				ForeignKey: foreignKey,
			})
		}
		tc[table] = &dbschemas_utils.TableConstraints{
			Constraints: fkConstraints,
		}
	}
	return &schemaConfig{
		Schemas:                schemas,
		TableConstraints:       tc,
		TablePrimaryKeys:       tablePrimaryKeys,
		InitTableStatementsMap: initTableStatementsMap,
	}, nil
}

func getDestinationSchemaConfig(
	ctx context.Context,
	connectiondataclient mgmtv1alpha1connect.ConnectionDataServiceClient,
	connection *mgmtv1alpha1.Connection,
	cmd *cmdConfig,
	sc *mgmtv1alpha1.ConnectionSchemaConfig,
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
		fmt.Println(bold.Render("No tables found.")) // nolint
		return nil, nil
	}
	schemas := []string{}
	for t := range tableColMap {
		schemas = append(schemas, t)
	}

	fmt.Println(printlog.Render("Building foreign table constraints...")) // nolint
	tableConstraints, err := getDestinationForeignConstraints(ctx, cmd.Destination.Driver, cmd.Destination.ConnectionUrl, schemas)
	if err != nil {
		return nil, err
	}

	fmt.Println(printlog.Render("Building primary key table constraints...")) // nolint
	tablePrimaryKeys, err := getDestinationPrimaryKeyConstraints(ctx, cmd.Destination.Driver, cmd.Destination.ConnectionUrl, schemas)
	if err != nil {
		return nil, err
	}

	initTableStatementsMap := map[string]string{}
	for t := range tableColMap {
		statements := []string{}
		if cmd.Destination.TruncateBeforeInsert {
			if cmd.Destination.TruncateCascade {
				statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", t))
			} else {
				statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s;", t))
			}
		}
		initTableStatementsMap[t] = strings.Join(statements, "\n")
	}

	return &schemaConfig{
		Schemas:                schemaResp.Msg.GetSchemas(),
		TableConstraints:       tableConstraints,
		TablePrimaryKeys:       tablePrimaryKeys,
		InitTableStatementsMap: initTableStatementsMap,
	}, nil
}

func getDestinationForeignConstraints(ctx context.Context, connectionDriver DriverType, connectionUrl string, schemas []string) (map[string]*dbschemas_utils.TableConstraints, error) {
	var constraints dbschemas_utils.TableDependency
	switch connectionDriver {
	case postgresDriver:
		pgquerier := pg_queries.New()
		pool, err := pgxpool.New(ctx, connectionUrl)
		if err != nil {
			return nil, err
		}
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := dbschemas_postgres.GetAllPostgresFkConstraints(pgquerier, cctx, pool, schemas)
		if err != nil {
			return nil, err
		}
		constraints = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
	case mysqlDriver:
		mysqlquerier := mysql_queries.New()
		conn, err := sql.Open(string(connectionDriver), connectionUrl)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println(fmt.Errorf("failed to close mysql connection: %w", err).Error()) // nolint
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := dbschemas_mysql.GetAllMysqlFkConstraints(mysqlquerier, cctx, conn, schemas)
		if err != nil {
			return nil, err
		}
		constraints = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
	default:
		return nil, errors.New("unsupported fk connection")
	}

	return constraints, nil
}

func getDestinationPrimaryKeyConstraints(ctx context.Context, connectionDriver DriverType, connectionUrl string, schemas []string) (map[string]*mgmtv1alpha1.PrimaryConstraint, error) {
	var pc map[string][]string
	switch connectionDriver {
	case postgresDriver:
		pgquerier := pg_queries.New()
		pool, err := pgxpool.New(ctx, connectionUrl)
		if err != nil {
			return nil, err
		}
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := dbschemas_postgres.GetAllPostgresPkConstraints(pgquerier, cctx, pool, schemas)
		if err != nil {
			return nil, err
		}
		pc = dbschemas_postgres.GetPostgresTablePrimaryKeys(allConstraints)
	case mysqlDriver:
		mysqlquerier := mysql_queries.New()
		conn, err := sql.Open(string(connectionDriver), connectionUrl)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println(fmt.Errorf("failed to close mysql connection: %w", err).Error()) // nolint
			}
		}()
		cctx, cancel := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
		defer cancel()
		allConstraints, err := dbschemas_mysql.GetAllMysqlPkConstraints(mysqlquerier, cctx, conn, schemas)
		if err != nil {
			return nil, err
		}
		pc = dbschemas_mysql.GetMysqlTablePrimaryKeys(allConstraints)
	default:
		return nil, errors.New("unsupported fk connection")
	}

	tableConstraints := map[string]*mgmtv1alpha1.PrimaryConstraint{}
	for tableName, cols := range pc {
		tableConstraints[tableName] = &mgmtv1alpha1.PrimaryConstraint{
			Columns: cols,
		}
	}
	return tableConstraints, nil
}

func getDependencyConfigs(
	tc map[string]*dbschemas_utils.TableConstraints,
	tables map[string][]string,
) map[string][]*tabledependency.RunConfig {
	tablesSlice := []string{}
	for t := range tables {
		tablesSlice = append(tablesSlice, t)
	}
	dependencyConfigs := tabledependency.GetRunConfigs(tc, tablesSlice)
	dependencyMap := map[string][]*tabledependency.RunConfig{}
	for _, cfg := range dependencyConfigs {
		_, ok := dependencyMap[cfg.Table]
		if ok {
			dependencyMap[cfg.Table] = append(dependencyMap[cfg.Table], cfg)
		} else {
			dependencyMap[cfg.Table] = []*tabledependency.RunConfig{cfg}
		}
	}
	return dependencyMap
}

func getConfigCount(groupedConfigs [][]*benthosConfigResponse) int {
	count := 0
	for _, group := range groupedConfigs {
		for _, config := range group {
			if config != nil {
				count++
			}
		}
	}
	return count
}

func computeMaxPgBatchCount(numCols int) int {
	if numCols < 1 {
		return maxPgParamLimit
	}
	return clampInt(maxPgParamLimit/numCols, 1, maxPgParamLimit) // automatically rounds down
}

func clampInt(input, low, high int) int {
	if input < low {
		return low
	}
	if input > high {
		return high
	}
	return input
}

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx := range cols {
		pieces[idx] = fmt.Sprintf("this.%s", cols[idx])
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}

func max(a, b int) int {
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
	if connection.ConnectionConfig.GetMysqlConfig() != nil {
		return mysqlConnection, nil
	}
	if connection.ConnectionConfig.GetPgConfig() != nil {
		return postgresConnection, nil
	}
	return "", errors.New("unsupported connection type")
}

func buildPostgresUpdateQuery(table string, columns, primaryKeys []string) string {
	values := make([]string, len(columns))
	var where string
	paramCount := 1
	for i, col := range columns {
		values[i] = fmt.Sprintf("%s = $%d", col, paramCount)
		paramCount++
	}
	if len(primaryKeys) > 0 {
		clauses := []string{}
		for _, col := range primaryKeys {
			clauses = append(clauses, fmt.Sprintf("%s = $%d", col, paramCount))
			paramCount++
		}
		where = fmt.Sprintf("WHERE %s", strings.Join(clauses, " AND "))
	}
	return fmt.Sprintf("UPDATE %s SET %s %s;", table, strings.Join(values, ", "), where)
}

func buildPostgresInsertQuery(table string, columns []string) string {
	values := make([]string, len(columns))
	paramCount := 1
	for i := range columns {
		values[i] = fmt.Sprintf("$%d", paramCount)
		paramCount++
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", table, strings.Join(columns, ", "), strings.Join(values, ", "))
}

func buildMysqlInsertQuery(table string, columns []string) string {
	values := make([]string, len(columns))
	for i := range columns {
		values[i] = "?"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", table, strings.Join(columns, ", "), strings.Join(values, ", "))
}

func buildMysqlUpdateQuery(table string, columns, primaryKeys []string) string {
	values := make([]string, len(columns))
	var where string
	for i, col := range columns {
		values[i] = fmt.Sprintf("%s = ?", col)
	}
	if len(primaryKeys) > 0 {
		clauses := []string{}
		for _, col := range primaryKeys {
			clauses = append(clauses, fmt.Sprintf("%s = ?", col))
		}
		where = fmt.Sprintf("WHERE %s", strings.Join(clauses, " AND "))
	}
	return fmt.Sprintf("UPDATE %s SET %s %s;", table, strings.Join(values, ", "), where)
}
