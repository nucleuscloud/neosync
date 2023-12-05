package sync_cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	"github.com/nucleuscloud/neosync/cli/internal/auth"
	neosync_benthos "github.com/nucleuscloud/neosync/cli/internal/benthos"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	_ "github.com/nucleuscloud/neosync/cli/internal/benthos/inputs"

	"github.com/benthosdev/benthos/v4/public/service"
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
			return sync(cmd.Context(), apiKey)
		},
	}

	return cmd
}

func sync(ctx context.Context, apiKey *string) error {
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

	sourceConnectionId := "3b4db2af-ef33-4e26-b0b9-f6df7518e78b"
	driver := "postgres"

	schemaResp, err := connectionclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		Id: sourceConnectionId,
	}))
	if err != nil {
		return err
	}

	tables := getSchemaTables(schemaResp.Msg.GetSchemas())
	schemaMap := map[string]string{}
	for _, t := range tables {
		schemaMap[t.Schema] = t.Schema
	}

	schemas := make([]string, 0, len(schemaMap))

	for schema := range schemaMap {
		schemas = append(schemas, schema)
	}
	jsonF, _ := json.MarshalIndent(tables, "", " ")
	fmt.Printf("\n\n  %s \n\n", string(jsonF))

	fkConnectionResp, err := connectionclient.GetConnectionForeignConstraints(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionForeignConstraintsRequest{ConnectionId: sourceConnectionId}))
	if err != nil {
		return err
	}
	tableConstraints := fkConnectionResp.Msg.GetTableConstraints()

	configs := []*benthosConfigResponse{}
	for _, table := range tables {
		name := fmt.Sprintf("%s.%s", table.Schema, table.Table)
		dependsOn := tableConstraints[name].GetTables()

		for _, n := range dependsOn {
			if name == n {
				return errors.New("circluar dependency")
			}
		}

		benthosConfig := generateBenthosConfig(table.Schema, table.Table, sourceConnectionId, serverconfig.GetApiBaseUrl(), driver, table.Columns, dependsOn, apiKey)
		configs = append(configs, benthosConfig)
	}

	sortedConfigs := sortConfigs(configs)
	for _, cfg := range sortedConfigs {
		fmt.Println(cfg.Name)
		x, _ := json.MarshalIndent(cfg.DependsOn, "", " ")
		fmt.Printf("%s \n", string(x))
		fmt.Println("--------------")
	}

	// order by dependency size
	// create list of completed configs
	// keep looping until completed configs list is of the table size

	numConfigs := len(sortedConfigs)
	completedConfigs := map[string]struct{}{}
	index := 0
	for len(completedConfigs) != numConfigs {
		fmt.Println("--------")
		jsonF, _ := json.MarshalIndent(completedConfigs, "", " ")
		fmt.Printf("\n  %s \n", string(jsonF))
		fmt.Println(index)

		cfg := sortedConfigs[index]
		fmt.Println(cfg.Name)
		x, _ := json.MarshalIndent(cfg.DependsOn, "", " ")
		fmt.Printf("\n  %s \n", string(x))
		_, completed := completedConfigs[cfg.Name]
		if completed || !isConfigReady(cfg, completedConfigs) {
			fmt.Println(fmt.Sprintf("completed: %t", completed))
			fmt.Println("either completed or not ready")
			index = (index + 1) % numConfigs
			continue
		}

		configbits, err := yaml.Marshal(cfg.Config)
		if err != nil {
			// logger.Error("unable to marshal benthos config", "err", err)
			// settable.SetError(fmt.Errorf("unable to marshal benthos config: %w", err))
			return err
		}
		// fmt.Printf("\n\n  %s \n\n", string(configbits))
		fmt.Println(string(configbits))

		var benthosStream *service.Stream
		go func() {
			for {
				select {
				// case <-time.After(1 * time.Second):
				// 	activity.RecordHeartbeat(ctx)
				case <-ctx.Done():
					if benthosStream != nil {
						// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
						// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
						err := benthosStream.Stop(ctx)
						if err != nil {
							// logger.Error(err.Error())
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
		completedConfigs[cfg.Name] = struct{}{}
		index = (index + 1) % numConfigs
	}

	return nil

}

func isConfigReady(config *benthosConfigResponse, completed map[string]struct{}) bool {
	if config == nil {
		return false
	}

	if len(config.DependsOn) == 0 {
		return true
	}
	for _, dep := range config.DependsOn {
		if _, ok := completed[dep]; !ok {
			return false
		}
	}
	return true
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

func sortConfigs(configs []*benthosConfigResponse) []*benthosConfigResponse {
	sort.SliceStable(configs, func(i, j int) bool {
		return len(configs[i].DependsOn) < len(configs[j].DependsOn)
	})
	return configs
}

const (
	maxPgParamLimit = 65535
)

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
	schema, table, sourceConnectionId, sourceDataStreamUrl, driver string,
	columns, dependsOn []string,
	authToken *string,
) *benthosConfigResponse {
	tableName := fmt.Sprintf("%s.%s", schema, table)

	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					Neosync: &neosync_benthos.Neosync{
						ApiKey:       authToken,
						ApiUrl:       sourceDataStreamUrl,
						ConnectionId: sourceConnectionId,
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
						Dsn:    "postgresql://postgres:foofar@localhost:5434/nucleus?sslmode=disable",
						// InitStatement: initStmt,
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
