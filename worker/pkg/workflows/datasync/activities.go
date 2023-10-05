package datasync

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/benthosdev/benthos/v4/public/service"
	"github.com/jackc/pgx/v5"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/plugins"
	neosync_plugins "github.com/nucleuscloud/neosync/worker/internal/benthos/plugins"
	dbschemas_postgres "github.com/nucleuscloud/neosync/worker/internal/dbschemas/postgres"
)

type GenerateBenthosConfigsRequest struct {
	JobId      string
	BackendUrl string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*benthosConfigResponse
}

type benthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

type Activities struct{}

func (a *Activities) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
) (*GenerateBenthosConfigsResponse, error) {
	job, err := a.getJobById(ctx, req.BackendUrl, req.JobId)
	if err != nil {
		return nil, err
	}
	responses := []*benthosConfigResponse{}

	sourceConnection, err := a.getConnectionById(ctx, req.BackendUrl, job.Source.ConnectionId)
	if err != nil {
		return nil, err
	}

	transformerConfigs, err := a.getTransformerConfigs(ctx, req.BackendUrl, job.AccountId)
	if err != nil {
		return nil, err
	}

	//ED:add more connection types here as we build more out and probably refactor cases into a separate file
	switch connection := sourceConnection.ConnectionConfig.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		dsn, err := getPgDsn(connection.PgConfig)
		if err != nil {
			return nil, err
		}

		groupedMappings := groupMappingsByTable(job.Mappings)
		for key, mappings := range groupedMappings {
			bc := &neosync_benthos.BenthosConfig{
				StreamConfig: neosync_benthos.StreamConfig{
					Input: &neosync_benthos.InputConfig{
						Inputs: neosync_benthos.Inputs{
							SqlSelect: &neosync_benthos.SqlSelect{
								Driver: "postgres",
								Dsn:    dsn,

								Table:   key,
								Columns: buildPlainColumns(mappings),
							},
						},
					},
					Pipeline: &neosync_benthos.PipelineConfig{
						Threads:    -1,
						Processors: []neosync_benthos.ProcessorConfig{},
					},
					Output: &neosync_benthos.OutputConfig{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			}

			mutation, err := buildProcessorMutation(job.Mappings, transformerConfigs)
			if err != nil {
				return nil, err
			}

			/*ED:when the component here is Mutation: benthos throws this error:
			field processor is invalid when the component type is mutation (processor)
			switching it to Bloblang makes it pass */

			if mutation != "" {
				bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, neosync_benthos.ProcessorConfig{
					Bloblang: mutation,
				})
			}

			responses = append(responses, &benthosConfigResponse{
				Name:   key, // todo: may need to expand on this
				Config: bc,
			})
		}

	default:
		return nil, fmt.Errorf("unsupported source connection")
	}

	for _, destination := range job.Destinations {
		destinationConnection, err := a.getConnectionById(ctx, req.BackendUrl, destination.ConnectionId)
		if err != nil {
			return nil, err
		}
		for _, resp := range responses {
			//ED:add more destination cases here and evntually probably refactor the case statements out into a separate file
			switch connection := destinationConnection.ConnectionConfig.Config.(type) {
			case *mgmtv1alpha1.ConnectionConfig_PgConfig:
				dsn, err := getPgDsn(connection.PgConfig)
				if err != nil {
					return nil, err
				}
				// todo: make this more efficient to reduce amount of times we have to connect to the source database
				initStmt, err := a.getInitStatementFromPostgres(ctx, resp.Config.Input.SqlSelect.Dsn, resp.Config.Input.SqlSelect.Table)
				if err != nil {
					return nil, err
				}
				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					SqlInsert: &neosync_benthos.SqlInsert{
						Driver: "postgres",
						Dsn:    dsn,

						Table:         resp.Config.Input.SqlSelect.Table,
						Columns:       resp.Config.Input.SqlSelect.Columns,
						ArgsMapping:   buildPlainInsertArgs(resp.Config.Input.SqlSelect.Columns),
						InitStatement: initStmt,
					},
				})
			case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
				s3pathpieces := []string{}
				if connection.AwsS3Config.PathPrefix != nil && *connection.AwsS3Config.PathPrefix != "" {
					s3pathpieces = append(s3pathpieces, strings.Trim(*connection.AwsS3Config.PathPrefix, "/"))
				}
				s3pathpieces = append(
					s3pathpieces,
					"workflows",
					req.WorkflowId,
					"activities",
					resp.Name, // may need to do more here
					"data",
					`${!count("files")}.json.gz}`,
				)

				resp.Config.Output.Broker.Outputs = append(resp.Config.Output.Broker.Outputs, neosync_benthos.Outputs{
					AwsS3: &neosync_benthos.AwsS3Insert{
						Bucket:      connection.AwsS3Config.BucketArn,
						MaxInFlight: 64,
						Path:        fmt.Sprintf("/%s", strings.Join(s3pathpieces, "/")),
						Batching: &neosync_benthos.Batching{
							Count:  100,
							Period: "5s",
							Processors: []*neosync_benthos.BatchProcessor{
								{Archive: &neosync_benthos.ArchiveProcessor{Format: "json_array"}},
								{Compress: &neosync_benthos.CompressProcessor{Algorithm: "gzip"}},
							},
						},
					},
				})
				// todo: configure provided aws creds
			default:
				return nil, fmt.Errorf("unsupported destination connection config")
			}
		}
	}

	return &GenerateBenthosConfigsResponse{
		BenthosConfigs: responses,
	}, nil
}

func (a *Activities) getInitStatementFromPostgres(
	ctx context.Context,
	dsn string,
	table string,
) (string, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return "", err
	}
	defer conn.Close(ctx)

	return dbschemas_postgres.GetTableCreateStatement(ctx, conn, &dbschemas_postgres.GetTableCreateStatementRequest{
		Table: table,
	})
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := []string{}

	for _, col := range mappings {
		if !col.Exclude {
			columns = append(columns, col.Column)
		}
	}

	return columns
}

type SyncRequest struct {
	BenthosConfig string
}
type SyncResponse struct{}

func (a *Activities) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	streambldr := service.NewStreamBuilder()

	fmt.Println("the config", req.BenthosConfig)

	err := streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}
	stream, err := streambldr.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos stream: %w", err)
	}

	err = stream.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run benthos stream: %w", err)
	}
	return &SyncResponse{}, nil
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) map[string][]*mgmtv1alpha1.JobMapping {
	output := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := buildBenthosTable(mapping.Schema, mapping.Table)
		output[key] = append(output[key], mapping)
	}
	return output
}

func buildBenthosTable(schema, table string) string {
	if schema != "" {
		return fmt.Sprintf("%s.%s", schema, table)
	}
	return table
}

func (a *Activities) getJobById(ctx context.Context, backendurl, jobId string) (*mgmtv1alpha1.Job, error) {
	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		backendurl,
	)

	getjobResp, err := jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{
		Id: jobId,
	}))
	if err != nil {
		return nil, err
	}

	return getjobResp.Msg.Job, nil
}

func (a *Activities) getConnectionById(ctx context.Context, backendurl, connectionId string) (*mgmtv1alpha1.Connection, error) {
	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		backendurl,
	)

	getConnResp, err := connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}

func getPgDsn(
	config *mgmtv1alpha1.PostgresConnectionConfig,
) (string, error) {
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		dburl := fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s",
			cfg.Connection.User,
			cfg.Connection.Pass,
			cfg.Connection.Host,
			cfg.Connection.Port,
			cfg.Connection.Name,
		)
		if cfg.Connection.SslMode != nil && *cfg.Connection.SslMode != "" {
			dburl = fmt.Sprintf("%s?sslmode=%s", dburl, *cfg.Connection.SslMode)
		}
		return dburl, nil
	case *mgmtv1alpha1.PostgresConnectionConfig_Url:
		return cfg.Url, nil
	default:
		return "", fmt.Errorf("unsupported postgres connection config type")
	}
}

func (a *Activities) getTransformerConfigs(ctx context.Context, backendurl, accountId string) ([]*mgmtv1alpha1.Transformer, error) {
	transformerClient := mgmtv1alpha1connect.NewTransformersServiceClient(
		http.DefaultClient,
		backendurl,
	)

	getTransformerResp, err := transformerClient.GetTransformers(ctx, connect.NewRequest(&mgmtv1alpha1.GetTransformersRequest{
		AccountId: accountId,
	}))
	if err != nil {
		return nil, err
	}

	return getTransformerResp.Msg.Transformers, nil
}

func buildProcessorMutation(cols []*mgmtv1alpha1.JobMapping, transformerConfigs []*mgmtv1alpha1.Transformer) (string, error) {

	transformerConfigMap := make(map[string]*mgmtv1alpha1.Transformer)

	for _, val := range transformerConfigs {
		transformerConfigMap[val.Name] = val
	}

	//ED: have to register the custom method, should build this list based on the selected transformers
	//so we're not calling and registering transformers that aren't being used
	neosync_plugins.Emailtransformer()

	mutations := []string{}

	for _, col := range cols {
		//ED: checks that the user-selected transformer is defined and in the transfomer map from the db
		if value, ok := transformerConfigMap[col.Transformer]; ok {
			mutation, err := computeMutationFunction(col, value)
			if err != nil {
				return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
			}

			mutations = append(mutations, mutation)

		} else {
			return "", fmt.Errorf("unable to recognize transformer")
		}
	}

	return strings.Join(mutations, "\n"), nil
}

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, transformer *mgmtv1alpha1.Transformer) (string, error) {

	switch transformer.Name {
	case "email":
		option := transformer.Config.GetEmailConfig()
		return fmt.Sprintf("root.%s = this.email.emailtransformer(%t, %t)", col.Column, option.PreserveLength, option.PreserveDomain), nil
	case "passthrough":
		return fmt.Sprintf("root.%s = %s", col.Column, col.Column), nil
	default:
		return "", fmt.Errorf("unable to recognize transformer")
	}
}

func buildPlainInsertArgs(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	pieces := make([]string, len(cols))
	for idx, col := range cols {
		pieces[idx] = fmt.Sprintf("this.%s", col)
	}
	return fmt.Sprintf("root = [%s]", strings.Join(pieces, ", "))
}
