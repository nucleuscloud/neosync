package shared

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sqlmanager_shared "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager/shared"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/pkg/benthos"

	benthosbuilder_shared "github.com/nucleuscloud/neosync/internal/benthos/benthos-builder/shared"
	http_client "github.com/nucleuscloud/neosync/internal/http/client"
	neosync_redis "github.com/nucleuscloud/neosync/internal/redis"
	"github.com/spf13/viper"
)

const (
	// The benthos value for null
	NullString = "null"

	runContext_ExternalId_BenthosConfig       = "benthosconfig"
	runContext_ExternalId_PostTableSyncConfig = "posttablesync"
	runContext_ExternalId_ConnectionIds       = "tablesync-connectionids"
	runContext_ExternalId_QueryContext        = "tablesync-querycontext"
)

func GetBenthosConfigExternalId(identifier string) string {
	return fmt.Sprintf("%s-%s", runContext_ExternalId_BenthosConfig, identifier)
}

func GetConnectionIdsExternalId() string {
	return runContext_ExternalId_ConnectionIds
}

func GetPostTableSyncConfigExternalId(identifier string) string {
	return fmt.Sprintf("%s-%s", runContext_ExternalId_PostTableSyncConfig, identifier)
}

func GetQueryContextExternalId(identifier string) string {
	return fmt.Sprintf("%s-%s", runContext_ExternalId_QueryContext, identifier)
}

type PostTableSyncConfig struct {
	DestinationConfigs map[string]*PostTableSyncDestConfig `json:"destinationConfigs"`
}

type PostTableSyncDestConfig struct {
	Statements []string `json:"statements"` // statements to run
}

// General workflow metadata struct that is intended to be common across activities
type WorkflowMetadata struct {
	WorkflowId string
	RunId      string
}

// Holds the environment variable name and the connection id that should replace it at runtime when the Sync activity is launched
type BenthosDsn struct {
	EnvVarKey string
	// Neosync Connection Id
	ConnectionId string
}

// Returns the neosync url found in the environment, otherwise defaults to localhost
func GetNeosyncUrl() string {
	neosyncUrl := viper.GetString("NEOSYNC_URL")
	if neosyncUrl == "" {
		return "http://localhost:8080"
	}
	return neosyncUrl
}

// Returns an instance of *http.Client that includes the Neosync API Token if one was found in the environment
func GetNeosyncHttpClient() *http.Client {
	apikey := viper.GetString("NEOSYNC_API_KEY")
	return http_client.NewWithBearerAuth(&apikey)
}

// Generic util method that turns any value into its pointer
func Ptr[T any](val T) *T {
	return &val
}

// Parses the job and returns the unique set of schemas.
func GetUniqueSchemasFromJob(job *mgmtv1alpha1.Job) []string {
	switch jobSourceConfig := job.Source.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		uniqueSchemas := map[string]struct{}{}
		for _, schema := range jobSourceConfig.AiGenerate.Schemas {
			uniqueSchemas[schema.Schema] = struct{}{}
		}
		schemas := []string{}
		for s := range uniqueSchemas {
			schemas = append(schemas, s)
		}
		return schemas
	default:
		return GetUniqueSchemasFromMappings(job.GetMappings())
	}
}

// Parses the job mappings and returns the unique set of schemas found
func GetUniqueSchemasFromMappings(mappings []*mgmtv1alpha1.JobMapping) []string {
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

// Parses the job and returns the unique set of tables.
func GetUniqueTablesMapFromJob(job *mgmtv1alpha1.Job) map[string]struct{} {
	switch jobSourceConfig := job.Source.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		uniqueTables := map[string]struct{}{}
		for _, schema := range jobSourceConfig.AiGenerate.Schemas {
			for _, table := range schema.Tables {
				uniqueTables[sqlmanager_shared.BuildTable(schema.Schema, table.Table)] = struct{}{}
			}
		}
		return uniqueTables
	default:
		return GetUniqueTablesFromMappings(job.GetMappings())
	}
}

// Parses the job mappings and returns the unique set of tables.
func GetUniqueTablesFromMappings(mappings []*mgmtv1alpha1.JobMapping) map[string]struct{} {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}
	for _, mapping := range mappings {
		tableName := sqlmanager_shared.BuildTable(mapping.Schema, mapping.Table)
		_, ok := groupedMappings[tableName]
		if ok {
			groupedMappings[tableName] = append(groupedMappings[tableName], mapping)
		} else {
			groupedMappings[tableName] = []*mgmtv1alpha1.JobMapping{mapping}
		}
	}

	filteredTables := map[string]struct{}{}

	for table := range groupedMappings {
		filteredTables[table] = struct{}{}
	}
	return filteredTables
}

func GetRedisConfig() *neosync_redis.RedisConfig {
	redisUrl := viper.GetString("REDIS_URL")
	if redisUrl == "" {
		return nil
	}

	kindEv := viper.GetString("REDIS_KIND")
	masterEv := viper.GetString("REDIS_MASTER")
	rootCertAuthority := viper.GetString("REDIS_TLS_ROOT_CERT_AUTHORITY")
	rootCertAuthorityFile := viper.GetString("REDIS_TLS_ROOT_CERT_AUTHORITY_FILE")
	var kind string
	var master *string
	if kindEv != "" {
		kind = kindEv
	} else {
		kind = "simple"
	}
	if masterEv != "" {
		master = &masterEv
	}
	return &neosync_redis.RedisConfig{
		Url:    redisUrl,
		Kind:   kind,
		Master: master,
		Tls: &neosync_redis.RedisTlsConfig{
			Enabled:               viper.GetBool("REDIS_TLS_ENABLED"),
			SkipCertVerify:        viper.GetBool("REDIS_TLS_SKIP_CERT_VERIFY"),
			EnableRenegotiation:   viper.GetBool("REDIS_TLS_ENABLE_RENEGOTIATION"),
			RootCertAuthority:     &rootCertAuthority,
			RootCertAuthorityFile: &rootCertAuthorityFile,
		},
	}
}

func BuildBenthosRedisTlsConfig(redisConfig *neosync_redis.RedisConfig) *neosync_benthos.RedisTlsConfig {
	var tls *neosync_benthos.RedisTlsConfig
	if redisConfig.Tls != nil && redisConfig.Tls.Enabled {
		tls = &neosync_benthos.RedisTlsConfig{
			Enabled:             redisConfig.Tls.Enabled,
			SkipCertVerify:      redisConfig.Tls.SkipCertVerify,
			EnableRenegotiation: redisConfig.Tls.EnableRenegotiation,
			RootCas:             redisConfig.Tls.RootCertAuthority,
			RootCasFile:         redisConfig.Tls.RootCertAuthorityFile,
		}
	}
	return tls
}

func GetJobSourceConnection(
	ctx context.Context,
	jobSource *mgmtv1alpha1.JobSource,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
) (*mgmtv1alpha1.Connection, error) {
	var connectionId string
	switch jobSourceConfig := jobSource.GetOptions().GetConfig().(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		connectionId = jobSourceConfig.Postgres.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		connectionId = jobSourceConfig.Mysql.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mssql:
		connectionId = jobSourceConfig.Mssql.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		connectionId = jobSourceConfig.Generate.GetFkSourceConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		connectionId = jobSourceConfig.AiGenerate.GetAiConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Mongodb:
		connectionId = jobSourceConfig.Mongodb.GetConnectionId()
	case *mgmtv1alpha1.JobSourceOptions_Dynamodb:
		connectionId = jobSourceConfig.Dynamodb.GetConnectionId()
	default:
		return nil, fmt.Errorf("unsupported job source options type for job source connection: %T", jobSourceConfig)
	}
	sourceConnection, err := GetConnectionById(ctx, connclient, connectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id (%s): %w", connectionId, err)
	}
	return sourceConnection, nil
}

// Returns the connection type as a string
// Should only be used for logging
func GetConnectionType(connection *mgmtv1alpha1.Connection) string {
	connectiontype, err := benthosbuilder_shared.GetConnectionType(connection)
	if err != nil {
		return "unknown"
	}
	return string(connectiontype)
}

func GetConnectionById(
	ctx context.Context,
	connclient mgmtv1alpha1connect.ConnectionServiceClient,
	connectionId string,
) (*mgmtv1alpha1.Connection, error) {
	getConnResp, err := connclient.GetConnection(ctx, connect.NewRequest(&mgmtv1alpha1.GetConnectionRequest{
		Id: connectionId,
	}))
	if err != nil {
		return nil, err
	}
	return getConnResp.Msg.Connection, nil
}

type SqlJobDestinationOpts struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool
	InitSchema           bool
}

func GetSqlJobDestinationOpts(
	options *mgmtv1alpha1.JobDestinationOptions,
) (*SqlJobDestinationOpts, error) {
	if options == nil {
		return &SqlJobDestinationOpts{}, nil
	}
	switch opts := options.GetConfig().(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		return &SqlJobDestinationOpts{
			TruncateBeforeInsert: opts.PostgresOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			TruncateCascade:      opts.PostgresOptions.GetTruncateTable().GetCascade(),
			InitSchema:           opts.PostgresOptions.GetInitTableSchema(),
		}, nil
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		return &SqlJobDestinationOpts{
			TruncateBeforeInsert: opts.MysqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			InitSchema:           opts.MysqlOptions.GetInitTableSchema(),
		}, nil
	case *mgmtv1alpha1.JobDestinationOptions_MssqlOptions:
		return &SqlJobDestinationOpts{
			TruncateBeforeInsert: opts.MssqlOptions.GetTruncateTable().GetTruncateBeforeInsert(),
			InitSchema:           opts.MssqlOptions.GetInitTableSchema(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported job destination options type: %T", opts)
	}
}
