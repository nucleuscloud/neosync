package sync_cmd

import (
	"context"
	"encoding/json"
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
	_ "github.com/nucleuscloud/neosync/cli/internal/benthos/processors"
	auth_interceptor "github.com/nucleuscloud/neosync/cli/internal/connect/interceptors/auth"
	"github.com/nucleuscloud/neosync/cli/internal/serverconfig"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/benthosdev/benthos/v4/public/service"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "One off sync job to local resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiKey, err := cmd.Flags().GetString("api-key")
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return sync(cmd.Context(), &apiKey)
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

	// stream, err := connectionclient.GetConnectionDataStream(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionDataStreamRequest{
	// 	SourceConnectionId: "3b4db2af-ef33-4e26-b0b9-f6df7518e78b",
	// 	Schema:             "public",
	// 	Table:              "regions",
	// }))
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// fmt.Println("get data stream")

	// for {
	// 	response := stream.Receive()
	// 	fmt.Println("stream")
	// 	if response {
	// 		// for _, v := range stream.Msg().Row {
	// 		// 	fmt.Println(string(v))
	// 		// }
	// 		jsonF, _ := json.MarshalIndent(stream.Msg().Row, "", " ")
	// 		fmt.Printf("\n\n  %s \n\n", string(jsonF))
	// 		// fmt.Println(string(stream.Msg().Data))

	// 	} else {
	// 		return nil
	// 	}

	// }

	sourceConnectionId := "3b4db2af-ef33-4e26-b0b9-f6df7518e78b"
	schemaResp, err := connectionclient.GetConnectionSchema(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionSchemaRequest{
		Id: sourceConnectionId,
	}))
	if err != nil {
		return err
	}

	tables := getSchemaTables(schemaResp.Msg.GetSchemas())
	jsonF, _ := json.MarshalIndent(tables, "", " ")
	fmt.Printf("\n\n  %s \n\n", string(jsonF))

	tablesTest := []*SqlTable{{
		Schema: "public",
		Table:  "regions",
		Columns: []string{"region_id",
			"region_name"},
	}}

	for _, table := range tablesTest {

		benthosConfig := generateBenthosConfig(table.Schema, table.Table, sourceConnectionId, "http://localhost:8080/mgmt.v1alpha1.ConnectionService/GetConnectionDataStream", table.Columns, apiKey)

		configbits, err := yaml.Marshal(benthosConfig.Config)
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
	}

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

func generateBenthosConfig(schema, table, sourceConnectionId, sourceDataStreamUrl string, columns []string, authToken *string) *benthosConfigResponse {
	payload := fmt.Sprintf(`{"source_connection_id": %q, "schema": %q, "table": %q}`, sourceConnectionId, table, schema)
	tableName := fmt.Sprintf("%s.%s", schema, table)

	contentType := "application/connect+json"
	acceptHeader := "application/grpc-web+json"
	bc := &neosync_benthos.BenthosConfig{
		StreamConfig: neosync_benthos.StreamConfig{
			Input: &neosync_benthos.InputConfig{
				Inputs: neosync_benthos.Inputs{
					HttpClient: &neosync_benthos.HttpClient{
						Url:  sourceDataStreamUrl,
						Verb: "POST",
						Headers: &neosync_benthos.Headers{
							Authorization: authToken,
							ContentType:   &contentType,
							Accept:        &acceptHeader,
						},
						Payload: &payload,
						Timeout: "5s",
						Stream: &neosync_benthos.Stream{
							Enabled: true,
							Codec:   "all-bytes",
						},
					},
				},
			},
			Pipeline: &neosync_benthos.PipelineConfig{
				Threads: -1,
				Processors: []neosync_benthos.ProcessorConfig{
					{
						DataStream: &neosync_benthos.DataStream{},
					},
				},
			},
			Output: &neosync_benthos.OutputConfig{
				Outputs: neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver:      "postgres",
						Dsn:         "postgresql://postgres:foofar@localhost:5434/nucleus?sslmode=disable",
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
		DependsOn: []string{},
	}
}
