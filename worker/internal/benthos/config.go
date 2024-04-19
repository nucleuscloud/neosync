package neosync_benthos

type BenthosConfig struct {
	// HTTP         HTTPConfig `json:"http" yaml:"http"`
	StreamConfig `json:",inline" yaml:",inline"`
}

// type HTTPConfig struct {
// 	Address string `json:"address" yaml:"address"`
// 	Enabled bool   `json:"enabled" yaml:"enabled"`
// 	// RootPath       string                     `json:"root_path" yaml:"root_path"`
// 	// DebugEndpoints bool                       `json:"debug_endpoints" yaml:"debug_endpoints"`
// 	// CertFile       string                     `json:"cert_file" yaml:"cert_file"`
// 	// KeyFile        string                     `json:"key_file" yaml:"key_file"`
// 	// CORS           httpserver.CORSConfig      `json:"cors" yaml:"cors"`
// 	// BasicAuth      httpserver.BasicAuthConfig `json:"basic_auth" yaml:"basic_auth"`
// }

type StreamConfig struct {
	Input          *InputConfig           `json:"input" yaml:"input"`
	Buffer         *BufferConfig          `json:"buffer,omitempty" yaml:"buffer,omitempty"`
	Pipeline       *PipelineConfig        `json:"pipeline" yaml:"pipeline"`
	Output         *OutputConfig          `json:"output" yaml:"output"`
	CacheResources []*CacheResourceConfig `json:"cache_resources,omitempty" yaml:"cache_resources,omitempty"`
	Metrics        *Metrics               `json:"metrics,omitempty" yaml:"metrics,omitempty"`
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
	SqlSelect         *SqlSelect         `json:"sql_select,omitempty" yaml:"sql_select,omitempty"`
	PooledSqlRaw      *InputPooledSqlRaw `json:"pooled_sql_raw,omitempty" yaml:"pooled_sql_raw,omitempty"`
	Generate          *Generate          `json:"generate,omitempty" yaml:"generate,omitempty"`
	GenerateSqlSelect *GenerateSqlSelect `json:"generate_sql_select,omitempty" yaml:"generate_sql_select,omitempty"`
}

type GenerateSqlSelect struct {
	Mapping         string              `json:"mapping" yaml:"mapping"`
	Count           int                 `json:"count" yaml:"count"`
	Driver          string              `json:"driver" yaml:"driver"`
	Dsn             string              `json:"dsn" yaml:"dsn"`
	TableColumnsMap map[string][]string `json:"table_columns_map" yaml:"table_columns_map"`
	ColumnNameMap   map[string]string   `json:"column_name_map,omitempty" yaml:"column_name_map,omitempty"`
}

type Generate struct {
	Mapping string `json:"mapping" yaml:"mapping"`

	Interval  string `json:"interval" yaml:"interval"`
	Count     int    `json:"count" yaml:"count"`
	BatchSize *int   `json:"batch_size,omitempty" yaml:"batch_size,omitempty"`
}

type InputPooledSqlRaw struct {
	Driver      string `json:"driver" yaml:"driver"`
	Dsn         string `json:"dsn" yaml:"dsn"`
	Query       string `json:"query" yaml:"query"`
	ArgsMapping string `json:"args_mapping,omitempty" yaml:"args_mapping,omitempty"`
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
	Mutation   *string               `json:"mutation,omitempty" yaml:"mutation,omitempty"`
	Javascript *JavascriptConfig     `json:"javascript,omitempty" yaml:"javascript,omitempty"`
	Branch     *BranchConfig         `json:"branch,omitempty" yaml:"branch,omitempty"`
	Cache      *CacheConfig          `json:"cache,omitempty" yaml:"cache,omitempty"`
	Mapping    *string               `json:"mapping,omitempty" yaml:"mapping,omitempty"`
	Redis      *RedisProcessorConfig `json:"redis,omitempty" yaml:"redis,omitempty"`
	Error      *ErrorProcessorConfig `json:"error,omitempty" yaml:"error,omitempty"`
	Catch      []*ProcessorConfig    `json:"catch,omitempty" yaml:"catch,omitempty"`
	While      *WhileProcessorConfig `json:"while,omitempty" yaml:"while,omitempty"`
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
	// Broker  *OutputBrokerConfig `json:"broker,omitempty" yaml:"broker,omitempty"`
}

type Outputs struct {
	SqlInsert       *SqlInsert             `json:"sql_insert,omitempty" yaml:"sql_insert,omitempty"`
	SqlRaw          *SqlRaw                `json:"sql_raw,omitempty" yaml:"sql_raw,omitempty"`
	PooledSqlRaw    *PooledSqlRaw          `json:"pooled_sql_raw,omitempty" yaml:"pooled_sql_raw,omitempty"`
	PooledSqlInsert *PooledSqlInsert       `json:"pooled_sql_insert,omitempty" yaml:"pooled_sql_insert,omitempty"`
	PooledSqlUpdate *PooledSqlUpdate       `json:"pooled_sql_update,omitempty" yaml:"pooled_sql_update,omitempty"`
	AwsS3           *AwsS3Insert           `json:"aws_s3,omitempty" yaml:"aws_s3,omitempty"`
	Retry           *RetryConfig           `json:"retry,omitempty" yaml:"retry,omitempty"`
	Broker          *OutputBrokerConfig    `json:"broker,omitempty" yaml:"broker,omitempty"`
	DropOn          *DropOnConfig          `json:"drop_on,omitempty" yaml:"drop_on,omitempty"`
	Drop            *DropConfig            `json:"drop,omitempty" yaml:"drop,omitempty"`
	Resource        string                 `json:"resource,omitempty" yaml:"resource,omitempty"`
	Fallback        []Outputs              `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	RedisHashOutput *RedisHashOutputConfig `json:"redis_hash_output,omitempty" yaml:"redis_hash_output,omitempty"`
	Error           *ErrorOutputConfig     `json:"error,omitempty" yaml:"error,omitempty"`
	Switch          *SwitchOutputConfig    `json:"switch,omitempty" yaml:"switch,omitempty"`
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
	ErrorMsg   string    `json:"error_msg" yaml:"error_msg"`
	MaxRetries *int      `json:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	Batching   *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
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

type PooledSqlRaw struct {
	Driver        string `json:"driver" yaml:"driver"`
	Dsn           string `json:"dsn" yaml:"dsn"`
	Query         string `json:"query" yaml:"query"`
	ArgsMapping   string `json:"args_mapping" yaml:"args_mapping"`
	InitStatement string `json:"init_statement" yaml:"init_statement"`
	// ConnMaxIdleTime string    `json:"conn_max_idle_time,omitempty" yaml:"conn_max_idle_time,omitempty"`
	// ConnMaxLifeTime string    `json:"conn_max_life_time,omitempty" yaml:"conn_max_life_time,omitempty"`
	// ConnMaxIdle     int       `json:"conn_max_idle,omitempty" yaml:"conn_max_idle,omitempty"`
	// ConnMaxOpen     int       `json:"conn_max_open,omitempty" yaml:"conn_max_open,omitempty"`
	Batching *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type PooledSqlUpdate struct {
	Driver       string    `json:"driver" yaml:"driver"`
	Dsn          string    `json:"dsn" yaml:"dsn"`
	Schema       string    `json:"schema" yaml:"schema"`
	Table        string    `json:"table" yaml:"table"`
	Columns      []string  `json:"columns" yaml:"columns"`
	WhereColumns []string  `json:"where_columns" yaml:"where_columns"`
	ArgsMapping  string    `json:"args_mapping" yaml:"args_mapping"`
	Batching     *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
}

type PooledSqlInsert struct {
	Driver              string    `json:"driver" yaml:"driver"`
	Dsn                 string    `json:"dsn" yaml:"dsn"`
	Schema              string    `json:"schema" yaml:"schema"`
	Table               string    `json:"table" yaml:"table"`
	Columns             []string  `json:"columns" yaml:"columns"`
	OnConflictDoNothing bool      `json:"on_conflict_do_nothing" yaml:"on_conflict_do_nothing"`
	TruncateOnRetry     bool      `json:"truncate_on_retry" yaml:"truncate_on_retry"`
	ArgsMapping         string    `json:"args_mapping" yaml:"args_mapping"`
	Batching            *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`
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
	Bucket      string    `json:"bucket" yaml:"bucket"`
	MaxInFlight int       `json:"max_in_flight" yaml:"max_in_flight"`
	Path        string    `json:"path" yaml:"path"`
	Batching    *Batching `json:"batching,omitempty" yaml:"batching,omitempty"`

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

type Batching struct {
	Count      int               `json:"count" yaml:"count"`
	ByteSize   int               `json:"byte_size" yaml:"byte_size"`
	Period     string            `json:"period" yaml:"period"`
	Check      string            `json:"check" yaml:"check"`
	Processors []*BatchProcessor `json:"processors" yaml:"processors"`
}

type BatchProcessor struct {
	Archive  *ArchiveProcessor  `json:"archive,omitempty" yaml:"archive,omitempty"`
	Compress *CompressProcessor `json:"compress,omitempty" yaml:"compress,omitempty"`
}

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
