package pg_models

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type ConnectionConfig struct {
	PgConfig    *PostgresConnectionConfig `json:"pgConfig,omitempty"`
	AwsS3Config *AwsS3ConnectionConfig    `json:"awsS3Config,omitempty"`
	MysqlConfig *MysqlConnectionConfig    `json:"mysqlConfig,omitempty"`
}

func (c *ConnectionConfig) ToDto() *mgmtv1alpha1.ConnectionConfig {
	if c.PgConfig != nil {
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
					},
				},
			}
		}
	} else if c.MysqlConfig != nil {
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
	}
	return nil
}

func (c *ConnectionConfig) FromDto(dto *mgmtv1alpha1.ConnectionConfig) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.ConnectionConfig_PgConfig:
		c.PgConfig = &PostgresConnectionConfig{}
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
	default:
		return fmt.Errorf("invalid connection config")
	}
	return nil
}

type PostgresConnectionConfig struct {
	Connection *PostgresConnection `json:"connection,omitempty"`
	Url        *string             `json:"url,omitempty"`
}

type PostgresConnection struct {
	Host    string  `json:"host"`
	Port    int32   `json:"port"`
	Name    string  `json:"name"`
	User    string  `json:"user"`
	Pass    string  `json:"pass"`
	SslMode *string `json:"sslMode,omitempty"`
}

type MysqlConnectionConfig struct {
	Connection *MysqlConnection `json:"connection,omitempty"`
	Url        *string          `json:"url,omitempty"`
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
	PathPrefix  *string
	Credentials *AwsS3Credentials
	Region      *string
	Endpoint    *string
}

func (a *AwsS3ConnectionConfig) ToDto() *mgmtv1alpha1.AwsS3ConnectionConfig {
	return &mgmtv1alpha1.AwsS3ConnectionConfig{
		BucketArn:   a.BucketArn,
		PathPrefix:  a.PathPrefix,
		Credentials: a.Credentials.ToDto(),
		Region:      a.Region,
		Endpoint:    a.Endpoint,
	}
}
func (a *AwsS3ConnectionConfig) FromDto(dto *mgmtv1alpha1.AwsS3ConnectionConfig) error {
	a.BucketArn = dto.BucketArn
	a.PathPrefix = dto.PathPrefix
	a.Credentials = &AwsS3Credentials{}
	a.Credentials.FromDto(dto.Credentials)
	a.Region = dto.Region
	a.Endpoint = dto.Endpoint
	return nil
}

type JobMapping struct {
	Schema      string       `json:"schema"`
	Table       string       `json:"table"`
	Column      string       `json:"column"`
	Transformer *Transformer `json:"transformer,omitempty"`
}

func (jm *JobMapping) ToDto() *mgmtv1alpha1.JobMapping {

	return &mgmtv1alpha1.JobMapping{
		Schema:      jm.Schema,
		Table:       jm.Table,
		Column:      jm.Column,
		Transformer: jm.Transformer.ToTransformerDto(),
	}
}

func (jm *JobMapping) FromDto(dto *mgmtv1alpha1.JobMapping) error {
	t := &Transformer{}
	if err := t.FromTransformerDto(dto.Transformer); err != nil {
		return err
	}
	jm.Schema = dto.Schema
	jm.Table = dto.Table
	jm.Column = dto.Column
	jm.Transformer = t
	return nil
}

type JobSourceOptions struct {
	PostgresOptions *PostgresSourceOptions `json:"postgresOptions,omitempty"`
	MysqlOptions    *MysqlSourceOptions    `json:"mysqlOptions,omitempty"`
	GenerateOptions *GenerateSourceOptions `json:"generateOptions,omitempty"`
}

type MysqlSourceOptions struct {
	HaltOnNewColumnAddition bool                       `json:"haltOnNewColumnAddition"`
	Schemas                 []*MysqlSourceSchemaOption `json:"schemas"`
	ConnectionId            string                     `json:"connectionId"`
}
type PostgresSourceOptions struct {
	HaltOnNewColumnAddition bool                          `json:"haltOnNewColumnAddition"`
	Schemas                 []*PostgresSourceSchemaOption `json:"schemas"`
	ConnectionId            string                        `json:"connectionId"`
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

func (s *PostgresSourceOptions) ToDto() *mgmtv1alpha1.PostgresSourceConnectionOptions {
	dto := &mgmtv1alpha1.PostgresSourceConnectionOptions{
		HaltOnNewColumnAddition: s.HaltOnNewColumnAddition,
		ConnectionId:            s.ConnectionId,
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
		HaltOnNewColumnAddition: s.HaltOnNewColumnAddition,
		ConnectionId:            s.ConnectionId,
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
	default:
		return fmt.Errorf("invalid job source options config")
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
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_PostgresOptions{
				PostgresOptions: &mgmtv1alpha1.PostgresDestinationConnectionOptions{
					TruncateTable:   j.PostgresOptions.TruncateTableConfig.ToDto(),
					InitTableSchema: j.PostgresOptions.InitTableSchema,
				},
			},
		}
	}
	if j.MysqlOptions != nil {
		if j.MysqlOptions.TruncateTableConfig == nil {
			j.MysqlOptions.TruncateTableConfig = &MysqlTruncateTableConfig{}
		}
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_MysqlOptions{
				MysqlOptions: &mgmtv1alpha1.MysqlDestinationConnectionOptions{
					TruncateTable:   j.MysqlOptions.TruncateTableConfig.ToDto(),
					InitTableSchema: j.MysqlOptions.InitTableSchema,
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
	case *mgmtv1alpha1.JobDestinationOptions_MysqlOptions:
		truncateCfg := &MysqlTruncateTableConfig{}
		truncateCfg.FromDto(config.MysqlOptions.TruncateTable)
		j.MysqlOptions = &MysqlDestinationOptions{
			InitTableSchema:     config.MysqlOptions.InitTableSchema,
			TruncateTableConfig: truncateCfg,
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
