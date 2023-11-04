package datasync_activities

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"

	_ "github.com/benthosdev/benthos/v4/public/components/aws"
	_ "github.com/benthosdev/benthos/v4/public/components/io"
	_ "github.com/benthosdev/benthos/v4/public/components/pure"
	_ "github.com/benthosdev/benthos/v4/public/components/pure/extended"
	_ "github.com/benthosdev/benthos/v4/public/components/sql"
	"github.com/benthosdev/benthos/v4/public/service"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	mysql_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/mysql"
	pg_queries "github.com/nucleuscloud/neosync/worker/gen/go/db/postgresql"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"
	_ "github.com/nucleuscloud/neosync/worker/internal/benthos/transformers"
)

const nullString = "null"

type GenerateBenthosConfigsRequest struct {
	JobId      string
	BackendUrl string
	WorkflowId string
}
type GenerateBenthosConfigsResponse struct {
	BenthosConfigs []*BenthosConfigResponse
}

type BenthosConfigResponse struct {
	Name      string
	DependsOn []string
	Config    *neosync_benthos.BenthosConfig
}

type Activities struct{}

func (a *Activities) GenerateBenthosConfigs(
	ctx context.Context,
	req *GenerateBenthosConfigsRequest,
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

	pgpoolmap := map[string]pg_queries.DBTX{}
	pgquerier := pg_queries.New()
	mysqlpoolmap := map[string]mysql_queries.DBTX{}
	mysqlquerier := mysql_queries.New()

	jobclient := mgmtv1alpha1connect.NewJobServiceClient(
		http.DefaultClient,
		req.BackendUrl,
	)

	connclient := mgmtv1alpha1connect.NewConnectionServiceClient(
		http.DefaultClient,
		req.BackendUrl,
	)
	bbuilder := newBenthosBuilder(
		pgpoolmap,
		pgquerier,
		mysqlpoolmap,
		mysqlquerier,
		jobclient,
		connclient,
	)
	return bbuilder.GenerateBenthosConfigs(ctx, req, logger)
}

type sourceTableOptions struct {
	WhereClause *string
}

func buildBenthosSourceConfigReponses(
	mappings []*TableMapping,
	dsn string,
	driver string,
	sourceTableOpts map[string]*sourceTableOptions,
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

		bc := &neosync_benthos.BenthosConfig{
			StreamConfig: neosync_benthos.StreamConfig{
				Input: &neosync_benthos.InputConfig{
					Inputs: neosync_benthos.Inputs{
						SqlSelect: &neosync_benthos.SqlSelect{
							Driver: driver,
							Dsn:    dsn,

							Table:   neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table),
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
		mutation, err := buildProcessorMutation(tableMapping.Mappings)
		if err != nil {
			return nil, err
		}
		if mutation != "" {
			bc.StreamConfig.Pipeline.Processors = append(bc.StreamConfig.Pipeline.Processors, neosync_benthos.ProcessorConfig{
				Mutation: mutation,
			})
		}
		responses = append(responses, &BenthosConfigResponse{
			Name:      neosync_benthos.BuildBenthosTable(tableMapping.Schema, tableMapping.Table), // todo: may need to expand on this
			Config:    bc,
			DependsOn: []string{},
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

func areMappingsSubsetOfSchemas(
	groupedSchemas map[string]map[string]struct{},
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

func shouldHaltOnSchemaAddition(
	groupedSchemas map[string]map[string]struct{},
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
		if col.Transformer.Value != nullString {
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

// used to record metadata in activity event history
type SyncMetadata struct {
	Schema string
	Table  string
}
type SyncRequest struct {
	BenthosConfig string
}
type SyncResponse struct{}

func (a *Activities) Sync(ctx context.Context, req *SyncRequest, metadata *SyncMetadata) (*SyncResponse, error) {
	logger := activity.GetLogger(ctx)
	var benthosStream *service.Stream
	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				activity.RecordHeartbeat(ctx)
			case <-ctx.Done():
				if benthosStream != nil {
					// this must be here because stream.Run(ctx) doesn't seem to fully obey a canceled context when
					// a sink is in an error state. We want to explicitly call stop here because the workflow has been canceled.
					err := benthosStream.Stop(ctx)
					if err != nil {
						logger.Error(err.Error())
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
		"metadata", metadata,
		"benthos", "true",
	))

	err := streambldr.SetYAML(req.BenthosConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to convert benthos config to yaml for stream builder: %w", err)
	}

	stream, err := streambldr.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build benthos stream: %w", err)
	}
	benthosStream = stream

	err = stream.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to run benthos stream: %w", err)
	}
	benthosStream = nil
	return &SyncResponse{}, nil
}

func groupPostgresSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.PostgresSourceSchemaOption,
) map[string]*sourceTableOptions {
	groupedMappings := map[string]*sourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sourceTableOptions{
				WhereClause: tableOpt.WhereClause,
			}
		}
	}

	return groupedMappings
}

func groupMysqlSourceOptionsByTable(
	schemaOptions []*mgmtv1alpha1.MysqlSourceSchemaOption,
) map[string]*sourceTableOptions {
	groupedMappings := map[string]*sourceTableOptions{}

	for idx := range schemaOptions {
		schemaOpt := schemaOptions[idx]
		for tidx := range schemaOpt.Tables {
			tableOpt := schemaOpt.Tables[tidx]
			key := neosync_benthos.BuildBenthosTable(schemaOpt.Schema, tableOpt.Table)
			groupedMappings[key] = &sourceTableOptions{
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

func buildProcessorMutation(cols []*mgmtv1alpha1.JobMapping) (string, error) {
	pieces := []string{}

	for _, col := range cols {
		if col.Transformer != nil && col.Transformer.Value != "" && col.Transformer.Value != "passthrough" {
			mutation, err := computeMutationFunction(col)
			if err != nil {
				return "", fmt.Errorf("%s is not a supported transformer: %w", col.Transformer, err)
			}
			pieces = append(pieces, fmt.Sprintf("root.%s = %s", col.Column, mutation))
		}
	}
	return strings.Join(pieces, "\n"), nil
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
method transformers
root.{destination_col} = this.{source_col}.transformermethod(args)

function transformers
root.{destination_col} = transformerfunction(args)
*/

func computeMutationFunction(col *mgmtv1alpha1.JobMapping) (string, error) {

	switch col.Transformer.Value {
	case "latitude":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "longitude":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "date":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "time_string":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "month_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "year_string":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "day_of_week":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "day_of_month":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "century":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "timezone":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "time_period":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "email":
		pd := col.Transformer.Config.GetEmailConfig().PreserveDomain
		pl := col.Transformer.Config.GetEmailConfig().PreserveLength
		return fmt.Sprintf("this.%s.emailtransformer(%t, %t)", col.Transformer.Value, pd, pl), nil
	case "mac_address":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "domain_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "url":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "username":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "ipv4":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "ipv6":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "password":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "jwt":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "word":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "sentence":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "paragraph":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "title_male":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "title_female":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "first_name":
		pl := col.Transformer.Config.GetFirstNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.firstnametransformer(%t)", col.Column, pl), nil
	case "first_name_female":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "last_name":
		pl := col.Transformer.Config.GetLastNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.lastnametransformer(%t)", col.Column, pl), nil
	case "full_name":
		pl := col.Transformer.Config.GetFullNameConfig().PreserveLength
		return fmt.Sprintf("this.%s.fullnametransformer(%t)", col.Column, pl), nil
	case "chinese_first_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "chinese_last_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "chinese_name":
		return fmt.Sprintf("fake(%q)", col.Transformer.Value), nil
	case "phone_number":
		pl := col.Transformer.Config.GetPhoneNumberConfig().PreserveLength
		ef := col.Transformer.Config.GetPhoneNumberConfig().E164Format
		ih := col.Transformer.Config.GetPhoneNumberConfig().IncludeHyphens
		return fmt.Sprintf("this.%s.phonetransformer(%t, %t, %t)", col.Column, pl, ef, ih), nil
	case "int_phone_number":
		pl := col.Transformer.Config.GetIntPhoneNumberConfig().PreserveLength
		return fmt.Sprintf("this.%s.intphonetransformer(%t)", col.Column, pl), nil
	case "uuid":
		ih := col.Transformer.Config.GetUuidConfig().IncludeHyphen
		return fmt.Sprintf("uuidtransformer(%t)", ih), nil
	case "null":
		return "null", nil
	case "random_bool":
		return "randombooltransformer()", nil
	case "random_string":
		pl := col.Transformer.Config.GetRandomStringConfig().PreserveLength
		sl := col.Transformer.Config.GetRandomStringConfig().StrLength
		return fmt.Sprintf(`this.%s.randomstringtransformer(%t, %d)`, col.Column, pl, sl), nil
	case "random_int":
		pl := col.Transformer.Config.GetRandomIntConfig().PreserveLength
		sl := col.Transformer.Config.GetRandomIntConfig().IntLength
		return fmt.Sprintf(`this.%s.randominttransformer(%t, %d)`, col.Column, pl, sl), nil
	case "random_float":
		pl := col.Transformer.Config.GetRandomFloatConfig().PreserveLength
		bd := col.Transformer.Config.GetRandomFloatConfig().DigitsBeforeDecimal
		ad := col.Transformer.Config.GetRandomFloatConfig().DigitsAfterDecimal
		return fmt.Sprintf(`this.%s.randomfloattransformer(%t, %d, %d)`, col.Column, pl, bd, ad), nil
	case "gender":
		ab := col.Transformer.Config.GetGenderConfig().Abbreviate
		return fmt.Sprintf(`gendertransformer(%t)`, ab), nil
	case "utc_timestamp":
		return "utctimestamptransformer()", nil
	case "unix_timestamp":
		return "unixtimestamptransformer()", nil
	case "street_address":
		return "streetaddresstransformer()", nil
	case "city":
		return "citytransformer()", nil
	case "zipcode":
		return "zipcodetransformer()", nil
	case "state":
		return "statetransformer()", nil
	case "full_address":
		return "fulladdresstransformer()", nil
	case "credit_card":
		luhn := true
		return fmt.Sprintf(`creditcardtransformer(%t)`, luhn), nil
	default:
		return "", fmt.Errorf("unsupported transformer")
	}
}
