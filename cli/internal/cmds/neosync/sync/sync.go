package sync_cmd

import (
	"context"
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

			connectionId, err := cmd.Flags().GetString("connection-id")
			if err != nil {
				return err
			}

			destConnUrl, err := cmd.Flags().GetString("destination-connection-url")
			if err != nil {
				return err
			}

			driver, err := cmd.Flags().GetString("destination-driver")
			if err != nil {
				return err
			}

			return sync(cmd.Context(), apiKey, connectionId, destConnUrl, driver)
		},
	}

	cmd.Flags().String("connection-id", "", "connection id for sync source (required)")
	cmd.Flags().String("destination-connection-url", "", "destination url for sync output (required)")
	cmd.Flags().String("destination-driver", "", "destination driver for sync output (required)")
	cmd.MarkFlagRequired("connection-id")
	cmd.MarkFlagRequired("destination-connection-url")
	cmd.MarkFlagRequired("destination-driver")

	return cmd
}

func sync(
	ctx context.Context,
	apiKey *string,
	connectionId, destConnUrl, driver string,
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

	fmt.Println("retrieving connection schema...") // nolint
	schemaResp, err := connectionclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		Id: connectionId,
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
	fkConnectionResp, err := connectionclient.GetConnectionForeignConstraints(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: connectionId}))
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

		benthosConfig := generateBenthosConfig(table.Schema, table.Table, connectionId, serverconfig.GetApiBaseUrl(), destConnUrl, driver, table.Columns, dependsOn, apiKey)
		configs = append(configs, benthosConfig)
	}

	groupedConfigs := groupConfigsByDependency(configs)

	for _, group := range groupedConfigs {
		errgrp, errctx := errgroup.WithContext(ctx)
		for _, cfg := range group {
			cfg := cfg
			errgrp.Go(func() error {
				fmt.Printf("syncing data for %s \n", cfg.Name) // nolint
				configbits, err := yaml.Marshal(cfg.Config)
				if err != nil {
					return err
				}

				var benthosStream *service.Stream
				go func() {
					for { // nolint
						select {
						// case <-time.After(1 * time.Second):
						// 	activity.RecordHeartbeat(ctx)
						case <-errctx.Done():
							if benthosStream != nil {
								// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
								// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
								err := benthosStream.Stop(errctx)
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

				err = stream.Run(errctx)
				if err != nil {
					return fmt.Errorf("unable to run benthos stream: %w", err)
				}
				benthosStream = nil
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

type benthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

func generateBenthosConfig(
	schema, table, connectionId, apiUrl, destConnUrl, driver string,
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
						ConnectionId: connectionId,
						Schema:       schema,
						Table:        table,
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver: driver,
						Dsn:    destConnUrl,
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
