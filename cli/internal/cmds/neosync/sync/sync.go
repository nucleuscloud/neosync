package sync_cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	neosync_benthos "github.com/nucleuscloud/neosync/cli/internal/benthos"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
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
)

const (
	maxPgParamLimit = 65535
)

type cmdConfig struct {
	ConnectionId string            `yaml:"connection-id"`
	Destination  destinationConfig `yaml:"destination"`
}

type destinationConfig struct {
	ConnectionUrl string `yaml:"connection-url"`
	Driver        string `yaml:"driver"`
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

			config := &cmdConfig{}
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
				config.ConnectionId = connectionId
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
			if driver != "" {
				config.Destination.Driver = driver
			}

			if config.ConnectionId == "" {
				return fmt.Errorf("must provide connection-id")
			}
			if config.Destination.Driver == "" {
				return fmt.Errorf("must provide destination-driver")
			}
			if config.Destination.ConnectionUrl == "" {
				return fmt.Errorf("must provide destination-connection-url")
			}

			accountId, err := cmd.Flags().GetString("account-id")
			if err != nil {
				return err
			}

			return sync(cmd.Context(), apiKey, &accountId, config)
		},
	}

	cmd.Flags().String("connection-id", "", "Connection id for sync source")
	cmd.Flags().String("destination-connection-url", "", "Connection url for sync output")
	cmd.Flags().String("destination-driver", "", "Connection driver for sync output")
	cmd.Flags().String("account-id", "", "Account source connection is in. Defaults to account id in cli context")
	cmd.Flags().String("config", "", `Location of config file`)
	return cmd
}

func sync(
	ctx context.Context,
	apiKey, accountIdFlag *string,
	cmd *cmdConfig,
) error {
	isAuthEnabled, err := auth.IsAuthEnabled(ctx)
	if err != nil {
		return err
	}

	var token *string
	if isAuthEnabled {
		if apiKey != nil && *apiKey != "" {
			token = apiKey
		} else {
			accessToken, err := userconfig.GetAccessToken()
			if err != nil {
				return err
			}
			token = &accessToken
		}
	}

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

	connectionclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		serverconfig.GetApiBaseUrl(),
		connect.WithInterceptors(
			auth_interceptor.NewInterceptor(isAuthEnabled, auth.AuthHeader, auth.GetAuthHeaderTokenFn(apiKey)),
		),
	)

	connection, err := connectionclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: cmd.ConnectionId,
	}))
	if err != nil {
		return err
	}

	if connection.Msg.Connection.AccountId != *accountId {
		return errors.New(fmt.Sprintf("Connection not found. AccountId: %s", *accountId)) // nolint
	}

	if cmd.Destination.Driver != "mysql" && cmd.Destination.Driver != "postgres" {
		return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
	}

	err = areSourceAndDestCompatible(connection.Msg.Connection, cmd.Destination.Driver)
	if err != nil {
		return err
	}

	fmt.Println("retrieving connection schema...") // nolint
	schemaResp, err := connectionclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		Id: cmd.ConnectionId,
	}))
	if err != nil {
		return err
	}

	tables := getSchemaTables(schemaResp.Msg.GetSchemas())
	schemaMap := map[string]string{}
	for _, t := range tables {
		schemaMap[t.Schema] = t.Schema
	}

	fmt.Println("building foreign table constraints...") // nolint
	fkConnectionResp, err := connectionclient.GetConnectionForeignConstraints(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: cmd.ConnectionId}))
	if err != nil {
		return err
	}
	tableConstraints := fkConnectionResp.Msg.GetTableConstraints()

	fmt.Println("generating configs...") // nolint
	configs := []*benthosConfigResponse{}
	for _, table := range tables {
		name := fmt.Sprintf("%s.%s", table.Schema, table.Table)
		dependsOn := tableConstraints[name].GetTables()

		for _, n := range dependsOn {
			if name == n {
				return fmt.Errorf("circular dependency detected. exiting...")
			}
		}

		benthosConfig := generateBenthosConfig(cmd, table.Schema, table.Table, serverconfig.GetApiBaseUrl(), table.Columns, dependsOn, token)
		configs = append(configs, benthosConfig)
	}

	groupedConfigs := groupConfigsByDependency(configs)
	for _, group := range groupedConfigs {
		errgrp, errctx := errgroup.WithContext(ctx)
		for _, cfg := range group {
			cfg := cfg
			errgrp.Go(func() error {
				err := syncData(errctx, cfg)
				if err != nil {
					return err
				}
				return nil
			})
		}

		if err := errgrp.Wait(); err != nil {
			return err
		}

	}

	fmt.Println("data sync complete") // nolint

	return nil
}

func areSourceAndDestCompatible(connection *mgmtv1alpha1.Connection, destinationDriver string) error {
	switch connection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		if destinationDriver != "postgres" {
			return fmt.Errorf("Connection and destination types are incompatible [postgres, %s]", destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		if destinationDriver != "mysql" {
			return fmt.Errorf("Connection and destination types are incompatible [mysql, %s]", destinationDriver)
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		return errors.New("AWS S3 is not a supported source.")
	default:
		return errors.New("unsupported destination driver. only postgres and mysql are currently supported")
	}
	return nil
}

func syncData(ctx context.Context, cfg *benthosConfigResponse) error {
	fmt.Printf("syncing data for %s \n", cfg.Name) // nolint
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
	schema, table, apiUrl string,
	columns, dependsOn []string,
	authToken *string,
) *benthosConfigResponse {
	tableName := fmt.Sprintf("%s.%s", schema, table)

	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					NeosyncConnectionData: &neosync_benthos.NeosyncConnectionData{
						ApiKey:       authToken,
						ApiUrl:       apiUrl,
						ConnectionId: cmd.ConnectionId,
						Schema:       schema,
						Table:        table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver: cmd.Destination.Driver,
						Dsn:    cmd.Destination.ConnectionUrl,
						// InitStatement: initStmt, TODO
						Table:       tableName,
						Columns:     columns,
						ArgsMapping: buildPlainInsertArgs(columns),
						ConnMaxIdle: 2,
						ConnMaxOpen: 2,
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

func computeMaxPgBatchCount(numCols int) int {
	if numCols < 1 {
		return maxPgParamLimit
	}
	return clampInt(maxPgParamLimit/numCols, 1, maxPgParamLimit) // automatically rounds down
}

// clamps the input between low, high
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
