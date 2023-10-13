package jsonmodels

import (
	"fmt"

	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
)

type ConnectionConfig struct {
	PgConfig    *PostgresConnectionConfig
	AwsS3Config *AwsS3ConnectionConfig
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
	Exclude     bool
}

func (jm *JobMapping) ToDto() *mgmtv1alpha1.JobMapping {

	return &mgmtv1alpha1.JobMapping{
		Schema:      jm.Schema,
		Table:       jm.Table,
		Column:      jm.Column,
		Transformer: jm.Transformer.ToDto(),
		Exclude:     jm.Exclude,
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
	jm.Exclude = dto.Exclude
	return nil
}

type Transformer struct {
	Value  string
	Config *TransformerConfigs
}

type TransformerConfigs struct {
	EmailConfig *EmailConfigs
	FirstName   *FirstNameConfig
	Uuidv4      *Uuidv4Config
	PhoneNumber *PhoneNumberConfig
	Passthrough *PassthroughConfig
}

type EmailConfigs struct {
	PreserveLength bool
	PreserveDomain bool
}

type FirstNameConfig struct {
}
type Uuidv4Config struct {
}
type PhoneNumberConfig struct {
}
type PassthroughConfig struct {
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
			FirstName: &FirstNameConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_PassthroughConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Passthrough: &PassthroughConfig{},
		}
	case *mgmtv1alpha1.TransformerConfig_UuidConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			Uuidv4: &Uuidv4Config{},
		}
	case *mgmtv1alpha1.TransformerConfig_PhoneNumberConfig:
		t.Value = tr.Value
		t.Config = &TransformerConfigs{
			PhoneNumber: &PhoneNumberConfig{},
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
					FirstNameConfig: &mgmtv1alpha1.FirstName{},
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
					PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{},
				},
			},
		}
	case t.Config.Uuidv4 != nil:
		return &mgmtv1alpha1.Transformer{
			Value: t.Value,
			Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
					UuidConfig: &mgmtv1alpha1.Uuidv4{},
				},
			},
		}
	default:
		return &mgmtv1alpha1.Transformer{Value: t.Value}
	}
}

type JobSourceOptions struct {
	SqlOptions *SqlSourceOptions
}
type SqlSourceOptions struct {
	HaltOnNewColumnAddition bool
}

func (j *JobSourceOptions) ToDto() *mgmtv1alpha1.JobSourceOptions {
	if j.SqlOptions != nil {
		return &mgmtv1alpha1.JobSourceOptions{
			Config: &mgmtv1alpha1.JobSourceOptions_SqlOptions{
				SqlOptions: &mgmtv1alpha1.SqlSourceConnectionOptions{
					HaltOnNewColumnAddition: j.SqlOptions.HaltOnNewColumnAddition,
				},
			},
		}
	}
	return nil
}

func (j *JobSourceOptions) FromDto(dto *mgmtv1alpha1.JobSourceOptions) error {
	switch config := dto.Config.(type) {
	case *mgmtv1alpha1.JobSourceOptions_SqlOptions:
		j.SqlOptions = &SqlSourceOptions{
			HaltOnNewColumnAddition: config.SqlOptions.HaltOnNewColumnAddition,
		}
	default:
		return fmt.Errorf("invalid config")
	}
	return nil
}

type JobDestinationOptions struct {
	SqlOptions   *SqlDestinationOptions
	AwsS3Options *AwsS3DestinationOptions
}
type AwsS3DestinationOptions struct{}
type SqlDestinationOptions struct {
	TruncateTableConfig *TruncateTableConfig
	InitTableSchema     bool
}
type TruncateTableConfig struct {
	TruncateBeforeInsert bool
	TruncateCascade      bool
}

func (t *TruncateTableConfig) ToDto() *mgmtv1alpha1.TruncateTableConfig {
	return &mgmtv1alpha1.TruncateTableConfig{
		TruncateBeforeInsert: t.TruncateBeforeInsert,
		Cascade:              t.TruncateCascade,
	}
}

func (t *TruncateTableConfig) FromDto(dto *mgmtv1alpha1.TruncateTableConfig) {
	t.TruncateBeforeInsert = dto.TruncateBeforeInsert
	t.TruncateCascade = dto.Cascade
}

func (j *JobDestinationOptions) ToDto() *mgmtv1alpha1.JobDestinationOptions {
	if j.SqlOptions != nil {
		if j.SqlOptions.TruncateTableConfig == nil {
			j.SqlOptions.TruncateTableConfig = &TruncateTableConfig{}
		}
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_SqlOptions{
				SqlOptions: &mgmtv1alpha1.SqlDestinationConnectionOptions{
					TruncateTable:   j.SqlOptions.TruncateTableConfig.ToDto(),
					InitTableSchema: j.SqlOptions.InitTableSchema,
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
	case *mgmtv1alpha1.JobDestinationOptions_SqlOptions:
		truncateCfg := &TruncateTableConfig{}
		truncateCfg.FromDto(config.SqlOptions.TruncateTable)
		j.SqlOptions = &SqlDestinationOptions{
			InitTableSchema:     config.SqlOptions.InitTableSchema,
			TruncateTableConfig: truncateCfg,
		}
	case *mgmtv1alpha1.JobDestinationOptions_AwsS3Options:
		j.AwsS3Options = &AwsS3DestinationOptions{}
	default:
		return fmt.Errorf("invalid config")
	}
	return nil
}
