package datasync_activities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/spf13/viper"

	mysql_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/mysql"
	pg_queries "github.com/nucleuscloud/neosync/backend/gen/go/db/dbschemas/postgresql"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	dbschemas_utils "github.com/nucleuscloud/neosync/backend/pkg/dbschemas"
	"github.com/nucleuscloud/neosync/backend/pkg/sqlconnect"
	tabledependency "github.com/nucleuscloud/neosync/backend/pkg/table-dependency"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
	http_client "github.com/nucleuscloud/neosync/worker/internal/http/client"
	"github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/activities/shared"
)

const nullString = "null"

type GenerateBenthosConfigsRequest struct {
	JobId      string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*BenthosConfigResponse
}

type BenthosConfigResponse struct {
	Name        string
	DependsOn   []*tabledependency.DependsOn
	Config      *neosync_benthos.BenthosConfig
	TableSchema string
	TableName   string
	Columns     []string

	BenthosDsns []*shared.BenthosDsn

	primaryKeys    []string
	excludeColumns []string
	updateConfig   *tabledependency.RunConfig
}

func getNeosyncHttpClient(apiKey string) *http.Client {
	if apiKey != "" {
		return http_client.NewWithHeaders(
			map[string]string{"Authorization": fmt.Sprintf("Bearer %s", apiKey)},
		)
	}
	return http.DefaultClient
}

func getNeosyncUrl() string {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		return "http://localhost:8080"
	}
	return neosyncUrl
}

type Activities struct{}

type RetrieveActivityOptionsRequest struct {
	JobId string
}
type RetrieveActivityOptionsResponse struct {
	SyncActivityOptions *workflow.ActivityOptions
}

func (a *Activities) RetrieveActivityOptions(
	ctx context.Context,
	req *RetrieveActivityOptionsRequest,
	wfmetadata *shared.WorkflowMetadata,
) (*RetrieveActivityOptionsResponse, error) {
	logger := activity.GetLogger(ctx)
	_ = logger

	neosyncUrl := getNeosyncUrl()
	httpClient := getNeosyncHttpClient(viper.GetString("NEOSYNC_API_KEY"))

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)

	jobResp, err := jobclient.GetJob(ctx, connect.NewRequest(&mgmtv1alpha1.GetJobRequest{Id: req.JobId}))
	if err != nil {
		return nil, err
	}
	job := jobResp.Msg.Job
	return &RetrieveActivityOptionsResponse{
		SyncActivityOptions: getSyncActivityOptionsFromJob(job),
	}, nil
}

const (
	defaultStartCloseTimeout = 10 * time.Minute
	defaultMaxAttempts       = 1
)

func getSyncActivityOptionsFromJob(job *mgmtv1alpha1.Job) *workflow.ActivityOptions {
	syncActivityOptions := &workflow.ActivityOptions{}
	if job.SyncOptions != nil {
		if job.SyncOptions.StartToCloseTimeout != nil {
			syncActivityOptions.StartToCloseTimeout = time.Duration(*job.SyncOptions.StartToCloseTimeout)
		}
		if job.SyncOptions.ScheduleToCloseTimeout != nil {
			syncActivityOptions.ScheduleToCloseTimeout = time.Duration(*job.SyncOptions.ScheduleToCloseTimeout)
		}
		if job.SyncOptions.RetryPolicy != nil {
			if job.SyncOptions.RetryPolicy.MaximumAttempts != nil {
				if syncActivityOptions.RetryPolicy == nil {
					syncActivityOptions.RetryPolicy = &temporal.RetryPolicy{}
				}
				syncActivityOptions.RetryPolicy.MaximumAttempts = *job.SyncOptions.RetryPolicy.MaximumAttempts
			}
		}
	} else {
		return &workflow.ActivityOptions{
			StartToCloseTimeout: defaultStartCloseTimeout, // backwards compatible default for pre-existing jobs that do not have sync options defined
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: defaultMaxAttempts, // backwards compatible default for pre-existing jobs that do not have sync options defined
			},
		}
	}
	if syncActivityOptions.StartToCloseTimeout == 0 && syncActivityOptions.ScheduleToCloseTimeout == 0 {
		syncActivityOptions.StartToCloseTimeout = defaultStartCloseTimeout
	}
	if syncActivityOptions.RetryPolicy == nil {
		syncActivityOptions.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: defaultMaxAttempts}
	}
	return syncActivityOptions
}

func (a *Activities) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
	wfmetadata *shared.WorkflowMetadata,
) (*GenerateBenthosConfigsResponse, error) {
	logger := activity.GetLogger(ctx)
	_ = logger
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	neosyncUrl := getNeosyncUrl()
	httpClient := getNeosyncHttpClient(viper.GetString("NEOSYNC_API_KEY"))

	pgpoolmap := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.New()
	mysqlpoolmap := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.New()

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)

	transformerclient := mgmtv1alpha1connect.NewTransformersServiceClient(
		httpClient,
		neosyncUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)
	bbuilder := newBenthosBuilder(
		pgpoolmap,
		pgquerier,
		mysqlpoolmap,
		mysqlquerier,
		jobclient,
		connclient,
		transformerclient,
		&sqlconnect.SqlOpenConnector{},
	)
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slogger = slogger.With(
		"WorkflowID", wfmetadata.WorkflowId,
		"RunID", wfmetadata.RunId,
	)
	return bbuilder.GenerateBenthosConfigs(ctx, req, slogger)
}

type sqlSourceTableOptions struct {
	WhereClause *string
}

func (b *benthosBuilder) buildBenthosSqlSourceConfigResponses(
	ctx context.Context,
	mappings []*TableMapping,
	dsnConnectionId string,
	driver string,
	sourceTableOpts map[string]*sqlSourceTableOptions,
	groupedColumnInfo map[string]map[string]*dbschemas_utils.ColumnInfo,
) ([]*BenthosConfigResponse, error) {
	responses := []*BenthosConfigResponse{}

	for i := range mappings {
		tableMapping := mappings[i]
		cols := buildPlainColumns(tableMapping.Mappings)
		if areAllColsNull(tableMapping.Mappings) {
			// skipping table as no columns are mapped
			continue
		}

		var where string
		tableOpt := sourceTableOpts[neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)]
		if tableOpt != nil && tableOpt.WhereClause != nil {
			where = *tableOpt.WhereClause
		}

		table := neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table)
		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						SqlSelect: &neosync_benthos.SqlSelect{
							Driver: driver,
							Dsn:    "${SOURCE_CONNECTION_DSN}",

							Table:   table,
							Where:   where,
							Columns: cols,
						},
					},
				},
				Pipeline: &neosync_benthos.PipelineConfig{
					Threads:    -1,
					Processors: []neosync_benthos.ProcessorConfig{},
				},
				Output: &neosync_benthos.OutputConfig{
					Outputs: neosync_benthos.Outputs{
						Broker: &neosync_benthos.OutputBrokerConfig{
							Pattern: "fan_out",
							Outputs: []neosync_benthos.Outputs{},
						},
					},
				},
			},
		}

		processorConfigs, err := b.buildProcessorConfigs(ctx, tableMapping.Mappings, groupedColumnInfo[table])
		if err != nil {
			return nil, err
		}

		for _, pc := range processorConfigs {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, *pc)
		}

		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []*tabledependency.DependsOn{},

			BenthosDsns: []*shared.BenthosDsn{{ConnectionId: dsnConnectionId, EnvVarKey: "SOURCE_CONNECTION_DSN"}},

			TableSchema: tableMapping.Schema,
			TableName:   tableMapping.Table,
		})
	}

	return responses, nil
}

func buildBenthosS3Credentials(mgmtCreds *mgmtv1alpha1.AwsS3Credentials) *neosync_benthos.AwsCredentials {
	if mgmtCreds == nil {
		return nil
	}
	creds := &neosync_benthos.AwsCredentials{}
	if mgmtCreds.Profile != nil {
		creds.Profile = *mgmtCreds.Profile
	}
	if mgmtCreds.AccessKeyId != nil {
		creds.Id = *mgmtCreds.AccessKeyId
	}
	if mgmtCreds.SecretAccessKey != nil {
		creds.Secret = *mgmtCreds.SecretAccessKey
	}
	if mgmtCreds.SessionToken != nil {
		creds.Token = *mgmtCreds.SessionToken
	}
	if mgmtCreds.FromEc2Role != nil {
		creds.FromEc2Role = *mgmtCreds.FromEc2Role
	}
	if mgmtCreds.RoleArn != nil {
		creds.Role = *mgmtCreds.RoleArn
	}
	if mgmtCreds.RoleExternalId != nil {
		creds.RoleExternalId = *mgmtCreds.RoleExternalId
	}

	return creds
}

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]*dbschemas_utils.ColumnInfo,
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	for key := range groupedSchemas {
		// For this method, we only care about the schemas+tables that we currently have mappings for
		if _, ok := tableColMappings[key]; !ok {
			delete(groupedSchemas, key)
		}
	}

	if len(tableColMappings) != len(groupedSchemas) {
		return false
	}

	// tests to make sure that every column in the col mappings is present in the db schema
	for table, cols := range tableColMappings {
		schemaCols, ok := groupedSchemas[table]
		if !ok {
			return false
		}
		// job mappings has more columns than the schema
		if len(cols) > len(schemaCols) {
			return false
		}
		for col := range cols {
			if _, ok := schemaCols[col]; !ok {
				return false
			}
		}
	}
	return true
}

func getUniqueColMappingsMap(
	mappings []*mgmtv1alpha1.JobMapping,
) map[string]map[string]struct{} {
	tableColMappings := map[string]map[string]struct{}{}
	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		if _, ok := tableColMappings[key]; ok {
			tableColMappings[key][mapping.Column] = struct{}{}
		} else {
			tableColMappings[key] = map[string]struct{}{
				mapping.Column: {},
			}
		}
	}
	return tableColMappings
}

type RunSqlInitTableStatementsRequest struct {
	JobId      string
	WorkflowId string
}

type RunSqlInitTableStatementsResponse struct {
}

func (a *Activities) RunSqlInitTableStatements(
	ctx context.Context,
	req *RunSqlInitTableStatementsRequest,
) (*RunSqlInitTableStatementsResponse, error) {
	logger := activity.GetLogger(ctx)
	_ = logger
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		neosyncUrl = "http://localhost:8080"
	}

	neosyncApiKey := viper.GetString("NEOSYNC_API_KEY")

	pgpoolmap := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.New()
	mysqlpoolmap := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.New()

	httpClient := http.DefaultClient
	if neosyncApiKey != "" {
		httpClient = http_client.NewWithHeaders(
			map[string]string{"Authorization": fmt.Sprintf("Bearer %s", neosyncApiKey)},
		)
	}

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		httpClient,
		neosyncUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		httpClient,
		neosyncUrl,
	)
	builder := newInitStatementBuilder(
		pgpoolmap,
		pgquerier,
		mysqlpoolmap,
		mysqlquerier,
		jobclient,
		connclient,
		&sqlconnect.SqlOpenConnector{},
	)
	slogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	slogger = slogger.With(
		"WorkflowID", req.WorkflowId,
	)
	return builder.RunSqlInitTableStatements(ctx, req, slogger)
}

func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]*dbschemas_utils.ColumnInfo,
	mappings []*mgmtv1alpha1.JobMapping,
) bool {
	tableColMappings := getUniqueColMappingsMap(mappings)

	if len(tableColMappings) != len(groupedSchemas) {
		return true
	}

	for table, cols := range groupedSchemas {
		mappingCols, ok := tableColMappings[table]
		if !ok {
			return true
		}
		if len(cols) > len(mappingCols) {
			return true
		}
		for col := range cols {
			if _, ok := mappingCols[col]; !ok {
				return true
			}
		}
	}
	return false
}

func areAllColsNull(mappings []*mgmtv1alpha1.JobMapping) bool {
	for _, col := range mappings {
		if col.Transformer.Source != nullString {
			return false
		}
	}
	return true
}

func buildPlainColumns(mappings []*mgmtv1alpha1.JobMapping) []string {
	columns := make([]string, len(mappings))
	for idx := range mappings {
		columns[idx] = mappings[idx].Column
	}
	return columns
}

func splitTableKey(key string) (schema, table string) {
	pieces := strings.Split(key, ".")
	if len(pieces) == 1 {
		return "public", pieces[0]
	}
	return pieces[0], pieces[1]
}

func groupGenerateSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.GenerateSourceSchemaOption,
) map[string]*generateSourceTableOptions {
	groupedMappings := map[string]*generateSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &generateSourceTableOptions{
				Count: int(tableOpt.RowCount), // todo: probably need to update rowcount int64 to int32
			}
		}
	}

	return groupedMappings
}

func groupPostgresSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.PostgresSourceSchemaOption,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sqlSourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}

	return groupedMappings
}

func groupMysqlSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.MysqlSourceSchemaOption,
) map[string]*sqlSourceTableOptions {
	groupedMappings := map[string]*sqlSourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sqlSourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}

	return groupedMappings
}

func groupMappingsByTable(
	mappings []*mgmtv1alpha1.JobMapping,
) []*TableMapping {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}

	for _, mapping := range mappings {
		key := neosync_benthos.BuildBenthosTable(mapping.Schema, mapping.Table)
		groupedMappings[key] = append(groupedMappings[key], mapping)
	}

	output := make([]*TableMapping, 0, len(groupedMappings))
	for key, mappings := range groupedMappings {
		schema, table := splitTableKey(key)
		output = append(output, &TableMapping{
			Schema:   schema,
			Table:    table,
			Mappings: mappings,
		})
	}
	return output
}

type TableMapping struct {
	Schema   string
	Table    string
	Mappings []*mgmtv1alpha1.JobMapping
}

func getUniqueSchemasFromMappings(mappings []*mgmtv1alpha1.JobMapping) []string {
	schemas := map[string]struct{}{}
	for _, mapping := range mappings {
		schemas[mapping.Schema] = struct{}{}
	}

	output := make([]string, 0, len(schemas))

	for schema := range schemas {
		output = append(output, schema)
	}
	return output
}

func getPgDsn(
	config *mgmtv1alpha1.PostgresConnectionConfig,
) (string, error) {
	if config == nil {
		return "", errors.New("must provide non-nil config")
	}
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
		if cfg.Connection == nil {
			return "", errors.New("must provide non-nil connection config")
		}
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

func getMysqlDsn(
	config *mgmtv1alpha1.MysqlConnectionConfig,
) (string, error) {
	if config == nil {
		return "", errors.New("must provide non-nil config")
	}
	switch cfg := config.ConnectionConfig.(type) {
	case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
		if cfg.Connection == nil {
			return "", errors.New("must provide non-nil connection config")
		}
		dburl := fmt.Sprintf(
			"%s:%s@%s(%s:%d)/%s",
			cfg.Connection.User,
			cfg.Connection.Pass,
			cfg.Connection.Protocol,
			cfg.Connection.Host,
			cfg.Connection.Port,
			cfg.Connection.Name,
		)
		return dburl, nil
	case *mgmtv1alpha1.MysqlConnectionConfig_Url:
		return cfg.Url, nil
	default:
		return "", fmt.Errorf("unsupported mysql connection config type")
	}
}

func (b *benthosBuilder) buildProcessorConfigs(ctx context.Context, cols []*mgmtv1alpha1.JobMapping, tableColumnInfo map[string]*dbschemas_utils.ColumnInfo) ([]*neosync_benthos.ProcessorConfig, error) {
	jsCode, err := b.extractJsFunctionsAndOutputs(ctx, cols)
	if err != nil {
		return nil, err
	}

	mutations, err := b.buildMutationConfigs(ctx, cols, tableColumnInfo)
	if err != nil {
		return nil, err
	}

	var processorConfigs []*neosync_benthos.ProcessorConfig
	if mutations != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Mutation: &mutations})
	}
	if jsCode != "" {
		processorConfigs = append(processorConfigs, &neosync_benthos.ProcessorConfig{Javascript: &neosync_benthos.JavascriptConfig{Code: jsCode}})
	}

	return processorConfigs, err
}

func (b *benthosBuilder) extractJsFunctionsAndOutputs(ctx context.Context, cols []*mgmtv1alpha1.JobMapping) (string, error) {
	var benthosOutputs []string
	var jsFunctions []string

	for _, col := range cols {
		if shouldProcessColumn(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				val, err := b.convertUserDefinedFunctionConfig(ctx, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if col.Transformer.Source == "transform_javascript" {
				code := col.Transformer.Config.GetTransformJavascriptConfig().Code
				if code != "" {
					jsFunctions = append(jsFunctions, constructJsFunction(code, col.Column))
					benthosOutputs = append(benthosOutputs, constructBenthosOutput(col.Column))
				}
			}
		}
	}

	if len(jsFunctions) > 0 {
		return constructBenthosJsProcessor(jsFunctions, benthosOutputs), nil
	} else {
		return "", nil
	}
}

func (b *benthosBuilder) buildMutationConfigs(ctx context.Context, cols []*mgmtv1alpha1.JobMapping, tableColumnInfo map[string]*dbschemas_utils.ColumnInfo) (string, error) {
	mutations := []string{}

	for _, col := range cols {
		colInfo := tableColumnInfo[col.Column]
		if shouldProcessColumn(col.Transformer) {
			if _, ok := col.Transformer.Config.Config.(*mgmtv1alpha1.TransformerConfig_UserDefinedTransformerConfig); ok {
				// handle user defined transformer -> get the user defined transformer configs using the id
				val, err := b.convertUserDefinedFunctionConfig(ctx, col.Transformer)
				if err != nil {
					return "", errors.New("unable to look up user defined transformer config by id")
				}
				col.Transformer = val
			}
			if col.Transformer.Source != "transform_javascript" {
				mutation, err := computeMutationFunction(col, colInfo)
				if err != nil {
					return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
				}
				mutations = append(mutations, fmt.Sprintf("root.%s = %s", col.Column, mutation))
			}
		}
	}

	return strings.Join(mutations, "\n"), nil
}

func shouldProcessColumn(t *mgmtv1alpha1.JobMappingTransformer) bool {
	return t != nil &&
		t.Source != "" &&
		t.Source != "passthrough" &&
		t.Source != "generate_default"
}

func constructJsFunction(jsCode, col string) string {
	if jsCode != "" {
		return fmt.Sprintf(`
function fn_%s(value, input){
  %s
};
`, col, jsCode)
	} else {
		return ""
	}
}

func constructBenthosJsProcessor(jsFunctions, benthosOutputs []string) string {
	jsFunctionStrings := strings.Join(jsFunctions, "\n")

	benthosOutputString := strings.Join(benthosOutputs, "\n")

	jsCode := fmt.Sprintf(`
(() => {
%s
const input = benthos.v0_msg_as_structured();
const output = { ...input };
%s
benthos.v0_msg_set_structured(output);
})();`, jsFunctionStrings, benthosOutputString)
	return jsCode
}

func constructBenthosOutput(col string) string {
	return fmt.Sprintf(`output["%[1]s"] = fn_%[1]s(input["%[1]s"], input);`, col)
}

// takes in an user defined config with just an id field and return the right transformer config for that user defined function id
func (b *benthosBuilder) convertUserDefinedFunctionConfig(ctx context.Context, t *mgmtv1alpha1.JobMappingTransformer) (*mgmtv1alpha1.JobMappingTransformer, error) {
	transformer, err := b.transformerclient.GetUserDefinedTransformerById(ctx, connect.NewRequest(&mgmtv1alpha1.GetUserDefinedTransformerByIdRequest{TransformerId: t.Config.GetUserDefinedTransformerConfig().Id}))
	if err != nil {
		return nil, err
	}

	return &mgmtv1alpha1.JobMappingTransformer{
		Source: transformer.Msg.Transformer.Source,
		Config: transformer.Msg.Transformer.Config,
	}, nil
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

/*
function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping, colInfo *dbschemas_utils.ColumnInfo) (string, error) {
	var maxLen int32 = 10000
	if colInfo != nil && colInfo.CharacterMaximumLength != nil {
		maxLen = *colInfo.CharacterMaximumLength
	}

	switch col.Transformer.Source {
	case "generate_categorical":
		categories := col.Transformer.Config.GetGenerateCategoricalConfig().Categories
		return fmt.Sprintf(`generate_categorical(categories: %q)`, categories), nil
	case "generate_email":
		return fmt.Sprintf(`generate_email(max_length:%d)`, maxLen), nil
	case "transform_email":
		pd := col.Transformer.Config.GetTransformEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetTransformEmailConfig().PreserveLength
		excludedDomains := col.Transformer.Config.GetTransformEmailConfig().ExcludedDomains

		sliceBytes, err := json.Marshal(excludedDomains)
		if err != nil {
			return "", err
		}

		excludedDomainstStr := string(sliceBytes)
		return fmt.Sprintf("transform_email(email:this.%s,preserve_domain:%t,preserve_length:%t,excluded_domains:%v,max_length:%d)", col.Column, pd, pl, excludedDomainstStr, maxLen), nil
	case "generate_bool":
		return "generate_bool()", nil
	case "generate_card_number":
		luhn := col.Transformer.Config.GetGenerateCardNumberConfig().ValidLuhn
		return fmt.Sprintf(`generate_card_number(valid_luhn:%t)`, luhn), nil
	case "generate_city":
		return fmt.Sprintf(`generate_city(max_length:%d)`, maxLen), nil
	case "generate_e164_phone_number":
		min := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateE164PhoneNumberConfig().Max
		return fmt.Sprintf(`generate_e164_phone_number(min:%d,max:%d)`, min, max), nil
	case "generate_first_name":
		return fmt.Sprintf(`generate_first_name(max_length:%d)`, maxLen), nil
	case "generate_float64":
		randomSign := col.Transformer.Config.GetGenerateFloat64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateFloat64Config().Min
		max := col.Transformer.Config.GetGenerateFloat64Config().Max
		precision := col.Transformer.Config.GetGenerateFloat64Config().Precision
		return fmt.Sprintf(`generate_float64(randomize_sign:%t, min:%f, max:%f, precision:%d)`, randomSign, min, max, precision), nil
	case "generate_full_address":
		return fmt.Sprintf(`generate_full_address(max_length:%d)`, maxLen), nil
	case "generate_full_name":
		return fmt.Sprintf(`generate_full_name(max_length:%d)`, maxLen), nil
	case "generate_gender":
		ab := col.Transformer.Config.GetGenerateGenderConfig().Abbreviate
		return fmt.Sprintf(`generate_gender(abbreviate:%t,max_length:%d)`, ab, maxLen), nil
	case "generate_int64_phone_number":
		return "generate_int64_phone_number()", nil
	case "generate_int64":
		sign := col.Transformer.Config.GetGenerateInt64Config().RandomizeSign
		min := col.Transformer.Config.GetGenerateInt64Config().Min
		max := col.Transformer.Config.GetGenerateInt64Config().Max
		return fmt.Sprintf(`generate_int64(randomize_sign:%t,min:%d, max:%d)`, sign, min, max), nil
	case "generate_last_name":
		return fmt.Sprintf(`generate_last_name(max_length:%d)`, maxLen), nil
	case "generate_sha256hash":
		return `generate_sha256hash()`, nil
	case "generate_ssn":
		return "generate_ssn()", nil
	case "generate_state":
		return "generate_state()", nil
	case "generate_street_address":
		return fmt.Sprintf(`generate_street_address(max_length:%d)`, maxLen), nil
	case "generate_string_phone_number":
		min := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Min
		max := col.Transformer.Config.GetGenerateStringPhoneNumberConfig().Max
		return fmt.Sprintf("generate_string_phone_number(min:%d,max:%d,max_length:%d)", min, max, maxLen), nil
	case "generate_string":
		min := col.Transformer.Config.GetGenerateStringConfig().Min
		max := col.Transformer.Config.GetGenerateStringConfig().Max
		return fmt.Sprintf(`generate_string(min:%d,max:%d,max_length:%d)`, min, max, maxLen), nil
	case "generate_unixtimestamp":
		return "generate_unixtimestamp()", nil
	case "generate_username":
		return fmt.Sprintf(`generate_username(max_length:%d)`, maxLen), nil
	case "generate_utctimestamp":
		return "generate_utctimestamp()", nil
	case "generate_uuid":
		ih := col.Transformer.Config.GetGenerateUuidConfig().IncludeHyphens
		return fmt.Sprintf("generate_uuid(include_hyphens:%t)", ih), nil
	case "generate_zipcode":
		return "generate_zipcode()", nil
	case "transform_e164_phone_number":
		pl := col.Transformer.Config.GetTransformE164PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_e164_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_first_name":
		pl := col.Transformer.Config.GetTransformFirstNameConfig().PreserveLength
		return fmt.Sprintf("transform_first_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_float64":
		rMin := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformFloat64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_float64(value:this.%s,randomization_range_min:%f,randomization_range_max:%f)`, col.Column, rMin, rMax), nil
	case "transform_full_name":
		pl := col.Transformer.Config.GetTransformFullNameConfig().PreserveLength
		return fmt.Sprintf("transform_full_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_int64_phone_number":
		pl := col.Transformer.Config.GetTransformInt64PhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_int64_phone_number(value:this.%s,preserve_length:%t)", col.Column, pl), nil
	case "transform_int64":
		rMin := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMin
		rMax := col.Transformer.Config.GetTransformInt64Config().RandomizationRangeMax
		return fmt.Sprintf(`transform_int64(value:this.%s,randomization_range_min:%d,randomization_range_max:%d)`, col.Column, rMin, rMax), nil
	case "transform_last_name":
		pl := col.Transformer.Config.GetTransformLastNameConfig().PreserveLength
		return fmt.Sprintf("transform_last_name(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_phone_number":
		pl := col.Transformer.Config.GetTransformPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("transform_phone_number(value:this.%s,preserve_length:%t,max_length:%d)", col.Column, pl, maxLen), nil
	case "transform_string":
		pl := col.Transformer.Config.GetTransformStringConfig().PreserveLength
		return fmt.Sprintf(`transform_string(value:this.%s,preserve_length:%t,max_length:%d)`, col.Column, pl, maxLen), nil
	case "null":
		return "null", nil
	case "generate_default":
		return "default", nil
	case "transform_character_scramble":
		return fmt.Sprintf(`transform_character_scramble(value:this.%s)`, col.Column), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
