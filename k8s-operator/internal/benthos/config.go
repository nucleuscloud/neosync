package neosync_benthos

type BenthosConfig struct {
	HTTP         HTTPConfig `json:"http" yaml:"http"`
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
	Inputs `json:"inline" yaml:",inline"`
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
	Broker  *OutputBrokerConfig `json:"broker,omitempty" yaml:"broker,omitempty"`
}

type Outputs struct {
	SqlInsert *SqlInsert `json:"sql_insert,omitempty" yaml:"sql_insert,omitempty"`
}

type SqlInsert struct {
	Driver        string   `json:"driver" yaml:"driver"`
	Dsn           string   `json:"dsn" yaml:"dsn"`
	Table         string   `json:"table" yaml:"table"`
	Columns       []string `json:"columns" yaml:"columns"`
	ArgsMapping   string   `json:"args_mapping" yaml:"args_mapping"`
	InitStatement string   `json:"init_statement" yaml:"init_statement"`
}

type OutputBrokerConfig struct {
	Pattern string    `json:"pattern" yaml:"pattern"`
	Outputs []Outputs `json:"outputs" yaml:"outputs"`
}
