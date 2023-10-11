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
	Input    *InputConfig    `json:"input" yaml:"input"`
	Buffer   *BufferConfig   `json:"buffer" yaml:"buffer"`
	Pipeline *PipelineConfig `json:"pipeline" yaml:"pipeline"`
	Output   *OutputConfig   `json:"output" yaml:"output"`
}

type InputConfig struct {
	Label  string `json:"label" yaml:"label"`
	Inputs `json:",inline" yaml:",inline"`
}

type Inputs struct {
	SqlSelect *SqlSelect `json:"sql_select,omitempty" yaml:"sql_select,omitempty"`
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
	Mutation string `json:"mutation" yaml:"mutation"`
}

type OutputConfig struct {
	Label   string `json:"label" yaml:"label"`
	Outputs `json:",inline" yaml:",inline"`
	// Broker  *OutputBrokerConfig `json:"broker,omitempty" yaml:"broker,omitempty"`
}

type Outputs struct {
	SqlInsert *SqlInsert          `json:"sql_insert,omitempty" yaml:"sql_insert,omitempty"`
	AwsS3     *AwsS3Insert        `json:"aws_s3,omitempty" yaml:"aws_s3,omitempty"`
	Retry     *RetryConfig        `json:"retry,omitempty" yaml:"retry,omitempty"`
	Broker    *OutputBrokerConfig `json:"broker,omitempty" yaml:"broker,omitempty"`
	DropOn    *DropOnConfig       `json:"drop_on,omitempty" yaml:"drop_on,omitempty"`
	Drop      *DropConfig         `json:"drop,omitempty" yaml:"drop,omitempty"`
	Resource  string              `json:"resource,omitempty" yaml:"resource,omitempty"`
	Fallback  []Outputs           `json:"fallback,omitempty" yaml:"fallback,omitempty"`
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
	InitialInterval string `json:"initial_interval" yaml:"initial_interval"`
	MaxInterval     string `json:"max_interval" yaml:"max_interval"`
	MaxElapsedTime  string `json:"max_elapsed_time" yaml:"max_elapsed_time"`
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
