package neosync_benthos

type BenthosConfig struct {
	StreamConfig `json:",inline" yaml:",inline"`
}

type StreamConfig struct {
	Logger   *LoggerConfig   `json:"logger"            yaml:"logger,omitempty"`
	Input    *InputConfig    `json:"input"             yaml:"input"`
	Pipeline *PipelineConfig `json:"pipeline"          yaml:"pipeline"`
	Output   *OutputConfig   `json:"output"            yaml:"output"`
	Metrics  *Metrics        `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

type LoggerConfig struct {
	Level        string `json:"level"         yaml:"level"`
	AddTimestamp bool   `json:"add_timestamp" yaml:"add_timestamp"`
}

type Metrics struct {
	OtelCollector *MetricsOtelCollector `json:"otel_collector,omitempty" yaml:"otel_collector,omitempty"`
	Mapping       string                `json:"mapping,omitempty"        yaml:"mapping,omitempty"`
}

type MetricsOtelCollector struct {
}

type InputConfig struct {
	Label  string `json:"label"   yaml:"label"`
	Inputs `       json:",inline" yaml:",inline"`
}

type Inputs struct {
	PooledSqlRaw          *InputPooledSqlRaw     `json:"pooled_sql_raw,omitempty"          yaml:"pooled_sql_raw,omitempty"`
	Generate              *Generate              `json:"generate,omitempty"                yaml:"generate,omitempty"`
	OpenAiGenerate        *OpenAiGenerate        `json:"openai_generate,omitempty"         yaml:"openai_generate,omitempty"`
	PooledMongoDB         *InputMongoDb          `json:"pooled_mongodb,omitempty"          yaml:"pooled_mongodb,omitempty"`
	AwsDynamoDB           *InputAwsDynamoDB      `json:"aws_dynamodb,omitempty"            yaml:"aws_dynamodb,omitempty"`
	NeosyncConnectionData *NeosyncConnectionData `json:"neosync_connection_data,omitempty" yaml:"neosync_connection_data,omitempty"`
	Broker                *InputBrokerConfig     `json:"broker,omitempty"                  yaml:"broker,omitempty"`
}

type NeosyncConnectionData struct {
	ConnectionId   string  `json:"connection_id"        yaml:"connection_id"`
	ConnectionType string  `json:"connection_type"      yaml:"connection_type"`
	JobId          *string `json:"job_id,omitempty"     yaml:"job_id,omitempty"`
	JobRunId       *string `json:"job_run_id,omitempty" yaml:"job_run_id,omitempty"`
	Schema         string  `json:"schema"               yaml:"schema"`
	Table          string  `json:"table"                yaml:"table"`
}

type InputAwsDynamoDB struct {
	Table          string  `json:"table"           yaml:"table"`
	Where          *string `json:"where,omitempty" yaml:"where,omitempty"`
	ConsistentRead bool    `json:"consistent_read" yaml:"consistent_read"`

	Region   string `json:"region,omitempty"   yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type OutputAwsDynamoDB struct {
	Table          string            `json:"table"                      yaml:"table"`
	JsonMapColumns map[string]string `json:"json_map_columns,omitempty" yaml:"json_map_columns,omitempty"`

	Region   string `json:"region,omitempty"   yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`

	MaxInFlight *int      `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
	Batching    *Batching `json:"batching,omitempty"      yaml:"batching,omitempty"`
}

type InputMongoDb struct {
	ConnectionId    string         `json:"connection_id"               yaml:"connection_id"`
	Database        string         `json:"database"                    yaml:"database"`
	Username        string         `json:"username,omitempty"          yaml:"username,omitempty"`
	Password        string         `json:"password,omitempty"          yaml:"password,omitempty"`
	Operation       *string        `json:"operation,omitempty"         yaml:"operation,omitempty"`
	Collection      string         `json:"collection"                  yaml:"collection"`
	JsonMarshalMode *string        `json:"json_marshal_mode,omitempty" yaml:"json_marshal_mode,omitempty"`
	Query           string         `json:"query"                       yaml:"query"`
	AutoReplayNacks *bool          `json:"auto_replay_nacks,omitempty" yaml:"auto_replay_nacks,omitempty"`
	BatchSize       *int32         `json:"batch_size,omitempty"        yaml:"batch_size,omitempty"`
	Sort            map[string]int `json:"sort,omitempty"              yaml:"sort,omitempty"`
	Limit           *int32         `json:"limit,omitempty"             yaml:"limit,omitempty"`
}

type OutputMongoDb struct {
	ConnectionId string             `json:"connection_id"           yaml:"connection_id"`
	Database     string             `json:"database"                yaml:"database"`
	Username     string             `json:"username,omitempty"      yaml:"username,omitempty"`
	Password     string             `json:"password,omitempty"      yaml:"password,omitempty"`
	Operation    string             `json:"operation"               yaml:"operation"`
	Collection   string             `json:"collection"              yaml:"collection"`
	DocumentMap  string             `json:"document_map"            yaml:"document_map"`
	FilterMap    string             `json:"filter_map"              yaml:"filter_map"`
	HintMap      string             `json:"hint_map"                yaml:"hint_map"`
	Upsert       bool               `json:"upsert"                  yaml:"upsert"`
	MaxInFlight  *int               `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
	Batching     *Batching          `json:"batching,omitempty"      yaml:"batching,omitempty"`
	WriteConcern *MongoWriteConcern `json:"write_concern,omitempty" yaml:"write_concern,omitempty"`
}

type MongoWriteConcern struct {
	W        string `json:"w,omitempty"         yaml:"w,omitempty"`
	J        string `json:"j,omitempty"         yaml:"j,omitempty"`
	WTimeout string `json:"w_timeout,omitempty" yaml:"w_timeout,omitempty"`
}

type OpenAiGenerate struct {
	ApiUrl     string   `json:"api_url"               yaml:"api_url"`
	ApiKey     string   `json:"api_key"               yaml:"api_key"`
	UserPrompt *string  `json:"user_prompt,omitempty" yaml:"user_prompt,omitempty"`
	Columns    []string `json:"columns"               yaml:"columns"`
	DataTypes  []string `json:"data_types"            yaml:"data_types"`
	Model      string   `json:"model"                 yaml:"model"`
	Count      int      `json:"count"                 yaml:"count"`
	BatchSize  int      `json:"batch_size"            yaml:"batch_size"`
}

type Generate struct {
	Mapping string `json:"mapping" yaml:"mapping"`

	Interval  string `json:"interval"             yaml:"interval"`
	Count     int    `json:"count"                yaml:"count"`
	BatchSize *int   `json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
}

type InputPooledSqlRaw struct {
	ConnectionId string `json:"connection_id"         yaml:"connection_id"`
	Query        string `json:"query"                 yaml:"query"`
	PagedQuery   string `json:"paged_query,omitempty" yaml:"paged_query,omitempty"`

	ExpectedTotalRows *int     `json:"expected_total_rows,omitempty" yaml:"expected_total_rows,omitempty"`
	OrderByColumns    []string `json:"order_by_columns,omitempty"    yaml:"order_by_columns,omitempty"`
}

type PipelineConfig struct {
	Threads    int               `json:"threads"    yaml:"threads"`
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
}

type ProcessorConfig struct {
	Mutation                  *string                          `json:"mutation,omitempty"                    yaml:"mutation,omitempty"`
	NeosyncJavascript         *NeosyncJavascriptConfig         `json:"neosync_javascript,omitempty"          yaml:"neosync_javascript,omitempty"`
	Branch                    *BranchConfig                    `json:"branch,omitempty"                      yaml:"branch,omitempty"`
	Mapping                   *string                          `json:"mapping,omitempty"                     yaml:"mapping,omitempty"`
	Redis                     *RedisProcessorConfig            `json:"redis,omitempty"                       yaml:"redis,omitempty"`
	Error                     *ErrorProcessorConfig            `json:"error,omitempty"                       yaml:"error,omitempty"`
	Catch                     []*ProcessorConfig               `json:"catch,omitempty"                       yaml:"catch,omitempty"`
	NeosyncDefaultTransformer *NeosyncDefaultTransformerConfig `json:"neosync_default_transformer,omitempty" yaml:"neosync_default_transformer,omitempty"`
}

type NeosyncDefaultTransformerConfig struct {
	JobSourceOptionsString string   `json:"job_source_options_string" yaml:"job_source_options_string"`
	MappedKeys             []string `json:"mapped_keys"               yaml:"mapped_keys"`
}

type NeosyncJavascriptConfig struct {
	Code string `json:"code" yaml:"code"`
}

type ErrorProcessorConfig struct {
	ErrorMsg string `json:"error_msg" yaml:"error_msg"`
}

type RedisProcessorConfig struct {
	Command     string `json:"command"          yaml:"command"`
	ArgsMapping string `json:"args_mapping"     yaml:"args_mapping"`
}

type BranchConfig struct {
	Processors []ProcessorConfig `json:"processors"            yaml:"processors"`
	RequestMap *string           `json:"request_map,omitempty" yaml:"request_map,omitempty"`
	ResultMap  *string           `json:"result_map,omitempty"  yaml:"result_map,omitempty"`
}

type OutputConfig struct {
	Label      string `json:"label"                yaml:"label"`
	Outputs    `                  json:",inline"              yaml:",inline"`
	Processors []ProcessorConfig `json:"processors,omitempty" yaml:"processors,omitempty"`
}

type Outputs struct {
	PooledSqlInsert *PooledSqlInsert       `json:"pooled_sql_insert,omitempty" yaml:"pooled_sql_insert,omitempty"`
	PooledSqlUpdate *PooledSqlUpdate       `json:"pooled_sql_update,omitempty" yaml:"pooled_sql_update,omitempty"`
	AwsS3           *AwsS3Insert           `json:"aws_s3,omitempty"            yaml:"aws_s3,omitempty"`
	GcpCloudStorage *GcpCloudStorageOutput `json:"gcp_cloud_storage,omitempty" yaml:"gcp_cloud_storage,omitempty"`
	Retry           *RetryConfig           `json:"retry,omitempty"             yaml:"retry,omitempty"`
	Broker          *OutputBrokerConfig    `json:"broker,omitempty"            yaml:"broker,omitempty"`
	Fallback        []Outputs              `json:"fallback,omitempty"          yaml:"fallback,omitempty"`
	RedisHashOutput *RedisHashOutputConfig `json:"redis_hash_output,omitempty" yaml:"redis_hash_output,omitempty"`
	Error           *ErrorOutputConfig     `json:"error,omitempty"             yaml:"error,omitempty"`
	PooledMongoDB   *OutputMongoDb         `json:"pooled_mongodb,omitempty"    yaml:"pooled_mongodb,omitempty"`
	AwsDynamoDB     *OutputAwsDynamoDB     `json:"aws_dynamodb,omitempty"      yaml:"aws_dynamodb,omitempty"`
}
type ErrorOutputConfig struct {
	ErrorMsg      string    `json:"error_msg"          yaml:"error_msg"`
	IsGenerateJob bool      `json:"is_generate_job"    yaml:"is_generate_job"`
	Batching      *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type RedisHashOutputConfig struct {
	Key            string `json:"key"                     yaml:"key"`
	WalkMetadata   bool   `json:"walk_metadata"           yaml:"walk_metadata"`
	WalkJsonObject bool   `json:"walk_json_object"        yaml:"walk_json_object"`
	FieldsMapping  string `json:"fields_mapping"          yaml:"fields_mapping"`
	MaxInFlight    *int   `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
}

type RetryConfig struct {
	Output            OutputConfig `json:"output"  yaml:"output"`
	InlineRetryConfig `             json:",inline" yaml:",inline"`
}

type InlineRetryConfig struct {
	MaxRetries uint64  `json:"max_retries" yaml:"max_retries"`
	Backoff    Backoff `json:"backoff"     yaml:"backoff"`
}

type Backoff struct {
	InitialInterval string `json:"initial_interval,omitempty" yaml:"initial_interval,omitempty"`
	MaxInterval     string `json:"max_interval,omitempty"     yaml:"max_interval,omitempty"`
	MaxElapsedTime  string `json:"max_elapsed_time,omitempty" yaml:"max_elapsed_time,omitempty"`
}

type PooledSqlUpdate struct {
	ConnectionId             string    `json:"connection_id"               yaml:"connection_id"`
	Schema                   string    `json:"schema"                      yaml:"schema"`
	Table                    string    `json:"table"                       yaml:"table"`
	Columns                  []string  `json:"columns"                     yaml:"columns"`
	WhereColumns             []string  `json:"where_columns"               yaml:"where_columns"`
	SkipForeignKeyViolations bool      `json:"skip_foreign_key_violations" yaml:"skip_foreign_key_violations"`
	Batching                 *Batching `json:"batching,omitempty"          yaml:"batching,omitempty"`
	MaxInFlight              int       `json:"max_in_flight,omitempty"     yaml:"max_in_flight,omitempty"`
}

type ColumnDefaultProperties struct {
	NeedsReset            bool `json:"needs_reset"             yaml:"needs_reset"`
	NeedsOverride         bool `json:"needs_override"          yaml:"needs_override"`
	HasDefaultTransformer bool `json:"has_default_transformer" yaml:"has_default_transformer"`
}

type PooledSqlInsert struct {
	ConnectionId                string    `json:"connection_id"                  yaml:"connection_id"`
	Schema                      string    `json:"schema"                         yaml:"schema"`
	Table                       string    `json:"table"                          yaml:"table"`
	PrimaryKeyColumns           []string  `json:"primary_key_columns"            yaml:"primary_key_columns"`
	ColumnUpdatesDisallowed     []string  `json:"column_updates_disallowed"      yaml:"column_updates_disallowed"`
	OnConflictDoNothing         bool      `json:"on_conflict_do_nothing"         yaml:"on_conflict_do_nothing"`
	OnConflictDoUpdate          bool      `json:"on_conflict_do_update"          yaml:"on_conflict_do_update"`
	TruncateOnRetry             bool      `json:"truncate_on_retry"              yaml:"truncate_on_retry"`
	SkipForeignKeyViolations    bool      `json:"skip_foreign_key_violations"    yaml:"skip_foreign_key_violations"`
	ShouldOverrideColumnDefault bool      `json:"should_override_column_default" yaml:"should_override_column_default"`
	Batching                    *Batching `json:"batching,omitempty"             yaml:"batching,omitempty"`
	Prefix                      *string   `json:"prefix,omitempty"               yaml:"prefix,omitempty"`
	Suffix                      *string   `json:"suffix,omitempty"               yaml:"suffix,omitempty"`
	MaxInFlight                 int       `json:"max_in_flight,omitempty"        yaml:"max_in_flight,omitempty"`
}

type AwsS3Insert struct {
	Bucket       string    `json:"bucket"                  yaml:"bucket"`
	MaxInFlight  int       `json:"max_in_flight"           yaml:"max_in_flight"`
	Path         string    `json:"path"                    yaml:"path"`
	Batching     *Batching `json:"batching,omitempty"      yaml:"batching,omitempty"`
	Timeout      string    `json:"timeout,omitempty"       yaml:"timeout,omitempty"`
	StorageClass string    `json:"storage_class,omitempty" yaml:"storage_class,omitempty"`
	ContentType  string    `json:"content_type,omitempty"  yaml:"content_type,omitempty"`

	Region   string `json:"region,omitempty"   yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type AwsCredentials struct {
	Profile        string `json:"profile,omitempty"          yaml:"profile,omitempty"`
	Id             string `json:"id,omitempty"               yaml:"id,omitempty"`
	Secret         string `json:"secret,omitempty"           yaml:"secret,omitempty"`
	Token          string `json:"token,omitempty"            yaml:"token,omitempty"`
	FromEc2Role    bool   `json:"from_ec2_role,omitempty"    yaml:"from_ec2_role,omitempty"`
	Role           string `json:"role,omitempty"             yaml:"role,omitempty"`
	RoleExternalId string `json:"role_external_id,omitempty" yaml:"role_external_id,omitempty"`
}

type GcpCloudStorageOutput struct {
	Bucket      string    `json:"bucket"             yaml:"bucket"`
	Path        string    `json:"path"               yaml:"path"`
	MaxInFlight int       `json:"max_in_flight"      yaml:"max_in_flight"`
	Batching    *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`

	ContentType     *string `json:"content_type,omitempty"     yaml:"content_type,omitempty"`
	ContentEncoding *string `json:"content_encoding,omitempty" yaml:"content_encoding,omitempty"`
	CollisionMode   *string `json:"collision_mode,omitempty"   yaml:"collision_mode,omitempty"`
	ChunkSize       *int    `json:"chunk_size,omitempty"       yaml:"chunk_size,omitempty"`
	Timeout         *string `json:"timeout,omitempty"          yaml:"timeout,omitempty"`
}

type Batching struct {
	Count      int               `json:"count"      yaml:"count"`
	ByteSize   int               `json:"byte_size"  yaml:"byte_size"`
	Period     string            `json:"period"     yaml:"period"`
	Check      string            `json:"check"      yaml:"check"`
	Processors []*BatchProcessor `json:"processors" yaml:"processors"`
}

type BatchProcessor struct {
	Archive        *ArchiveProcessor     `json:"archive,omitempty"          yaml:"archive,omitempty"`
	Compress       *CompressProcessor    `json:"compress,omitempty"         yaml:"compress,omitempty"`
	NeosyncToJson  *NeosyncToJsonConfig  `json:"neosync_to_json,omitempty"  yaml:"neosync_to_json,omitempty"`
	NeosyncToPgx   *NeosyncToPgxConfig   `json:"neosync_to_pgx,omitempty"   yaml:"neosync_to_pgx,omitempty"`
	NeosyncToMysql *NeosyncToMysqlConfig `json:"neosync_to_mysql,omitempty" yaml:"neosync_to_mysql,omitempty"`
	NeosyncToMssql *NeosyncToMssqlConfig `json:"neosync_to_mssql,omitempty" yaml:"neosync_to_mssql,omitempty"`
}

type NeosyncToPgxConfig struct {
	Columns                 []string                            `json:"columns"                   yaml:"columns"`
	ColumnDataTypes         map[string]string                   `json:"column_data_types"         yaml:"column_data_types"`
	ColumnDefaultProperties map[string]*ColumnDefaultProperties `json:"column_default_properties" yaml:"column_default_properties"`
}

type NeosyncToMysqlConfig struct {
	Columns                 []string                            `json:"columns"                   yaml:"columns"`
	ColumnDataTypes         map[string]string                   `json:"column_data_types"         yaml:"column_data_types"`
	ColumnDefaultProperties map[string]*ColumnDefaultProperties `json:"column_default_properties" yaml:"column_default_properties"`
}

type NeosyncToMssqlConfig struct {
	Columns                 []string                            `json:"columns"                   yaml:"columns"`
	ColumnDataTypes         map[string]string                   `json:"column_data_types"         yaml:"column_data_types"`
	ColumnDefaultProperties map[string]*ColumnDefaultProperties `json:"column_default_properties" yaml:"column_default_properties"`
}

type NeosyncToJsonConfig struct{}

type ArchiveProcessor struct {
	Format string  `json:"format"         yaml:"format"`
	Path   *string `json:"path,omitempty" yaml:"path,omitempty"`
}

type CompressProcessor struct {
	Algorithm string `json:"algorithm" yaml:"algorithm"`
}

type OutputBrokerConfig struct {
	Pattern string    `json:"pattern" yaml:"pattern"`
	Outputs []Outputs `json:"outputs" yaml:"outputs"`
}

type InputBrokerConfig struct {
	Pattern string   `json:"pattern" yaml:"pattern"`
	Inputs  []Inputs `json:"inputs"  yaml:"inputs"`
}
