package shared

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1/mgmtv1alpha1connect"
	sql_manager "github.com/nucleuscloud/neosync/backend/pkg/sqlmanager"
	neosync_benthos "github.com/nucleuscloud/neosync/worker/internal/benthos"

	http_client "github.com/nucleuscloud/neosync/worker/internal/http/client"
	"github.com/spf13/viper"
)

const (
	// The benthos value for null
	NullString = "null"
)

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
	return http_client.NewWithAuth(&apikey)
}

// Generic util method that turns any value into its pointer
func Ptr[T any](val T) *T {
	return &val
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

// Parses the job mappings and returns the unique set of tables.
// Does not include a table if all of the columns are set to null
func GetUniqueTablesFromMappings(mappings []*mgmtv1alpha1.JobMapping) map[string]struct{} {
	groupedMappings := map[string][]*mgmtv1alpha1.JobMapping{}
	for _, mapping := range mappings {
		tableName := sql_manager.BuildTable(mapping.Schema, mapping.Table)
		_, ok := groupedMappings[tableName]
		if ok {
			groupedMappings[tableName] = append(groupedMappings[tableName], mapping)
		} else {
			groupedMappings[tableName] = []*mgmtv1alpha1.JobMapping{mapping}
		}
	}

	filteredTables := map[string]struct{}{}

	for table, mappings := range groupedMappings {
		if !AreAllColsNull(mappings) {
			filteredTables[table] = struct{}{}
		}
	}
	return filteredTables
}

// Checks each transformer source in the set of mappings and returns true if they are all source=null
func AreAllColsNull(mappings []*mgmtv1alpha1.JobMapping) bool {
	for _, col := range mappings {
		if col.Transformer.Source != mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_GENERATE_NULL {
			return false
		}
	}
	return true
}

type RedisConfig struct {
	Url    string
	Kind   string
	Master *string
	Tls    *RedisTlsConfig
}

type RedisTlsConfig struct {
	Enabled               bool
	SkipCertVerify        bool
	EnableRenegotiation   bool
	RootCertAuthority     *string
	RootCertAuthorityFile *string
}

func GetRedisConfig() *RedisConfig {
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
	return &RedisConfig{
		Url:    redisUrl,
		Kind:   kind,
		Master: master,
		Tls: &RedisTlsConfig{
			Enabled:               viper.GetBool("REDIS_TLS_ENABLED"),
			SkipCertVerify:        viper.GetBool("REDIS_TLS_SKIP_CERT_VERIFY"),
			EnableRenegotiation:   viper.GetBool("REDIS_TLS_ENABLE_RENEGOTIATION"),
			RootCertAuthority:     &rootCertAuthority,
			RootCertAuthorityFile: &rootCertAuthorityFile,
		},
	}
}

func BuildBenthosRedisTlsConfig(redisConfig *RedisConfig) *neosync_benthos.RedisTlsConfig {
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
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		connectionId = jobSourceConfig.Generate.GetFkSourceConnectionId()
	default:
		return nil, errors.New("unsupported job source options type")
	}
	sourceConnection, err := GetConnectionById(ctx, connclient, connectionId)
	if err != nil {
		return nil, fmt.Errorf("unable to get connection by id (%s): %w", connectionId, err)
	}
	return sourceConnection, nil
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
