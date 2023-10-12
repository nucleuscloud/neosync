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
		var credentials *mgmtv1alpha1.AwsS3Credentials
		if c.AwsS3Config.Credentials != nil {
			credentials = &mgmtv1alpha1.AwsS3Credentials{
				AccessKeyId: c.AwsS3Config.Credentials.AccessKeyId,
				AccessKey:   c.AwsS3Config.Credentials.AccessKey,
			}
		}
		return &mgmtv1alpha1.ConnectionConfig{
			Config: &mgmtv1alpha1.ConnectionConfig_AwsS3Config{
				AwsS3Config: &mgmtv1alpha1.AwsS3ConnectionConfig{
					BucketArn:   c.AwsS3Config.BucketArn,
					PathPrefix:  c.AwsS3Config.PathPrefix,
					RoleArn:     c.AwsS3Config.RoleArn,
					Credentials: credentials,
				},
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
		var credentials *AwsS3Credentials
		if config.AwsS3Config.Credentials != nil {
			credentials = &AwsS3Credentials{
				AccessKeyId: config.AwsS3Config.Credentials.AccessKeyId,
				AccessKey:   config.AwsS3Config.Credentials.AccessKey,
			}
		}
		c.AwsS3Config = &AwsS3ConnectionConfig{
			BucketArn:   config.AwsS3Config.BucketArn,
			PathPrefix:  config.AwsS3Config.PathPrefix,
			RoleArn:     config.AwsS3Config.RoleArn,
			Credentials: credentials,
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
	AccessKeyId string
	AccessKey   string
}
type AwsS3ConnectionConfig struct {
	BucketArn   string
	PathPrefix  *string
	RoleArn     *string
	Credentials *AwsS3Credentials
}

type JobMapping struct {
	Schema      string
	Table       string
	Column      string
	Transformer string
	Exclude     bool
}

func (jm *JobMapping) ToDto() *mgmtv1alpha1.JobMapping {
	return &mgmtv1alpha1.JobMapping{
		Schema:      jm.Schema,
		Table:       jm.Table,
		Column:      jm.Column,
		Transformer: jm.Transformer,
		Exclude:     jm.Exclude,
	}
}

func (jm *JobMapping) FromDto(dto *mgmtv1alpha1.JobMapping) error {
	jm.Schema = dto.Schema
	jm.Table = dto.Table
	jm.Column = dto.Column
	jm.Transformer = dto.Transformer
	jm.Exclude = dto.Exclude
	return nil
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
