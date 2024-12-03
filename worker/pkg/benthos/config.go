package neosync_benthos

type BenthosConfig struct {
	// HTTP         HTTPConfig `json:"http" yaml:"http"`
	StreamConfig `json:",inline" yaml:",inline"`
}

type HTTPConfig struct {
	Address string `json:"address" yaml:"address"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
	// RootPath       string                     `json:"root_path" yaml:"root_path"`
	// DebugEndpoints bool                       `json:"debug_endpoints" yaml:"debug_endpoints"`
	// CertFile       string                     `json:"cert_file" yaml:"cert_file"`
	// KeyFile        string                     `json:"key_file" yaml:"key_file"`
	// CORS           httpserver.CORSConfig      `json:"cors" yaml:"cors"`
	// BasicAuth      httpserver.BasicAuthConfig `json:"basic_auth" yaml:"basic_auth"`
}

type StreamConfig struct {
	Logger         *LoggerConfig          `json:"logger" yaml:"logger,omitempty"`
	Input          *InputConfig           `json:"input" yaml:"input"`
	Buffer         *BufferConfig          `json:"buffer,omitempty" yaml:"buffer,omitempty"`
	Pipeline       *PipelineConfig        `json:"pipeline" yaml:"pipeline"`
	Output         *OutputConfig          `json:"output" yaml:"output"`
	CacheResources []*CacheResourceConfig `json:"cache_resources,omitempty" yaml:"cache_resources,omitempty"`
	Metrics        *Metrics               `json:"metrics,omitempty" yaml:"metrics,omitempty"`
}

type LoggerConfig struct {
	Level        string `json:"level" yaml:"level"`
	AddTimestamp bool   `json:"add_timestamp" yaml:"add_timestamp"`
}

type Metrics struct {
	OtelCollector *MetricsOtelCollector `json:"otel_collector,omitempty" yaml:"otel_collector,omitempty"`
	Mapping       string                `json:"mapping,omitempty" yaml:"mapping,omitempty"`
}

type MetricsOtelCollector struct {
}
type MetricsStatsD struct {
	Address     string `json:"address" yaml:"address"`
	FlushPeriod string `json:"flush_period,omitempty" yaml:"flush_period,omitempty"`
	TagFormat   string `json:"tag_format,omitempty" yaml:"tag_format,omitempty"`
}

type CacheResourceConfig struct {
	Label string            `json:"label" yaml:"label"`
	Redis *RedisCacheConfig `json:"redis,omitempty" yaml:"redis,omitempty"`
}

type RedisCacheConfig struct {
	Url    string  `json:"url" yaml:"url"`
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
}

type InputConfig struct {
	Label  string `json:"label" yaml:"label"`
	Inputs `json:",inline" yaml:",inline"`
}

type Inputs struct {
	SqlSelect             *SqlSelect             `json:"sql_select,omitempty" yaml:"sql_select,omitempty"`
	PooledSqlRaw          *InputPooledSqlRaw     `json:"pooled_sql_raw,omitempty" yaml:"pooled_sql_raw,omitempty"`
	Generate              *Generate              `json:"generate,omitempty" yaml:"generate,omitempty"`
	OpenAiGenerate        *OpenAiGenerate        `json:"openai_generate,omitempty" yaml:"openai_generate,omitempty"`
	MongoDB               *InputMongoDb          `json:"mongodb,omitempty" yaml:"mongodb,omitempty"`
	PooledMongoDB         *InputMongoDb          `json:"pooled_mongodb,omitempty" yaml:"pooled_mongodb,omitempty"`
	AwsDynamoDB           *InputAwsDynamoDB      `json:"aws_dynamodb,omitempty" yaml:"aws_dynamodb,omitempty"`
	NeosyncConnectionData *NeosyncConnectionData `json:"neosync_connection_data,omitempty" yaml:"neosync_connection_data,omitempty"`
}

type NeosyncConnectionData struct {
	ConnectionId   string  `json:"connection_id" yaml:"connection_id"`
	ConnectionType string  `json:"connection_type" yaml:"connection_type"`
	JobId          *string `json:"job_id,omitempty" yaml:"job_id,omitempty"`
	JobRunId       *string `json:"job_run_id,omitempty" yaml:"job_run_id,omitempty"`
	Schema         string  `json:"schema" yaml:"schema"`
	Table          string  `json:"table" yaml:"table"`
}

type InputAwsDynamoDB struct {
	Table          string  `json:"table" yaml:"table"`
	Where          *string `json:"where,omitempty" yaml:"where,omitempty"`
	ConsistentRead bool    `json:"consistent_read" yaml:"consistent_read"`

	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type OutputAwsDynamoDB struct {
	Table          string            `json:"table" yaml:"table"`
	JsonMapColumns map[string]string `json:"json_map_columns,omitempty" yaml:"json_map_columns,omitempty"`

	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`

	MaxInFlight *int      `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
	Batching    *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type InputMongoDb struct {
	ConnectionId    string         `json:"connection_id" yaml:"connection_id"`
	Database        string         `json:"database" yaml:"database"`
	Username        string         `json:"username,omitempty" yaml:"username,omitempty"`
	Password        string         `json:"password,omitempty" yaml:"password,omitempty"`
	Operation       *string        `json:"operation,omitempty" yaml:"operation,omitempty"`
	Collection      string         `json:"collection" yaml:"collection"`
	JsonMarshalMode *string        `json:"json_marshal_mode,omitempty" yaml:"json_marshal_mode,omitempty"`
	Query           string         `json:"query" yaml:"query"`
	AutoReplayNacks *bool          `json:"auto_replay_nacks,omitempty" yaml:"auto_replay_nacks,omitempty"`
	BatchSize       *int32         `json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
	Sort            map[string]int `json:"sort,omitempty" yaml:"sort,omitempty"`
	Limit           *int32         `json:"limit,omitempty" yaml:"limit,omitempty"`
}

type OutputMongoDb struct {
	ConnectionId string             `json:"connection_id" yaml:"connection_id"`
	Database     string             `json:"database" yaml:"database"`
	Username     string             `json:"username,omitempty" yaml:"username,omitempty"`
	Password     string             `json:"password,omitempty" yaml:"password,omitempty"`
	Operation    string             `json:"operation" yaml:"operation"`
	Collection   string             `json:"collection" yaml:"collection"`
	DocumentMap  string             `json:"document_map" yaml:"document_map"`
	FilterMap    string             `json:"filter_map" yaml:"filter_map"`
	HintMap      string             `json:"hint_map" yaml:"hint_map"`
	Upsert       bool               `json:"upsert" yaml:"upsert"`
	MaxInFlight  *int               `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
	Batching     *Batching          `json:"batching,omitempty" yaml:"batching,omitempty"`
	WriteConcern *MongoWriteConcern `json:"write_concern,omitempty" yaml:"write_concern,omitempty"`
}

type MongoWriteConcern struct {
	W        string `json:"w,omitempty" yaml:"w,omitempty"`
	J        string `json:"j,omitempty" yaml:"j,omitempty"`
	WTimeout string `json:"w_timeout,omitempty" yaml:"w_timeout,omitempty"`
}

type OpenAiGenerate struct {
	ApiUrl     string   `json:"api_url" yaml:"api_url"`
	ApiKey     string   `json:"api_key" yaml:"api_key"`
	UserPrompt *string  `json:"user_prompt,omitempty" yaml:"user_prompt,omitempty"`
	Columns    []string `json:"columns" yaml:"columns"`
	DataTypes  []string `json:"data_types" yaml:"data_types"`
	Model      string   `json:"model" yaml:"model"`
	Count      int      `json:"count" yaml:"count"`
	BatchSize  int      `json:"batch_size" yaml:"batch_size"`
}

type Generate struct {
	Mapping string `json:"mapping" yaml:"mapping"`

	Interval  string `json:"interval" yaml:"interval"`
	Count     int    `json:"count" yaml:"count"`
	BatchSize *int   `json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
}

type InputPooledSqlRaw struct {
	ConnectionId string `json:"connection_id" yaml:"connection_id"`
	Query        string `json:"query" yaml:"query"`
	ArgsMapping  string `json:"args_mapping,omitempty" yaml:"args_mapping,omitempty"`
}

type SqlSelect struct {
	Driver        string   `json:"driver" yaml:"driver"`
	Dsn           string   `json:"dsn" yaml:"dsn"`
	Table         string   `json:"table" yaml:"table"`
	Columns       []string `json:"columns" yaml:"columns"`
	Where         string   `json:"where,omitempty" yaml:"where,omitempty"`
	ArgsMapping   string   `json:"args_mapping,omitempty" yaml:"args_mapping,omitempty"`
	InitStatement string   `json:"init_statement,omitempty" yaml:"init_statement,omitempty"`
}

type BufferConfig struct{}

type PipelineConfig struct {
	Threads    int               `json:"threads" yaml:"threads"`
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
}

type ProcessorConfig struct {
	Mutation                  *string                          `json:"mutation,omitempty" yaml:"mutation,omitempty"`
	Javascript                *JavascriptConfig                `json:"javascript,omitempty" yaml:"javascript,omitempty"`
	NeosyncJavascript         *NeosyncJavascriptConfig         `json:"neosync_javascript,omitempty" yaml:"neosync_javascript,omitempty"`
	Branch                    *BranchConfig                    `json:"branch,omitempty" yaml:"branch,omitempty"`
	Cache                     *CacheConfig                     `json:"cache,omitempty" yaml:"cache,omitempty"`
	Mapping                   *string                          `json:"mapping,omitempty" yaml:"mapping,omitempty"`
	Redis                     *RedisProcessorConfig            `json:"redis,omitempty" yaml:"redis,omitempty"`
	Error                     *ErrorProcessorConfig            `json:"error,omitempty" yaml:"error,omitempty"`
	Catch                     []*ProcessorConfig               `json:"catch,omitempty" yaml:"catch,omitempty"`
	While                     *WhileProcessorConfig            `json:"while,omitempty" yaml:"while,omitempty"`
	NeosyncDefaultTransformer *NeosyncDefaultTransformerConfig `json:"neosync_default_transformer,omitempty" yaml:"neosync_default_transformer,omitempty"`
}

type NeosyncDefaultTransformerConfig struct {
	JobSourceOptionsString string   `json:"job_source_options_string" yaml:"job_source_options_string"`
	MappedKeys             []string `json:"mapped_keys" yaml:"mapped_keys"`
}

type NeosyncJavascriptConfig struct {
	Code string `json:"code" yaml:"code"`
}

type WhileProcessorConfig struct {
	AtLeastOnce bool               `json:"at_least_once" yaml:"at_least_once"`
	MaxLoops    *int               `json:"max_loops,omitempty" yaml:"max_loops,omitempty"`
	Check       string             `json:"check,omitempty" yaml:"check,omitempty"`
	Processors  []*ProcessorConfig `json:"processors,omitempty" yaml:"processors,omitempty"`
}

type ErrorProcessorConfig struct {
	ErrorMsg string `json:"error_msg" yaml:"error_msg"`
}

type RedisProcessorConfig struct {
	Url         string          `json:"url" yaml:"url"`
	Command     string          `json:"command" yaml:"command"`
	ArgsMapping string          `json:"args_mapping" yaml:"args_mapping"`
	Kind        *string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	Master      *string         `json:"master,omitempty" yaml:"master,omitempty"`
	Tls         *RedisTlsConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
}

type RedisTlsConfig struct {
	Enabled             bool    `json:"enabled" yaml:"enabled"`
	SkipCertVerify      bool    `json:"skip_cert_verify" yaml:"skip_cert_verify"`
	EnableRenegotiation bool    `json:"enable_renegotiation" yaml:"enable_renegotiation"`
	RootCas             *string `json:"root_cas,omitempty" yaml:"root_cas,omitempty"`
	RootCasFile         *string `json:"root_cas_file,omitempty"  yaml:"root_cas_file,omitempty"`
}

type CacheConfig struct {
	Resource string `json:"resource" yaml:"resource"`
	Operator string `json:"operator" yaml:"operator"`
	Key      string `json:"key" yaml:"key"`
	Value    string `json:"value" yaml:"value"`
	Ttl      string `json:"ttl" yaml:"ttl"`
}

type BranchConfig struct {
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
	RequestMap *string           `json:"request_map,omitempty" yaml:"request_map,omitempty"`
	ResultMap  *string           `json:"result_map,omitempty" yaml:"result_map,omitempty"`
}

type JavascriptConfig struct {
	Code string `json:"code" yaml:"code"`
}

type OutputConfig struct {
	Label      string `json:"label" yaml:"label"`
	Outputs    `json:",inline" yaml:",inline"`
	Processors []ProcessorConfig `json:"processors,omitempty" yaml:"processors,omitempty"`
}

type Outputs struct {
	SqlInsert       *SqlInsert             `json:"sql_insert,omitempty" yaml:"sql_insert,omitempty"`
	SqlRaw          *SqlRaw                `json:"sql_raw,omitempty" yaml:"sql_raw,omitempty"`
	PooledSqlInsert *PooledSqlInsert       `json:"pooled_sql_insert,omitempty" yaml:"pooled_sql_insert,omitempty"`
	PooledSqlUpdate *PooledSqlUpdate       `json:"pooled_sql_update,omitempty" yaml:"pooled_sql_update,omitempty"`
	AwsS3           *AwsS3Insert           `json:"aws_s3,omitempty" yaml:"aws_s3,omitempty"`
	GcpCloudStorage *GcpCloudStorageOutput `json:"gcp_cloud_storage,omitempty" yaml:"gcp_cloud_storage,omitempty"`
	Retry           *RetryConfig           `json:"retry,omitempty" yaml:"retry,omitempty"`
	Broker          *OutputBrokerConfig    `json:"broker,omitempty" yaml:"broker,omitempty"`
	DropOn          *DropOnConfig          `json:"drop_on,omitempty" yaml:"drop_on,omitempty"`
	Drop            *DropConfig            `json:"drop,omitempty" yaml:"drop,omitempty"`
	Resource        string                 `json:"resource,omitempty" yaml:"resource,omitempty"`
	Fallback        []Outputs              `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	RedisHashOutput *RedisHashOutputConfig `json:"redis_hash_output,omitempty" yaml:"redis_hash_output,omitempty"`
	Error           *ErrorOutputConfig     `json:"error,omitempty" yaml:"error,omitempty"`
	Switch          *SwitchOutputConfig    `json:"switch,omitempty" yaml:"switch,omitempty"`
	PooledMongoDB   *OutputMongoDb         `json:"pooled_mongodb,omitempty" yaml:"pooled_mongodb,omitempty"`
	AwsDynamoDB     *OutputAwsDynamoDB     `json:"aws_dynamodb,omitempty" yaml:"aws_dynamodb,omitempty"`
}

type SwitchOutputConfig struct {
	RetryUntilSuccess bool               `json:"retry_until_success,omitempty" yaml:"retry_until_success,omitempty"`
	StrictMode        bool               `json:"strict_mode,omitempty" yaml:"strict_mode,omitempty"`
	Cases             []SwitchOutputCase `json:"cases,omitempty" yaml:"cases,omitempty"`
}

type SwitchOutputCase struct {
	Check    string  `json:"check,omitempty" yaml:"check,omitempty"`
	Continue bool    `json:"continue,omitempty" yaml:"continue,omitempty"`
	Output   Outputs `json:"output,omitempty" yaml:"output,omitempty"`
}
type ErrorOutputConfig struct {
	ErrorMsg string    `json:"error_msg" yaml:"error_msg"`
	Batching *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type RedisHashOutputConfig struct {
	Url            string          `json:"url" yaml:"url"`
	Key            string          `json:"key" yaml:"key"`
	WalkMetadata   bool            `json:"walk_metadata" yaml:"walk_metadata"`
	WalkJsonObject bool            `json:"walk_json_object" yaml:"walk_json_object"`
	FieldsMapping  string          `json:"fields_mapping" yaml:"fields_mapping"`
	MaxInFlight    *int            `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
	Kind           *string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	Master         *string         `json:"master,omitempty" yaml:"master,omitempty"`
	Tls            *RedisTlsConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
}

type RedisHashConfig struct {
	Url            string         `json:"url" yaml:"url"`
	Key            string         `json:"key" yaml:"key"`
	WalkMetadata   bool           `json:"walk_metadata" yaml:"walk_metadata"`
	WalkJsonObject bool           `json:"walk_json_object" yaml:"walk_json_object"`
	Fields         map[string]any `json:"fields" yaml:"fields"`
	MaxInFlight    *int           `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
}

type RedisHashFields struct {
	Value string `json:"value" yaml:"value"`
}

type DropConfig struct{}

type DropOnConfig struct {
	Error        bool    `json:"error" yaml:"error"`
	Backpressure string  `json:"back_pressure" yaml:"back_pressure"`
	Output       Outputs `json:"output" yaml:"output"`
}

type RetryConfig struct {
	Output            OutputConfig `json:"output" yaml:"output"`
	InlineRetryConfig `json:",inline" yaml:",inline"`
}

type InlineRetryConfig struct {
	MaxRetries uint64  `json:"max_retries" yaml:"max_retries"`
	Backoff    Backoff `json:"backoff" yaml:"backoff"`
}

type Backoff struct {
	InitialInterval string `json:"initial_interval,omitempty" yaml:"initial_interval,omitempty"`
	MaxInterval     string `json:"max_interval,omitempty" yaml:"max_interval,omitempty"`
	MaxElapsedTime  string `json:"max_elapsed_time,omitempty" yaml:"max_elapsed_time,omitempty"`
}

type SqlRaw struct {
	Driver          string    `json:"driver" yaml:"driver"`
	Dsn             string    `json:"dsn" yaml:"dsn"`
	Query           string    `json:"query" yaml:"query"`
	ArgsMapping     string    `json:"args_mapping" yaml:"args_mapping"`
	InitStatement   string    `json:"init_statement" yaml:"init_statement"`
	ConnMaxIdleTime string    `json:"conn_max_idle_time,omitempty" yaml:"conn_max_idle_time,omitempty"`
	ConnMaxLifeTime string    `json:"conn_max_life_time,omitempty" yaml:"conn_max_life_time,omitempty"`
	ConnMaxIdle     int       `json:"conn_max_idle,omitempty" yaml:"conn_max_idle,omitempty"`
	ConnMaxOpen     int       `json:"conn_max_open,omitempty" yaml:"conn_max_open,omitempty"`
	Batching        *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type PooledSqlUpdate struct {
	ConnectionId             string    `json:"connection_id" yaml:"connection_id"`
	Schema                   string    `json:"schema" yaml:"schema"`
	Table                    string    `json:"table" yaml:"table"`
	Columns                  []string  `json:"columns" yaml:"columns"`
	WhereColumns             []string  `json:"where_columns" yaml:"where_columns"`
	SkipForeignKeyViolations bool      `json:"skip_foreign_key_violations" yaml:"skip_foreign_key_violations"`
	ArgsMapping              string    `json:"args_mapping" yaml:"args_mapping"`
	Batching                 *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
	MaxRetryAttempts         *uint     `json:"max_retry_attempts,omitempty" yaml:"max_retry_attempts,omitempty"`
	RetryAttemptDelay        *string   `json:"retry_attempt_delay,omitempty" yaml:"retry_attempt_delay,omitempty"`
	MaxInFlight              int       `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
}

type ColumnDefaultProperties struct {
	NeedsReset            bool `json:"needs_reset" yaml:"needs_reset"`
	NeedsOverride         bool `json:"needs_override" yaml:"needs_override"`
	HasDefaultTransformer bool `json:"has_default_transformer" yaml:"has_default_transformer"`
}

type PooledSqlInsert struct {
	ConnectionId             string                              `json:"connection_id" yaml:"connection_id"`
	Schema                   string                              `json:"schema" yaml:"schema"`
	Table                    string                              `json:"table" yaml:"table"`
	Columns                  []string                            `json:"columns" yaml:"columns"`
	ColumnsDataTypes         []string                            `json:"column_data_types" yaml:"column_data_types"`
	ColumnDefaultProperties  map[string]*ColumnDefaultProperties `json:"column_default_properties" yaml:"column_default_properties"`
	OnConflictDoNothing      bool                                `json:"on_conflict_do_nothing" yaml:"on_conflict_do_nothing"`
	TruncateOnRetry          bool                                `json:"truncate_on_retry" yaml:"truncate_on_retry"`
	SkipForeignKeyViolations bool                                `json:"skip_foreign_key_violations" yaml:"skip_foreign_key_violations"`
	RawInsertMode            bool                                `json:"raw_insert_mode" yaml:"raw_insert_mode"`
	ArgsMapping              string                              `json:"args_mapping" yaml:"args_mapping"`
	Batching                 *Batching                           `json:"batching,omitempty" yaml:"batching,omitempty"`
	Prefix                   *string                             `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Suffix                   *string                             `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	MaxRetryAttempts         *uint                               `json:"max_retry_attempts,omitempty" yaml:"max_retry_attempts,omitempty"`
	RetryAttemptDelay        *string                             `json:"retry_attempt_delay,omitempty" yaml:"retry_attempt_delay,omitempty"`
	MaxInFlight              int                                 `json:"max_in_flight,omitempty" yaml:"max_in_flight,omitempty"`
}

type SqlInsert struct {
	Driver          string    `json:"driver" yaml:"driver"`
	Dsn             string    `json:"dsn" yaml:"dsn"`
	Table           string    `json:"table" yaml:"table"`
	Columns         []string  `json:"columns" yaml:"columns"`
	ArgsMapping     string    `json:"args_mapping" yaml:"args_mapping"`
	InitStatement   string    `json:"init_statement" yaml:"init_statement"`
	ConnMaxIdleTime string    `json:"conn_max_idle_time,omitempty" yaml:"conn_max_idle_time,omitempty"`
	ConnMaxLifeTime string    `json:"conn_max_life_time,omitempty" yaml:"conn_max_life_time,omitempty"`
	ConnMaxIdle     int       `json:"conn_max_idle,omitempty" yaml:"conn_max_idle,omitempty"`
	ConnMaxOpen     int       `json:"conn_max_open,omitempty" yaml:"conn_max_open,omitempty"`
	Batching        *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type AwsS3Insert struct {
	Bucket       string    `json:"bucket" yaml:"bucket"`
	MaxInFlight  int       `json:"max_in_flight" yaml:"max_in_flight"`
	Path         string    `json:"path" yaml:"path"`
	Batching     *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
	Timeout      string    `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	StorageClass string    `json:"storage_class,omitempty" yaml:"storage_class,omitempty"`
	ContentType  string    `json:"content_type,omitempty" yaml:"content_type,omitempty"`

	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`

	Credentials *AwsCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type AwsCredentials struct {
	Profile        string `json:"profile,omitempty" yaml:"profile,omitempty"`
	Id             string `json:"id,omitempty" yaml:"id,omitempty"`
	Secret         string `json:"secret,omitempty" yaml:"secret,omitempty"`
	Token          string `json:"token,omitempty" yaml:"token,omitempty"`
	FromEc2Role    bool   `json:"from_ec2_role,omitempty" yaml:"from_ec2_role,omitempty"`
	Role           string `json:"role,omitempty" yaml:"role,omitempty"`
	RoleExternalId string `json:"role_external_id,omitempty" yaml:"role_external_id,omitempty"`
}

type GcpCloudStorageOutput struct {
	Bucket      string    `json:"bucket" yaml:"bucket"`
	Path        string    `json:"path" yaml:"path"`
	MaxInFlight int       `json:"max_in_flight" yaml:"max_in_flight"`
	Batching    *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`

	ContentType     *string `json:"content_type,omitempty" yaml:"content_type,omitempty"`
	ContentEncoding *string `json:"content_encoding,omitempty" yaml:"content_encoding,omitempty"`
	CollisionMode   *string `json:"collision_mode,omitempty" yaml:"collision_mode,omitempty"`
	ChunkSize       *int    `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty"`
	Timeout         *string `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

type Batching struct {
	Count      int               `json:"count" yaml:"count"`
	ByteSize   int               `json:"byte_size" yaml:"byte_size"`
	Period     string            `json:"period" yaml:"period"`
	Check      string            `json:"check" yaml:"check"`
	Processors []*BatchProcessor `json:"processors" yaml:"processors"`
}

type BatchProcessor struct {
	Archive      *ArchiveProcessor   `json:"archive,omitempty" yaml:"archive,omitempty"`
	Compress     *CompressProcessor  `json:"compress,omitempty" yaml:"compress,omitempty"`
	SqlToJson    *SqlToJsonConfig    `json:"sql_to_json,omitempty" yaml:"sql_to_json,omitempty"`
	JsonToSql    *JsonToSqlConfig    `json:"json_to_sql,omitempty" yaml:"json_to_sql,omitempty"`
	NeosyncToPgx *NeosyncToPgxConfig `json:"neosync_to_pgx,omitempty" yaml:"neosync_to_pgx,omitempty"`
}

type NeosyncToPgxConfig struct {
}

type JsonToSqlConfig struct {
	ColumnDataTypes map[string]string `json:"column_data_types" yaml:"column_data_types"`
}

type SqlToJsonConfig struct{}

type ArchiveProcessor struct {
	Format string  `json:"format" yaml:"format"`
	Path   *string `json:"path,omitempty" yaml:"path,omitempty"`
}

type CompressProcessor struct {
	Algorithm string `json:"algorithm" yaml:"algorithm"`
}

type OutputBrokerConfig struct {
	Pattern string    `json:"pattern" yaml:"pattern"`
	Outputs []Outputs `json:"outputs" yaml:"outputs"`
}
