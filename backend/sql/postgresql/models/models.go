package pg_models

import (
	"fmt"
	"strings"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type ConnectionConfig struct {
	PgConfig             *PostgresConnectionConfig       `json:"pgConfig,omitempty"`
	AwsS3Config          *AwsS3ConnectionConfig          `json:"awsS3Config,omitempty"`
	MysqlConfig          *MysqlConnectionConfig          `json:"mysqlConfig,omitempty"`
	LocalDirectoryConfig *LocalDirectoryConnectionConfig `json:"localDirConfig,omitempty"`
	OpenAiConfig         *OpenAiConnectionConfig         `json:"openaiConfig,omitempty"`
}

func (c *ConnectionConfig) ToDto() *mgmtv1alpha1.ConnectionConfig {
	if c.PgConfig != nil {
		var tunnel *mgmtv1alpha1.SSHTunnel
		if c.PgConfig.SSHTunnel != nil {
			tunnel = c.PgConfig.SSHTunnel.ToDto()
		}
		var connectionOptions *mgmtv1alpha1.SqlConnectionOptions
		if c.PgConfig.ConnectionOptions != nil {
			connectionOptions = c.PgConfig.ConnectionOptions.ToDto()
		}
		var clientTls *mgmtv1alpha1.ClientTlsConfig
		if c.PgConfig.ClientTls != nil {
			clientTls = c.PgConfig.ClientTls.ToDto()
		}
		if c.PgConfig.Connection != nil {
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.PostgresConnection{
								Host:    c.PgConfig.Connection.Host,
								Port:    c.PgConfig.Connection.Port,
								Name:    c.PgConfig.Connection.Name,
								User:    c.PgConfig.Connection.User,
								Pass:    c.PgConfig.Connection.Pass,
								SslMode: c.PgConfig.Connection.SslMode,
							},
						},
						Tunnel:            tunnel,
						ConnectionOptions: connectionOptions,
						ClientTls:         clientTls,
					},
				},
			}
		} else if c.PgConfig.Url != nil {
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_PgConfig{
					PgConfig: &mgmtv1alpha1.PostgresConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.PostgresConnectionConfig_Url{
							Url: *c.PgConfig.Url,
						},
						Tunnel:            tunnel,
						ConnectionOptions: connectionOptions,
						ClientTls:         clientTls,
					},
				},
			}
		}
	} else if c.MysqlConfig != nil {
		var tunnel *mgmtv1alpha1.SSHTunnel
		if c.MysqlConfig.SSHTunnel != nil {
			tunnel = c.MysqlConfig.SSHTunnel.ToDto()
		}
		var connectionOptions *mgmtv1alpha1.SqlConnectionOptions
		if c.MysqlConfig.ConnectionOptions != nil {
			connectionOptions = c.MysqlConfig.ConnectionOptions.ToDto()
		}
		if c.MysqlConfig.Connection != nil {
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Connection{
							Connection: &mgmtv1alpha1.MysqlConnection{
								User:     c.MysqlConfig.Connection.User,
								Pass:     c.MysqlConfig.Connection.Pass,
								Protocol: c.MysqlConfig.Connection.Protocol,
								Host:     c.MysqlConfig.Connection.Host,
								Port:     c.MysqlConfig.Connection.Port,
								Name:     c.MysqlConfig.Connection.Name,
							},
						},
						Tunnel:            tunnel,
						ConnectionOptions: connectionOptions,
					},
				},
			}
		} else if c.MysqlConfig.Url != nil {
			return &mgmtv1alpha1.ConnectionConfig{
				Config: &mgmtv1alpha1.ConnectionConfig_MysqlConfig{
					MysqlConfig: &mgmtv1alpha1.MysqlConnectionConfig{
						ConnectionConfig: &mgmtv1alpha1.MysqlConnectionConfig_Url{
							Url: *c.MysqlConfig.Url,
						},
						Tunnel: tunnel,
					},
				},
			}
		}
	} else if c.AwsS3Config != nil {
		return &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
				AwsS3Config: c.AwsS3Config.ToDto(),
			},
		}
	} else if c.LocalDirectoryConfig != nil {
		return &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_LocalDirConfig{
				LocalDirConfig: c.LocalDirectoryConfig.ToDto(),
			},
		}
	} else if c.OpenAiConfig != nil {
		return &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_OpenaiConfig{
				OpenaiConfig: c.OpenAiConfig.ToDto(),
			},
		}
	}
	return nil
}

func (c *ConnectionConfig) FromDto(dto *mgmtv1alpha1.ConnectionConfig) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		c.PgConfig = &PostgresConnectionConfig{}
		if config.PgConfig.Tunnel != nil {
			c.PgConfig.SSHTunnel = &SSHTunnel{}
			c.PgConfig.SSHTunnel.FromDto(config.PgConfig.Tunnel)
		}
		if config.PgConfig.ConnectionOptions != nil {
			c.PgConfig.ConnectionOptions = &ConnectionOptions{}
			c.PgConfig.ConnectionOptions.FromDto(config.PgConfig.ConnectionOptions)
		}
		if config.PgConfig.GetClientTls() != nil {
			c.PgConfig.ClientTls = &ClientTls{
				RootCert:   config.PgConfig.GetClientTls().RootCert,
				ClientCert: config.PgConfig.GetClientTls().ClientCert,
				ClientKey:  config.PgConfig.GetClientTls().ClientKey,
			}
		}
		switch pgcfg := config.PgConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.PostgresConnectionConfig_Connection:
			c.PgConfig.Connection = &PostgresConnection{
				Host:    pgcfg.Connection.Host,
				Port:    pgcfg.Connection.Port,
				Name:    pgcfg.Connection.Name,
				User:    pgcfg.Connection.User,
				Pass:    pgcfg.Connection.Pass,
				SslMode: pgcfg.Connection.SslMode,
			}
		case *mgmtv1alpha1.PostgresConnectionConfig_Url:
			c.PgConfig.Url = &pgcfg.Url
		default:
			return fmt.Errorf("invalid postgres format")
		}
	case *mgmtv1alpha1.ConnectionConfig_MysqlConfig:
		c.MysqlConfig = &MysqlConnectionConfig{}
		if config.MysqlConfig.Tunnel != nil {
			c.MysqlConfig.SSHTunnel = &SSHTunnel{}
			c.MysqlConfig.SSHTunnel.FromDto(config.MysqlConfig.Tunnel)
		}
		if config.MysqlConfig.ConnectionOptions != nil {
			c.MysqlConfig.ConnectionOptions = &ConnectionOptions{}
			c.MysqlConfig.ConnectionOptions.FromDto(config.MysqlConfig.ConnectionOptions)
		}
		switch mysqlcfg := config.MysqlConfig.ConnectionConfig.(type) {
		case *mgmtv1alpha1.MysqlConnectionConfig_Connection:
			c.MysqlConfig.Connection = &MysqlConnection{
				User:     mysqlcfg.Connection.User,
				Pass:     mysqlcfg.Connection.Pass,
				Protocol: mysqlcfg.Connection.Protocol,
				Host:     mysqlcfg.Connection.Host,
				Port:     mysqlcfg.Connection.Port,
				Name:     mysqlcfg.Connection.Name,
			}
		case *mgmtv1alpha1.MysqlConnectionConfig_Url:
			c.MysqlConfig.Url = &mysqlcfg.Url
		default:
			return fmt.Errorf("invalid mysql format")
		}
	case *mgmtv1alpha1.ConnectionConfig_AwsS3Config:
		c.AwsS3Config = &AwsS3ConnectionConfig{}
		err := c.AwsS3Config.FromDto(config.AwsS3Config)
		if err != nil {
			return err
		}
	case *mgmtv1alpha1.ConnectionConfig_LocalDirConfig:
		c.LocalDirectoryConfig = &LocalDirectoryConnectionConfig{}
		c.LocalDirectoryConfig.FromDto(config.LocalDirConfig)
	case *mgmtv1alpha1.ConnectionConfig_OpenaiConfig:
		c.OpenAiConfig = &OpenAiConnectionConfig{}
		c.OpenAiConfig.FromDto(config.OpenaiConfig)
	default:
		return fmt.Errorf("unable to convert to ConnectionConfig from DTO ConnectionConfig, type not supported: %T", config)
	}
	return nil
}

type PostgresConnectionConfig struct {
	Connection        *PostgresConnection `json:"connection,omitempty"`
	Url               *string             `json:"url,omitempty"`
	SSHTunnel         *SSHTunnel          `json:"sshTunnel,omitempty"`
	ConnectionOptions *ConnectionOptions  `json:"connectionOptions,omitempty"`
	ClientTls         *ClientTls          `json:"clientTls,omitempty"`
}

type PostgresConnection struct {
	Host    string  `json:"host"`
	Port    int32   `json:"port"`
	Name    string  `json:"name"`
	User    string  `json:"user"`
	Pass    string  `json:"pass"`
	SslMode *string `json:"sslMode,omitempty"`
}

type ConnectionOptions struct {
	MaxConnectionLimit *int32 `json:"maxConnectionLimit,omitempty"`
}

func (s *ConnectionOptions) ToDto() *mgmtv1alpha1.SqlConnectionOptions {
	return &mgmtv1alpha1.SqlConnectionOptions{
		MaxConnectionLimit: s.MaxConnectionLimit,
	}
}

func (s *ConnectionOptions) FromDto(dto *mgmtv1alpha1.SqlConnectionOptions) {
	s.MaxConnectionLimit = dto.MaxConnectionLimit
}

type SSHTunnel struct {
	Host string `json:"host"`
	Port int32  `json:"port"`
	User string `json:"user"`

	KnownHostPublicKey *string            `json:"knownHostPublicKey,omitempty"`
	SSHAuthentication  *SSHAuthentication `json:"sshAuthentication,omitempty"`
}

func (s *SSHTunnel) ToDto() *mgmtv1alpha1.SSHTunnel {
	var auth *mgmtv1alpha1.SSHAuthentication
	if s.SSHAuthentication != nil {
		auth = s.SSHAuthentication.ToDto()
	}
	return &mgmtv1alpha1.SSHTunnel{
		Host:               s.Host,
		Port:               s.Port,
		User:               s.User,
		KnownHostPublicKey: s.KnownHostPublicKey,
		Authentication:     auth,
	}
}

func (s *SSHTunnel) FromDto(dto *mgmtv1alpha1.SSHTunnel) {
	s.Host = dto.Host
	s.Port = dto.Port
	s.User = dto.User
	s.KnownHostPublicKey = dto.KnownHostPublicKey

	if dto.Authentication != nil {
		auth := &SSHAuthentication{}
		auth.FromDto(dto.Authentication)
		s.SSHAuthentication = auth
	}
}

type SSHAuthentication struct {
	SSHPassphrase *SSHPassphrase `json:"sshPassphrase,omitempty"`
	SSHPrivateKey *SSHPrivateKey `json:"sshPrivateKey,omitempty"`
}

func (s *SSHAuthentication) ToDto() *mgmtv1alpha1.SSHAuthentication {
	if s.SSHPassphrase != nil {
		return &mgmtv1alpha1.SSHAuthentication{
			AuthConfig: &mgmtv1alpha1.SSHAuthentication_Passphrase{
				Passphrase: &mgmtv1alpha1.SSHPassphrase{Value: s.SSHPassphrase.Value},
			},
		}
	} else if s.SSHPrivateKey != nil {
		return &mgmtv1alpha1.SSHAuthentication{
			AuthConfig: &mgmtv1alpha1.SSHAuthentication_PrivateKey{
				PrivateKey: &mgmtv1alpha1.SSHPrivateKey{
					Value:      s.SSHPrivateKey.Value,
					Passphrase: s.SSHPrivateKey.Passphrase,
				},
			},
		}
	}
	return nil
}

func (s *SSHAuthentication) FromDto(dto *mgmtv1alpha1.SSHAuthentication) {
	switch config := dto.AuthConfig.(type) {
	case *mgmtv1alpha1.SSHAuthentication_Passphrase:
		s.SSHPassphrase = &SSHPassphrase{
			Value: config.Passphrase.Value,
		}
	case *mgmtv1alpha1.SSHAuthentication_PrivateKey:
		s.SSHPrivateKey = &SSHPrivateKey{
			Value:      config.PrivateKey.Value,
			Passphrase: config.PrivateKey.Passphrase,
		}
	}
}

type SSHPassphrase struct {
	Value string `json:"value"`
}

type SSHPrivateKey struct {
	Value      string  `json:"value"`
	Passphrase *string `json:"passphrase,omitempty"`
}

type ClientTls struct {
	RootCert   *string `json:"rootCert,omitempty"`
	ClientCert *string `json:"clientCert,omitempty"`
	ClientKey  *string `json:"clientKey,omitempty"`
}

func (c *ClientTls) ToDto() *mgmtv1alpha1.ClientTlsConfig {
	return &mgmtv1alpha1.ClientTlsConfig{
		RootCert:   c.RootCert,
		ClientCert: c.ClientCert,
		ClientKey:  c.ClientKey,
	}
}

func (c *ClientTls) FromDto(dto *mgmtv1alpha1.ClientTlsConfig) {
	if dto == nil {
		dto = &mgmtv1alpha1.ClientTlsConfig{}
	}
	c.RootCert = dto.RootCert
	c.ClientCert = dto.ClientCert
	c.ClientKey = dto.ClientKey
}

type MysqlConnectionConfig struct {
	Connection        *MysqlConnection   `json:"connection,omitempty"`
	Url               *string            `json:"url,omitempty"`
	SSHTunnel         *SSHTunnel         `json:"sshTunnel,omitempty"`
	ConnectionOptions *ConnectionOptions `json:"connectionOptions,omitempty"`
}

type MysqlConnection struct {
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Name     string `json:"name"`
}

type AwsS3Credentials struct {
	Profile         *string `json:"profile,omitempty"`
	AccessKeyId     *string `json:"accessKeyId,omitempty"`
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
	SessionToken    *string `json:"sessionToken,omitempty"`
	FromEc2Role     *bool   `json:"fromEc2Role,omitempty"`
	RoleArn         *string `json:"roleArn,omitempty"`
	RoleExternalId  *string `json:"roleExternalId,omitempty"`
}

type LocalDirectoryConnectionConfig struct {
	Path string `json:"path"`
}

func (l *LocalDirectoryConnectionConfig) ToDto() *mgmtv1alpha1.LocalDirectoryConnectionConfig {
	return &mgmtv1alpha1.LocalDirectoryConnectionConfig{
		Path: l.Path,
	}
}
func (l *LocalDirectoryConnectionConfig) FromDto(dto *mgmtv1alpha1.LocalDirectoryConnectionConfig) {
	l.Path = dto.Path
}

type OpenAiConnectionConfig struct {
	ApiUrl string `json:"apiUrl"`
	ApiKey string `json:"apiKey"`
}

func (o *OpenAiConnectionConfig) ToDto() *mgmtv1alpha1.OpenAiConnectionConfig {
	return &mgmtv1alpha1.OpenAiConnectionConfig{
		ApiKey: o.ApiKey,
		ApiUrl: o.ApiUrl,
	}
}

func (o *OpenAiConnectionConfig) FromDto(dto *mgmtv1alpha1.OpenAiConnectionConfig) {
	if dto == nil {
		return
	}
	o.ApiKey = dto.ApiKey
	o.ApiUrl = dto.ApiUrl
}

func (a *AwsS3Credentials) ToDto() *mgmtv1alpha1.AwsS3Credentials {
	return &mgmtv1alpha1.AwsS3Credentials{
		Profile:         a.Profile,
		AccessKeyId:     a.AccessKeyId,
		SecretAccessKey: a.SecretAccessKey,
		SessionToken:    a.SessionToken,
		FromEc2Role:     a.FromEc2Role,
		RoleArn:         a.RoleArn,
		RoleExternalId:  a.RoleExternalId,
	}
}
func (a *AwsS3Credentials) FromDto(dto *mgmtv1alpha1.AwsS3Credentials) {
	if dto == nil {
		return
	}
	a.Profile = dto.Profile
	a.AccessKeyId = dto.AccessKeyId
	a.SecretAccessKey = dto.SecretAccessKey
	a.SessionToken = dto.SessionToken
	a.FromEc2Role = dto.FromEc2Role
	a.RoleArn = dto.RoleArn
	a.RoleExternalId = dto.RoleExternalId
}

type AwsS3ConnectionConfig struct {
	BucketArn   string
	Bucket      string
	PathPrefix  *string
	Credentials *AwsS3Credentials
	Region      *string
	Endpoint    *string
}

func (a *AwsS3ConnectionConfig) ToDto() *mgmtv1alpha1.AwsS3ConnectionConfig {
	var bucket = a.Bucket
	if a.Bucket == "" && a.BucketArn != "" {
		bucket = strings.ReplaceAll(a.BucketArn, "arn:aws:s3:::", "")
	}

	return &mgmtv1alpha1.AwsS3ConnectionConfig{
		Bucket:      bucket,
		PathPrefix:  a.PathPrefix,
		Credentials: a.Credentials.ToDto(),
		Region:      a.Region,
		Endpoint:    a.Endpoint,
	}
}
func (a *AwsS3ConnectionConfig) FromDto(dto *mgmtv1alpha1.AwsS3ConnectionConfig) error {
	a.Bucket = dto.Bucket
	a.PathPrefix = dto.PathPrefix
	a.Credentials = &AwsS3Credentials{}
	a.Credentials.FromDto(dto.Credentials)
	a.Region = dto.Region
	a.Endpoint = dto.Endpoint
	return nil
}

type JobMapping struct {
	Schema                string                      `json:"schema"`
	Table                 string                      `json:"table"`
	Column                string                      `json:"column"`
	JobMappingTransformer *JobMappingTransformerModel `json:"jobMappingTransformerModel,omitempty"`
}

func (jm *JobMapping) ToDto() *mgmtv1alpha1.JobMapping {
	return &mgmtv1alpha1.JobMapping{
		Schema:      jm.Schema,
		Table:       jm.Table,
		Column:      jm.Column,
		Transformer: jm.JobMappingTransformer.ToTransformerDto(),
	}
}

func (jm *JobMapping) FromDto(dto *mgmtv1alpha1.JobMapping) error {
	t := &JobMappingTransformerModel{}
	if err := t.FromTransformerDto(dto.Transformer); err != nil {
		return err
	}
	jm.Schema = dto.Schema
	jm.Table = dto.Table
	jm.Column = dto.Column
	jm.JobMappingTransformer = t
	return nil
}

type JobSourceOptions struct {
	PostgresOptions   *PostgresSourceOptions   `json:"postgresOptions,omitempty"`
	MysqlOptions      *MysqlSourceOptions      `json:"mysqlOptions,omitempty"`
	GenerateOptions   *GenerateSourceOptions   `json:"generateOptions,omitempty"`
	AiGenerateOptions *AiGenerateSourceOptions `json:"aiGenerateOptions,omitempty"`
}

type MysqlSourceOptions struct {
	HaltOnNewColumnAddition       bool                       `json:"haltOnNewColumnAddition"`
	SubsetByForeignKeyConstraints bool                       `json:"subsetByForeignKeyConstraints"`
	Schemas                       []*MysqlSourceSchemaOption `json:"schemas"`
	ConnectionId                  string                     `json:"connectionId"`
}
type PostgresSourceOptions struct {
	HaltOnNewColumnAddition       bool                          `json:"haltOnNewColumnAddition"`
	SubsetByForeignKeyConstraints bool                          `json:"subsetByForeignKeyConstraints"`
	Schemas                       []*PostgresSourceSchemaOption `json:"schemas"`
	ConnectionId                  string                        `json:"connectionId"`
}

type GenerateSourceOptions struct {
	Schemas              []*GenerateSourceSchemaOption `json:"schemas"`
	FkSourceConnectionId *string                       `json:"fkSourceConnectionId,omitempty"`
}

type GenerateSourceSchemaOption struct {
	Schema string                       `json:"schema"`
	Tables []*GenerateSourceTableOption `json:"tables"`
}
type GenerateSourceTableOption struct {
	Table    string `json:"table"`
	RowCount int64  `json:"rowCount,omitempty"`
}

type AiGenerateSourceOptions struct {
	AiConnectionId       string                          `json:"aiConnectionId"`
	Schemas              []*AiGenerateSourceSchemaOption `json:"schemas"`
	FkSourceConnectionId *string                         `json:"fkSourceConnectionId,omitempty"`
	ModelName            string                          `json:"modelName"`
	UserPrompt           *string                         `json:"userPrompt,omitempty"`
}

type AiGenerateSourceSchemaOption struct {
	Schema string                         `json:"schema"`
	Tables []*AiGenerateSourceTableOption `json:"tables"`
}
type AiGenerateSourceTableOption struct {
	Table    string `json:"table"`
	RowCount int64  `json:"rowCount,omitempty"`
}

func (s *PostgresSourceOptions) ToDto() *mgmtv1alpha1.PostgresSourceConnectionOptions {
	dto := &mgmtv1alpha1.PostgresSourceConnectionOptions{
		HaltOnNewColumnAddition:       s.HaltOnNewColumnAddition,
		SubsetByForeignKeyConstraints: s.SubsetByForeignKeyConstraints,
		ConnectionId:                  s.ConnectionId,
	}
	dto.Schemas = make([]*mgmtv1alpha1.PostgresSourceSchemaOption, len(s.Schemas))
	for idx := range s.Schemas {
		schema := s.Schemas[idx]
		tables := make([]*mgmtv1alpha1.PostgresSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &mgmtv1alpha1.PostgresSourceTableOption{
				Table:       table.Table,
				WhereClause: table.WhereClause,
			}
		}
		dto.Schemas[idx] = &mgmtv1alpha1.PostgresSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return dto
}
func (s *PostgresSourceOptions) FromDto(dto *mgmtv1alpha1.PostgresSourceConnectionOptions) {
	s.HaltOnNewColumnAddition = dto.HaltOnNewColumnAddition
	s.SubsetByForeignKeyConstraints = dto.SubsetByForeignKeyConstraints
	s.Schemas = FromDtoPostgresSourceSchemaOptions(dto.Schemas)
	s.ConnectionId = dto.ConnectionId
}

func FromDtoPostgresSourceSchemaOptions(dtos []*mgmtv1alpha1.PostgresSourceSchemaOption) []*PostgresSourceSchemaOption {
	output := make([]*PostgresSourceSchemaOption, len(dtos))
	for idx := range dtos {
		schema := dtos[idx]
		tables := make([]*PostgresSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &PostgresSourceTableOption{
				Table:       table.Table,
				WhereClause: table.WhereClause,
			}
		}
		output[idx] = &PostgresSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return output
}

func (s *MysqlSourceOptions) ToDto() *mgmtv1alpha1.MysqlSourceConnectionOptions {
	dto := &mgmtv1alpha1.MysqlSourceConnectionOptions{
		HaltOnNewColumnAddition:       s.HaltOnNewColumnAddition,
		SubsetByForeignKeyConstraints: s.SubsetByForeignKeyConstraints,
		ConnectionId:                  s.ConnectionId,
	}
	dto.Schemas = make([]*mgmtv1alpha1.MysqlSourceSchemaOption, len(s.Schemas))
	for idx := range s.Schemas {
		schema := s.Schemas[idx]
		tables := make([]*mgmtv1alpha1.MysqlSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &mgmtv1alpha1.MysqlSourceTableOption{
				Table:       table.Table,
				WhereClause: table.WhereClause,
			}
		}
		dto.Schemas[idx] = &mgmtv1alpha1.MysqlSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return dto
}
func (s *MysqlSourceOptions) FromDto(dto *mgmtv1alpha1.MysqlSourceConnectionOptions) {
	s.HaltOnNewColumnAddition = dto.HaltOnNewColumnAddition
	s.SubsetByForeignKeyConstraints = dto.SubsetByForeignKeyConstraints
	s.Schemas = FromDtoMysqlSourceSchemaOptions(dto.Schemas)
	s.ConnectionId = dto.ConnectionId
}

func FromDtoMysqlSourceSchemaOptions(dtos []*mgmtv1alpha1.MysqlSourceSchemaOption) []*MysqlSourceSchemaOption {
	output := make([]*MysqlSourceSchemaOption, len(dtos))
	for idx := range dtos {
		schema := dtos[idx]
		tables := make([]*MysqlSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &MysqlSourceTableOption{
				Table:       table.Table,
				WhereClause: table.WhereClause,
			}
		}
		output[idx] = &MysqlSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return output
}

func (s *GenerateSourceOptions) ToDto() *mgmtv1alpha1.GenerateSourceOptions {
	dto := &mgmtv1alpha1.GenerateSourceOptions{
		FkSourceConnectionId: s.FkSourceConnectionId,
	}
	dto.Schemas = make([]*mgmtv1alpha1.GenerateSourceSchemaOption, len(s.Schemas))
	for idx := range s.Schemas {
		schema := s.Schemas[idx]
		tables := make([]*mgmtv1alpha1.GenerateSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &mgmtv1alpha1.GenerateSourceTableOption{
				Table:    table.Table,
				RowCount: table.RowCount,
			}
		}
		dto.Schemas[idx] = &mgmtv1alpha1.GenerateSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return dto
}
func (s *GenerateSourceOptions) FromDto(dto *mgmtv1alpha1.GenerateSourceOptions) {
	s.FkSourceConnectionId = dto.FkSourceConnectionId
	s.Schemas = FromDtoGenerateSourceSchemaOptions(dto.Schemas)
}

func FromDtoGenerateSourceSchemaOptions(dtos []*mgmtv1alpha1.GenerateSourceSchemaOption) []*GenerateSourceSchemaOption {
	output := make([]*GenerateSourceSchemaOption, len(dtos))
	for idx := range dtos {
		schema := dtos[idx]
		tables := make([]*GenerateSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &GenerateSourceTableOption{
				Table:    table.Table,
				RowCount: table.RowCount,
			}
		}
		output[idx] = &GenerateSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return output
}

func (s *AiGenerateSourceOptions) ToDto() *mgmtv1alpha1.AiGenerateSourceOptions {
	dto := &mgmtv1alpha1.AiGenerateSourceOptions{
		FkSourceConnectionId: s.FkSourceConnectionId,
		AiConnectionId:       s.AiConnectionId,
		ModelName:            s.ModelName,
		UserPrompt:           s.UserPrompt,
	}
	dto.Schemas = make([]*mgmtv1alpha1.AiGenerateSourceSchemaOption, len(s.Schemas))
	for idx := range s.Schemas {
		schema := s.Schemas[idx]
		tables := make([]*mgmtv1alpha1.AiGenerateSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &mgmtv1alpha1.AiGenerateSourceTableOption{
				Table:    table.Table,
				RowCount: table.RowCount,
			}
		}
		dto.Schemas[idx] = &mgmtv1alpha1.AiGenerateSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}
	return dto
}
func (s *AiGenerateSourceOptions) FromDto(dto *mgmtv1alpha1.AiGenerateSourceOptions) {
	s.FkSourceConnectionId = dto.FkSourceConnectionId
	s.Schemas = FromDtoAiGenerateSourceSchemaOptions(dto.Schemas)
	s.AiConnectionId = dto.AiConnectionId
	s.ModelName = dto.ModelName
	s.UserPrompt = dto.UserPrompt
}

func FromDtoAiGenerateSourceSchemaOptions(dtos []*mgmtv1alpha1.AiGenerateSourceSchemaOption) []*AiGenerateSourceSchemaOption {
	output := make([]*AiGenerateSourceSchemaOption, len(dtos))
	for idx := range dtos {
		schema := dtos[idx]
		tables := make([]*AiGenerateSourceTableOption, len(schema.Tables))
		for tidx := range schema.Tables {
			table := schema.Tables[tidx]
			tables[tidx] = &AiGenerateSourceTableOption{
				Table:    table.Table,
				RowCount: table.RowCount,
			}
		}
		output[idx] = &AiGenerateSourceSchemaOption{
			Schema: schema.Schema,
			Tables: tables,
		}
	}

	return output
}

type PostgresSourceSchemaOption struct {
	Schema string                       `json:"schema"`
	Tables []*PostgresSourceTableOption `json:"tables"`
}
type PostgresSourceTableOption struct {
	Table       string  `json:"table"`
	WhereClause *string `json:"whereClause,omitempty"`
}

type MysqlSourceSchemaOption struct {
	Schema string                    `json:"schema"`
	Tables []*MysqlSourceTableOption `json:"tables"`
}
type MysqlSourceTableOption struct {
	Table       string  `json:"table"`
	WhereClause *string `json:"whereClause,omitempty"`
}

func (j *JobSourceOptions) ToDto() *mgmtv1alpha1.JobSourceOptions {
	if j.PostgresOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Postgres{
				Postgres: j.PostgresOptions.ToDto(),
			},
		}
	}
	if j.MysqlOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Mysql{
				Mysql: j.MysqlOptions.ToDto(),
			},
		}
	}
	if j.GenerateOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_Generate{
				Generate: j.GenerateOptions.ToDto(),
			},
		}
	}
	if j.AiGenerateOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_AiGenerate{
				AiGenerate: j.AiGenerateOptions.ToDto(),
			},
		}
	}
	return nil
}

func (j *JobSourceOptions) FromDto(dto *mgmtv1alpha1.JobSourceOptions) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_Postgres:
		sqlOpts := &PostgresSourceOptions{}
		sqlOpts.FromDto(config.Postgres)
		j.PostgresOptions = sqlOpts
	case *mgmtv1alpha1.JobSourceOptions_Mysql:
		sqlOpts := &MysqlSourceOptions{}
		sqlOpts.FromDto(config.Mysql)
		j.MysqlOptions = sqlOpts
	case *mgmtv1alpha1.JobSourceOptions_Generate:
		genOpts := &GenerateSourceOptions{}
		genOpts.FromDto(config.Generate)
		j.GenerateOptions = genOpts
	case *mgmtv1alpha1.JobSourceOptions_AiGenerate:
		genOpts := &AiGenerateSourceOptions{}
		genOpts.FromDto(config.AiGenerate)
		j.AiGenerateOptions = genOpts
	default:
		return fmt.Errorf("invalid job source options config, received type: %T", config)
	}
	return nil
}

type JobDestinationOptions struct {
	PostgresOptions *PostgresDestinationOptions `json:"postgresOptions,omitempty"`
	AwsS3Options    *AwsS3DestinationOptions    `json:"awsS3Options,omitempty"`
	MysqlOptions    *MysqlDestinationOptions    `json:"mysqlOptions,omitempty"`
}
type AwsS3DestinationOptions struct{}
type PostgresDestinationOptions struct {
	TruncateTableConfig *PostgresTruncateTableConfig `json:"truncateTableconfig,omitempty"`
	InitTableSchema     bool                         `json:"initTableSchema"`
	OnConflictConfig    *PostgresOnConflictConfig    `json:"onConflictConfig,omitempty"`
}

type PostgresOnConflictConfig struct {
	DoNothing bool `json:"doNothing"`
}

func (t *PostgresOnConflictConfig) ToDto() *mgmtv1alpha1.PostgresOnConflictConfig {
	return &mgmtv1alpha1.PostgresOnConflictConfig{
		DoNothing: t.DoNothing,
	}
}

func (t *PostgresOnConflictConfig) FromDto(dto *mgmtv1alpha1.PostgresOnConflictConfig) {
	t.DoNothing = dto.DoNothing
}

type PostgresTruncateTableConfig struct {
	TruncateBeforeInsert bool `json:"truncateBeforeInsert"`
	TruncateCascade      bool `json:"truncateCascade"`
}

func (t *PostgresTruncateTableConfig) ToDto() *mgmtv1alpha1.PostgresTruncateTableConfig {
	return &mgmtv1alpha1.PostgresTruncateTableConfig{
		TruncateBeforeInsert: t.TruncateBeforeInsert,
		Cascade:              t.TruncateCascade,
	}
}

func (t *PostgresTruncateTableConfig) FromDto(dto *mgmtv1alpha1.PostgresTruncateTableConfig) {
	t.TruncateBeforeInsert = dto.TruncateBeforeInsert
	t.TruncateCascade = dto.Cascade
}

type MysqlDestinationOptions struct {
	TruncateTableConfig *MysqlTruncateTableConfig `json:"truncateTableConfig,omitempty"`
	InitTableSchema     bool                      `json:"initTableSchema"`
	OnConflictConfig    *MysqlOnConflictConfig    `json:"onConflict,omitempty"`
}

type MysqlOnConflictConfig struct {
	DoNothing bool `json:"doNothing"`
}

func (t *MysqlOnConflictConfig) ToDto() *mgmtv1alpha1.MysqlOnConflictConfig {
	return &mgmtv1alpha1.MysqlOnConflictConfig{
		DoNothing: t.DoNothing,
	}
}

func (t *MysqlOnConflictConfig) FromDto(dto *mgmtv1alpha1.MysqlOnConflictConfig) {
	t.DoNothing = dto.DoNothing
}

type MysqlTruncateTableConfig struct {
	TruncateBeforeInsert bool `json:"truncateBeforeInsert"`
}

func (t *MysqlTruncateTableConfig) ToDto() *mgmtv1alpha1.MysqlTruncateTableConfig {
	return &mgmtv1alpha1.MysqlTruncateTableConfig{
		TruncateBeforeInsert: t.TruncateBeforeInsert,
	}
}

func (t *MysqlTruncateTableConfig) FromDto(dto *mgmtv1alpha1.MysqlTruncateTableConfig) {
	t.TruncateBeforeInsert = dto.TruncateBeforeInsert
}

func (j *JobDestinationOptions) ToDto() *mgmtv1alpha1.JobDestinationOptions {
	if j.PostgresOptions != nil {
		if j.PostgresOptions.TruncateTableConfig == nil {
			j.PostgresOptions.TruncateTableConfig = &PostgresTruncateTableConfig{}
		}
		if j.PostgresOptions.OnConflictConfig == nil {
			j.PostgresOptions.OnConflictConfig = &PostgresOnConflictConfig{}
		}
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
				PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
					TruncateTable:   j.PostgresOptions.TruncateTableConfig.ToDto(),
					InitTableSchema: j.PostgresOptions.InitTableSchema,
					OnConflict:      j.PostgresOptions.OnConflictConfig.ToDto(),
				},
			},
		}
	}
	if j.MysqlOptions != nil {
		if j.MysqlOptions.TruncateTableConfig == nil {
			j.MysqlOptions.TruncateTableConfig = &MysqlTruncateTableConfig{}
		}
		if j.MysqlOptions.OnConflictConfig == nil {
			j.MysqlOptions.OnConflictConfig = &MysqlOnConflictConfig{}
		}
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
				MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
					TruncateTable:   j.MysqlOptions.TruncateTableConfig.ToDto(),
					InitTableSchema: j.MysqlOptions.InitTableSchema,
					OnConflict:      j.MysqlOptions.OnConflictConfig.ToDto(),
				},
			},
		}
	}
	if j.AwsS3Options != nil {
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_AwsS3Options{
				AwsS3Options: &mgmtv1alpha1.AwsS3DestinationConnectionOptions{},
			},
		}
	}

	return nil
}

func (j *JobDestinationOptions) FromDto(dto *mgmtv1alpha1.JobDestinationOptions) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.JobDestinationOptions_PostgresOptions:
		truncateCfg := &PostgresTruncateTableConfig{}
		truncateCfg.FromDto(config.PostgresOptions.TruncateTable)
		j.PostgresOptions = &PostgresDestinationOptions{
			InitTableSchema:     config.PostgresOptions.InitTableSchema,
			TruncateTableConfig: truncateCfg,
		}
		if config.PostgresOptions.OnConflict != nil {
			onConflictCfg := &PostgresOnConflictConfig{}
			onConflictCfg.FromDto(config.PostgresOptions.OnConflict)
			j.PostgresOptions.OnConflictConfig = onConflictCfg
		}
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		truncateCfg := &MysqlTruncateTableConfig{}
		truncateCfg.FromDto(config.MysqlOptions.TruncateTable)
		j.MysqlOptions = &MysqlDestinationOptions{
			InitTableSchema:     config.MysqlOptions.InitTableSchema,
			TruncateTableConfig: truncateCfg,
		}
		if config.MysqlOptions.OnConflict != nil {
			onConflictCfg := &MysqlOnConflictConfig{}
			onConflictCfg.FromDto(config.MysqlOptions.OnConflict)
			j.MysqlOptions.OnConflictConfig = onConflictCfg
		}
	case *mgmtv1alpha1.JobDestinationOptions_AwsS3Options:
		j.AwsS3Options = &AwsS3DestinationOptions{}
	default:
		return fmt.Errorf("invalid job destination options config")
	}
	return nil
}

type TemporalConfig struct {
	Namespace        string `json:"namespace"`
	SyncJobQueueName string `json:"syncJobQueueName"`
	Url              string `json:"url"`
}

func (t *TemporalConfig) ToDto() *mgmtv1alpha1.AccountTemporalConfig {
	return &mgmtv1alpha1.AccountTemporalConfig{
		Url:              t.Url,
		Namespace:        t.Namespace,
		SyncJobQueueName: t.SyncJobQueueName,
	}
}

func (t *TemporalConfig) FromDto(dto *mgmtv1alpha1.AccountTemporalConfig) {
	t.Namespace = dto.Namespace
	t.SyncJobQueueName = dto.SyncJobQueueName
	t.Url = dto.Url
}

type WorkflowOptions struct {
	RunTimeout *int64 `json:"runTimeout,omitempty"`
}

func (a *WorkflowOptions) ToDto() *mgmtv1alpha1.WorkflowOptions {
	return &mgmtv1alpha1.WorkflowOptions{
		RunTimeout: a.RunTimeout,
	}
}

func (a *WorkflowOptions) FromDto(dto *mgmtv1alpha1.WorkflowOptions) {
	a.RunTimeout = dto.RunTimeout
}

type ActivityOptions struct {
	ScheduleToCloseTimeout *int64       `json:"scheduleToCloseTimeout,omitempty"`
	StartToCloseTimeout    *int64       `json:"startToCloseTimeout,omitempty"`
	RetryPolicy            *RetryPolicy `json:"retryPolicy,omitempty"`
}

func (a *ActivityOptions) ToDto() *mgmtv1alpha1.ActivityOptions {
	var retryPolicy *mgmtv1alpha1.RetryPolicy
	if a.RetryPolicy != nil {
		retryPolicy = a.RetryPolicy.ToDto()
	}
	return &mgmtv1alpha1.ActivityOptions{
		ScheduleToCloseTimeout: a.ScheduleToCloseTimeout,
		StartToCloseTimeout:    a.StartToCloseTimeout,
		RetryPolicy:            retryPolicy,
	}
}

func (a *ActivityOptions) FromDto(dto *mgmtv1alpha1.ActivityOptions) {
	a.ScheduleToCloseTimeout = dto.ScheduleToCloseTimeout
	a.StartToCloseTimeout = dto.StartToCloseTimeout
	if dto.RetryPolicy != nil {
		a.RetryPolicy = &RetryPolicy{}
		a.RetryPolicy.FromDto(dto.RetryPolicy)
	}
}

type RetryPolicy struct {
	MaximumAttempts *int32 `json:"maximumAttempts,omitempty"`
}

func (r *RetryPolicy) ToDto() *mgmtv1alpha1.RetryPolicy {
	return &mgmtv1alpha1.RetryPolicy{
		MaximumAttempts: r.MaximumAttempts,
	}
}

func (r *RetryPolicy) FromDto(dto *mgmtv1alpha1.RetryPolicy) {
	r.MaximumAttempts = dto.MaximumAttempts
}

type AccountOnboardingConfig struct {
	HasCreatedSourceConnection      bool `json:"hasCreatedSourceConnection"`
	HasCreatedDestinationConnection bool `json:"hasCreatedDestinationConnection"`
	HasCreatedJob                   bool `json:"hasCreatedJob"`
	HasInvitedMembers               bool `json:"hasInvitedMembers"`
}

func (t *AccountOnboardingConfig) ToDto() *mgmtv1alpha1.AccountOnboardingConfig {
	return &mgmtv1alpha1.AccountOnboardingConfig{
		HasCreatedSourceConnection:      t.HasCreatedSourceConnection,
		HasCreatedDestinationConnection: t.HasCreatedDestinationConnection,
		HasCreatedJob:                   t.HasCreatedJob,
		HasInvitedMembers:               t.HasInvitedMembers,
	}
}

func (t *AccountOnboardingConfig) FromDto(dto *mgmtv1alpha1.AccountOnboardingConfig) {
	t.HasCreatedSourceConnection = dto.GetHasCreatedSourceConnection()
	t.HasCreatedDestinationConnection = dto.GetHasCreatedDestinationConnection()
	t.HasCreatedJob = dto.GetHasCreatedJob()
	t.HasInvitedMembers = dto.GetHasInvitedMembers()
}
