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
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_mysql "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/mysql"
	dbschemas_postgres "github.com/nucleuscloud/neosync/backend/pkg/dbschemas/postgres"
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
	InitSchema           bool       `yaml:"init-table-schema,omitempty"`
	TruncateBeforeInsert bool       `yaml:"truncate-before-insert,omitempty"`
	TruncateCascade      bool       `yaml:"truncate-cascade,omitempty"`
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

	var tables []*SqlTable
	var tableConstraints map[string]*mgmtv1alpha1.ForeignConstraintTables
	var initTableStatementsMap map[string]string

	switch connectionType {
	case awsS3Connection:
		// TODO handle job id
		schemaResp, err := connectiondataclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: connection.Id,
			SchemaConfig: &mgmtv1alpha1.ConnectionSchemaConfig{
				Config: &mgmtv1alpha1.ConnectionSchemaConfig_AwsS3Config{
					AwsS3Config: &mgmtv1alpha1.AwsS3SchemaConfig{
						Id: &mgmtv1alpha1.AwsS3SchemaConfig_JobRunId{
							JobRunId: *cmd.Source.ConnectionOpts.JobRunId,
						},
					},
				},
			},
		}))
		if err != nil {
			return err
		}

		tables = getSchemaTables(schemaResp.Msg.GetSchemas())
		if len(tables) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		schemaMap := map[string]struct{}{}
		for _, s := range tables {
			schemaMap[s.Schema] = struct{}{}
		}
		schemas := []string{}
		for s := range schemaMap {
			schemas = append(schemas, s)
		}

		fmt.Println(printlog.Render("Building foreign table constraints...")) // nolint
		constraints, err := getDestinationForeignConstraints(ctx, cmd.Destination.Driver, cmd.Destination.ConnectionUrl, schemas)
		if err != nil {
			return err
		}
		tableConstraints = constraints

		initTableStatementsMap = map[string]string{}
		for _, t := range tables {
			statements := []string{}
			if cmd.Destination.TruncateBeforeInsert {
				if cmd.Destination.TruncateCascade {
					statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE;", t.Schema, t.Table))
				} else {
					statements = append(statements, fmt.Sprintf("TRUNCATE TABLE %s.%s;", t.Schema, t.Table))
				}
			}
			initTableStatementsMap[fmt.Sprintf("%s.%s", t.Schema, t.Table)] = strings.Join(statements, "\n")
		}

	case mysqlConnection:
		schemaResp, err := connectiondataclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: connection.Id,
			SchemaConfig: &mgmtv1alpha1.ConnectionSchemaConfig{
				Config: &mgmtv1alpha1.ConnectionSchemaConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlSchemaConfig{},
				},
			},
		}))
		if err != nil {
			return err
		}

		tables = getSchemaTables(schemaResp.Msg.GetSchemas())
		if len(tables) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		fmt.Println(printlog.Render("Building foreign table constraints...")) // nolint
		fkConnectionResp, err := connectiondataclient.GetConnectionForeignConstraints(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tableConstraints = fkConnectionResp.Msg.GetTableConstraints()

		initTableStatementsMap, err = getTableInitStatementMap(ctx, connectiondataclient, cmd.Source.ConnectionId, cmd.Destination)
		if err != nil {
			return err
		}
	case postgresConnection:
		schemaResp, err := connectiondataclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
			ConnectionId: connection.Id,
			SchemaConfig: &mgmtv1alpha1.ConnectionSchemaConfig{
				Config: &mgmtv1alpha1.ConnectionSchemaConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresSchemaConfig{},
				},
			},
		}))
		if err != nil {
			return err
		}

		tables = getSchemaTables(schemaResp.Msg.GetSchemas())
		if len(tables) == 0 {
			fmt.Println(bold.Render("No tables found.")) // nolint
			return nil
		}

		fmt.Println(printlog.Render("Building foreign table constraints...")) // nolint
		fkConnectionResp, err := connectiondataclient.GetConnectionForeignConstraints(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: cmd.Source.ConnectionId}))
		if err != nil {
			return err
		}
		tableConstraints = fkConnectionResp.Msg.GetTableConstraints()

		initTableStatementsMap, err = getTableInitStatementMap(ctx, connectiondataclient, cmd.Source.ConnectionId, cmd.Destination)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("this connection type is not currently supported")
	}

	fmt.Println(printlog.Render("Generating configs... \n")) // nolint
	configs := []*benthosConfigResponse{}
	for _, table := range tables {
		name := fmt.Sprintf("%s.%s", table.Schema, table.Table)
		dependsOn := tableConstraints[name].GetTables()

		for _, n := range dependsOn {
			if name == n {
				return fmt.Errorf("Circular dependency detected. exiting...")
			}
		}
		initStatement := initTableStatementsMap[name]

		benthosConfig := generateBenthosConfig(cmd, connectionType, table.Schema, table.Table, serverconfig.GetApiBaseUrl(), initStatement, table.Columns, dependsOn, token)
		configs = append(configs, benthosConfig)
	}

	groupedConfigs := groupConfigsByDependency(configs)

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
	log.Printf("Done! Completed %d tables.", len(configs))

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

func getSchemaTables(schemas []*mgmtv1alpha1.DatabaseColumn) []*SqlTable {
	tableColMap := map[string][]string{}
	for _, record := range schemas {
		table := fmt.Sprintf("%s.%s", record.Schema, record.Table)
		_, ok := tableColMap[table]
		if ok {
			tableColMap[table] = append(tableColMap[table], record.Column)
		} else {
			tableColMap[table] = []string{record.Column}
		}
	}

	tables := []*SqlTable{}
	for table, cols := range tableColMap {
		slice := strings.Split(table, ".")
		tables = append(tables, &SqlTable{
			Table:   slice[1],
			Schema:  slice[0],
			Columns: cols,
		})
	}
	return tables
}

type benthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

func generateBenthosConfig(
	cmd *cmdConfig,
	connectionType ConnectionType,
	schema, table, apiUrl, initStatement string,
	columns, dependsOn []string,
	authToken *string,
) *benthosConfigResponse {
	tableName := fmt.Sprintf("%s.%s", schema, table)

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
						Schema:         schema,
						Table:          table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver:        string(cmd.Destination.Driver),
						Dsn:           cmd.Destination.ConnectionUrl,
						InitStatement: initStatement,
						Table:         tableName,
						Columns:       columns,
						ArgsMapping:   buildPlainInsertArgs(columns),
						ConnMaxIdle:   2,
						ConnMaxOpen:   2,
						Batching: &neosync_benthos.Batching{
							Period: "1s",
							// max allowed by postgres in a single batch
							Count: computeMaxPgBatchCount(len(columns)),
						},
					},
				},
			},
		},
	}

	return &benthosConfigResponse{
		Name:      tableName,
		Config:    bc,
		DependsOn: dependsOn,
	}
}

func groupConfigsByDependency(configs []*benthosConfigResponse) [][]*benthosConfigResponse {
	configMap := make(map[string]*benthosConfigResponse, len(configs))
	for _, c := range configs {
		configMap[c.Name] = c
	}

	depGraph := make(map[string][]*benthosConfigResponse)
	indegree := make(map[string]int)

	for _, cfg := range configs {
		indegree[cfg.Name] = 0
		depGraph[cfg.Name] = []*benthosConfigResponse{}
	}

	for _, cfg := range configs {
		for _, dep := range cfg.DependsOn {
			depGraph[dep] = append(depGraph[dep], cfg)
			indegree[cfg.Name]++
		}
	}

	var queue []string
	for _, cfg := range configs {
		if indegree[cfg.Name] == 0 {
			queue = append(queue, cfg.Name)
		}
	}

	var groupedConfigs [][]*benthosConfigResponse

	for len(queue) > 0 {
		var group []*benthosConfigResponse
		var nextQueue []string

		for _, cfgName := range queue {
			cfg := configMap[cfgName]
			group = append(group, cfg)

			for _, nextCfg := range depGraph[cfgName] {
				indegree[nextCfg.Name]--
				if indegree[nextCfg.Name] == 0 {
					nextQueue = append(nextQueue, nextCfg.Name)
				}
			}
		}

		groupedConfigs = append(groupedConfigs, group)
		queue = nextQueue
	}

	return groupedConfigs
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

func getDestinationForeignConstraints(ctx context.Context, connectionDriver DriverType, connectionUrl string, schemas []string) (map[string]*mgmtv1alpha1.ForeignConstraintTables, error) {
	var td map[string][]string
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
		td = dbschemas_postgres.GetPostgresTableDependencies(allConstraints)
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
		td = dbschemas_mysql.GetMysqlTableDependencies(allConstraints)
	default:
		return nil, errors.New("unsupported fk connection")
	}

	constraints := map[string]*mgmtv1alpha1.ForeignConstraintTables{}
	for key, tables := range td {
		constraints[key] = &mgmtv1alpha1.ForeignConstraintTables{
			Tables: tables,
		}
	}

	return constraints, nil
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
