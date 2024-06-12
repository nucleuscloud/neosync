package cli_neosync_benthos

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
	Logger   *LoggerConfig   `json:"logger" yaml:"logger,omitempty"`
	Input    *InputConfig    `json:"input" yaml:"input"`
	Buffer   *BufferConfig   `json:"buffer,omitempty" yaml:"buffer,omitempty"`
	Pipeline *PipelineConfig `json:"pipeline" yaml:"pipeline"`
	Output   *OutputConfig   `json:"output" yaml:"output"`
}

type LoggerConfig struct {
	Level        string `json:"level" yaml:"level"`
	AddTimestamp bool   `json:"add_timestamp" yaml:"add_timestamp"`
}
type InputConfig struct {
	Label  string `json:"label" yaml:"label"`
	Inputs `json:",inline" yaml:",inline"`
}

type Inputs struct {
	NeosyncConnectionData *NeosyncConnectionData `json:"neosync_connection_data,omitempty" yaml:"neosync_connection_data,omitempty"`
}

type NeosyncConnectionData struct {
	ApiKey         *string `json:"api_key,omitempty" yaml:"api_key,omitempty"`
	ApiUrl         string  `json:"api_url" yaml:"api_url"`
	ConnectionId   string  `json:"connection_id" yaml:"connection_id"`
	ConnectionType string  `json:"connection_type" yaml:"connection_type"`
	JobId          *string `json:"job_id,omitempty" yaml:"job_id,omitempty"`
	JobRunId       *string `json:"job_run_id,omitempty" yaml:"job_run_id,omitempty"`
	Schema         string  `json:"schema" yaml:"schema"`
	Table          string  `json:"table" yaml:"table"`
}

type BufferConfig struct{}

type PipelineConfig struct {
	Threads    int               `json:"threads" yaml:"threads"`
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
}

type ProcessorConfig struct {
}

type BranchConfig struct {
	Processors []ProcessorConfig `json:"processors" yaml:"processors"`
	RequestMap *string           `json:"request_map,omitempty" yaml:"request_map,omitempty"`
	ResultMap  *string           `json:"result_map,omitempty" yaml:"result_map,omitempty"`
}

type OutputConfig struct {
	Label      string `json:"label" yaml:"label"`
	Outputs    `json:",inline" yaml:",inline"`
	Processors []ProcessorConfig `json:"processors,omitempty" yaml:"processors,omitempty"`
	// Broker  *OutputBrokerConfig `json:"broker,omitempty" yaml:"broker,omitempty"`
}

type Outputs struct {
	PooledSqlInsert *PooledSqlInsert `json:"pooled_sql_insert,omitempty" yaml:"pooled_sql_insert,omitempty"`
	PooledSqlUpdate *PooledSqlUpdate `json:"pooled_sql_update,omitempty" yaml:"pooled_sql_update,omitempty"`
	AwsS3           *AwsS3Insert     `json:"aws_s3,omitempty" yaml:"aws_s3,omitempty"`
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
