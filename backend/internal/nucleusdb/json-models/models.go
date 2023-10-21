package jsonmodels

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type ConnectionConfig struct {
	PgConfig    *PostgresConnectionConfig
	AwsS3Config *AwsS3ConnectionConfig
	MysqlConfig *MysqlConnectionConfig
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
		return fmt.Errorf("invalid config")
	}
	return nil
}

type PostgresConnectionConfig struct {
	Connection *PostgresConnection
	Url        *string
}

type PostgresConnection struct {
	Host    string
	Port    int32
	Name    string
	User    string
	Pass    string
	SslMode *string
}

type MysqlConnectionConfig struct {
	Connection *MysqlConnection
	Url        *string
}

type MysqlConnection struct {
	User     string
	Pass     string
	Protocol string
	Host     string
	Port     int32
	Name     string
}

type AwsS3Credentials struct {
	Profile         *string
	AccessKeyId     *string
	SecretAccessKey *string
	SessionToken    *string
	FromEc2Role     *bool
	RoleArn         *string
	RoleExternalId  *string
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
	Schema      string
	Table       string
	Column      string
	Transformer *Transformer
}

func (jm *JobMapping) ToDto() *mgmtv1alpha1.JobMapping {

	return &mgmtv1alpha1.JobMapping{
		Schema:      jm.Schema,
		Table:       jm.Table,
		Column:      jm.Column,
		Transformer: jm.Transformer.ToDto(),
	}
}

func (jm *JobMapping) FromDto(dto *mgmtv1alpha1.JobMapping) error {
	t := &Transformer{}
	if err := t.FromDto(dto.Transformer); err != nil {
		return err
	}
	jm.Schema = dto.Schema
	jm.Table = dto.Table
	jm.Column = dto.Column
	jm.Transformer = t
	return nil
}

type Transformer struct {
	Value  string
	Config *TransformerConfigs
}

type TransformerConfigs struct {
	EmailConfig    *EmailConfigs
	FirstName      *FirstNameConfig
	LastName       *LastNameConfig
	FullName       *FullNameConfig
	Uuid           *UuidConfig
	PhoneNumber    *PhoneNumberConfig
	IntPhoneNumber *IntPhoneNumberConfig
	Passthrough    *PassthroughConfig
	Null           *NullConfig
	RandomString   *RandomStringConfig
	RandomBool     *RandomBoolConfig
	RandomInt      *RandomIntConfig
}

type EmailConfigs struct {
	PreserveLength bool
	PreserveDomain bool
}

type FirstNameConfig struct {
	PreserveLength bool
}

type LastNameConfig struct {
	PreserveLength bool
}

type FullNameConfig struct {
	PreserveLength bool
}
type UuidConfig struct {
	IncludeHyphen bool
}
type PhoneNumberConfig struct {
	IncludeHyphens bool
	E164Format     bool
	PreserveLength bool
}

type IntPhoneNumberConfig struct {
	PreserveLength bool
}
type PassthroughConfig struct {
}

type NullConfig struct{}

type RandomStringConfig struct {
	PreserveLength bool
	StrLength      int64
	StrCase        string
}

type RandomBoolConfig struct{}

type RandomIntConfig struct {
	PreserveLength bool
	IntLength      int64
}

// from API -> DB
func (t *Transformer) FromDto(tr *mgmtv1alpha1.Transformer) error {

	switch tr.Config.Config.(type) {
	case *mgmtv1alpha1.TransformerConfig_EmailConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			EmailConfig: &EmailConfigs{
				PreserveLength: tr.Config.GetEmailConfig().PreserveLength,
				PreserveDomain: tr.Config.GetEmailConfig().PreserveDomain,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_FirstNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			FirstName: &FirstNameConfig{
				PreserveLength: tr.Config.GetFirstNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_LastNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			LastName: &LastNameConfig{
				PreserveLength: tr.Config.GetLastNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_FullNameConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			FullName: &FullNameConfig{
				PreserveLength: tr.Config.GetFullNameConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Passthrough: &PassthroughConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_UuidConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Uuid: &UuidConfig{
				IncludeHyphen: tr.Config.GetUuidConfig().IncludeHyphen,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_PhoneNumberConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			PhoneNumber: &PhoneNumberConfig{
				IncludeHyphens: tr.Config.GetPhoneNumberConfig().IncludeHyphens,
				E164Format:     tr.Config.GetPhoneNumberConfig().E164Format,
				PreserveLength: tr.Config.GetPhoneNumberConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			IntPhoneNumber: &IntPhoneNumberConfig{
				PreserveLength: tr.Config.GetIntPhoneNumberConfig().PreserveLength,
			},
		}
	case *mgmtv1alpha1.TransformerConfig_NullConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Null: &NullConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomStringConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomString: &RandomStringConfig{
				PreserveLength: tr.Config.GetRandomStringConfig().PreserveLength,
				StrLength:      tr.Config.GetRandomStringConfig().GetStrLength(),
				StrCase:        tr.Config.GetRandomStringConfig().StrCase.String(),
			},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomBoolConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomBool: &RandomBoolConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_RandomIntConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			RandomInt: &RandomIntConfig{
				PreserveLength: tr.Config.GetRandomIntConfig().PreserveLength,
				IntLength:      tr.Config.GetRandomIntConfig().IntLength,
			},
		}
	default:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{}
	}

	return nil
}

// DB -> API
func (t *Transformer) ToDto() *mgmtv1alpha1.Transformer {

	switch {
	case t.Config.EmailConfig != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
					EmailConfig: &mgmtv1alpha1.EmailConfig{
						PreserveDomain: t.Config.EmailConfig.PreserveDomain,
						PreserveLength: t.Config.EmailConfig.PreserveLength,
					},
				},
			},
		}
	case t.Config.FirstName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FirstNameConfig{
					FirstNameConfig: &mgmtv1alpha1.FirstName{
						PreserveLength: t.Config.FirstName.PreserveLength,
					},
				},
			},
		}
	case t.Config.LastName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_LastNameConfig{
					LastNameConfig: &mgmtv1alpha1.LastName{
						PreserveLength: t.Config.LastName.PreserveLength,
					},
				},
			},
		}
	case t.Config.FullName != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FullNameConfig{
					FullNameConfig: &mgmtv1alpha1.FullName{
						PreserveLength: t.Config.FullName.PreserveLength,
					},
				},
			},
		}
	case t.Config.Passthrough != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PassthroughConfig{
					PassthroughConfig: &mgmtv1alpha1.Passthrough{},
				},
			},
		}
	case t.Config.PhoneNumber != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PhoneNumberConfig{
					PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{
						PreserveLength: t.Config.PhoneNumber.PreserveLength,
						E164Format:     t.Config.PhoneNumber.E164Format,
						IncludeHyphens: t.Config.PhoneNumber.IncludeHyphens,
					},
				},
			},
		}
	case t.Config.IntPhoneNumber != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig{
					IntPhoneNumberConfig: &mgmtv1alpha1.IntPhoneNumber{
						PreserveLength: t.Config.IntPhoneNumber.PreserveLength,
					},
				},
			},
		}
	case t.Config.Uuid != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
					UuidConfig: &mgmtv1alpha1.Uuid{
						IncludeHyphen: t.Config.Uuid.IncludeHyphen,
					},
				},
			},
		}
	case t.Config.Null != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_NullConfig{
					NullConfig: &mgmtv1alpha1.Null{},
				},
			},
		}
	case t.Config.RandomString != nil:

		strCase, err := StrCaseFromString(t.Config.RandomString.StrCase)
		if err != nil {
			return &mgmtv1alpha1.Transformer{Value: t.Value}
		}

		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomStringConfig{
					RandomStringConfig: &mgmtv1alpha1.RandomString{
						PreserveLength: t.Config.RandomString.PreserveLength,
						StrLength:      t.Config.RandomString.StrLength,
						StrCase:        strCase,
					},
				},
			},
		}
	case t.Config.RandomBool != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomBoolConfig{
					RandomBoolConfig: &mgmtv1alpha1.RandomBool{},
				},
			},
		}
	case t.Config.RandomInt != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_RandomIntConfig{
					RandomIntConfig: &mgmtv1alpha1.RandomInt{
						PreserveLength: t.Config.RandomInt.PreserveLength,
						IntLength:      t.Config.RandomInt.IntLength,
					},
				},
			},
		}
	default:
		return &mgmtv1alpha1.Transformer{Value: t.Value}
	}
}

func StrCaseFromString(strCase string) (mgmtv1alpha1.RandomString_StringCase, error) {
	switch strCase {
	case "UPPER":
		return mgmtv1alpha1.RandomString_STRING_CASE_UPPER, nil
	case "LOWER":
		return mgmtv1alpha1.RandomString_STRING_CASE_LOWER, nil
	case "TITLE":
		return mgmtv1alpha1.RandomString_STRING_CASE_TITLE, nil
	default:
		return mgmtv1alpha1.RandomString_STRING_CASE_LOWER, fmt.Errorf("invalid string case: %s", strCase)
	}
}

type JobSourceOptions struct {
	PostgresOptions *PostgresSourceOptions
	MysqlOptions    *MysqlSourceOptions
}

type MysqlSourceOptions struct {
	HaltOnNewColumnAddition bool
	Schemas                 []*MysqlSourceSchemaOption
}
type PostgresSourceOptions struct {
	HaltOnNewColumnAddition bool
	Schemas                 []*PostgresSourceSchemaOption
}

func (s *PostgresSourceOptions) ToDto() *mgmtv1alpha1.PostgresSourceConnectionOptions {
	dto := &mgmtv1alpha1.PostgresSourceConnectionOptions{
		HaltOnNewColumnAddition: s.HaltOnNewColumnAddition,
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

type PostgresSourceSchemaOption struct {
	Schema string
	Tables []*PostgresSourceTableOption
}
type PostgresSourceTableOption struct {
	Table       string
	WhereClause *string
}

type MysqlSourceSchemaOption struct {
	Schema string
	Tables []*MysqlSourceTableOption
}
type MysqlSourceTableOption struct {
	Table       string
	WhereClause *string
}

func (j *JobSourceOptions) ToDto() *mgmtv1alpha1.JobSourceOptions {
	if j.PostgresOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_PostgresOptions{
				PostgresOptions: j.PostgresOptions.ToDto(),
			},
		}
	}
	if j.MysqlOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_MysqlOptions{
				MysqlOptions: j.MysqlOptions.ToDto(),
			},
		}
	}
	return nil
}

func (j *JobSourceOptions) FromDto(dto *mgmtv1alpha1.JobSourceOptions) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_PostgresOptions:
		sqlOpts := &PostgresSourceOptions{}
		sqlOpts.FromDto(config.PostgresOptions)
		j.PostgresOptions = sqlOpts
	case *mgmtv1alpha1.JobSourceOptions_MysqlOptions:
		sqlOpts := &MysqlSourceOptions{}
		sqlOpts.FromDto(config.MysqlOptions)
		j.MysqlOptions = sqlOpts
	default:
		return fmt.Errorf("invalid config")
	}
	return nil
}

type JobDestinationOptions struct {
	PostgresOptions *PostgresDestinationOptions
	AwsS3Options    *AwsS3DestinationOptions
	MysqlOptions    *MysqlDestinationOptions
}
type AwsS3DestinationOptions struct{}
type PostgresDestinationOptions struct {
	TruncateTableConfig *PostgresTruncateTableConfig
	InitTableSchema     bool
}
type PostgresTruncateTableConfig struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool
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
	TruncateTableConfig *MysqlTruncateTableConfig
	InitTableSchema     bool
}
type MysqlTruncateTableConfig struct {
	TruncateBeforeInsert bool
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
		return fmt.Errorf("invalid config")
	}
	return nil
}
