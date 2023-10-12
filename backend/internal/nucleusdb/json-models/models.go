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

type JobSchemaMapping struct {
	Schema        string
	TableMappings []*JobTableMapping
}

func (j *JobSchemaMapping) ToDto() *mgmtv1alpha1.JobSchemaMapping {
	jsm := &mgmtv1alpha1.JobSchemaMapping{}
	jsm.Schema = j.Schema
	jsm.TableMappings = make([]*mgmtv1alpha1.JobTableMapping, len(j.TableMappings))
	for i := range j.TableMappings {
		jsm.TableMappings[i] = j.TableMappings[i].ToDto()
	}
	return jsm
}
func (j *JobSchemaMapping) FromDto(dto *mgmtv1alpha1.JobSchemaMapping) error {
	j.Schema = dto.Schema
	j.TableMappings = make([]*JobTableMapping, len(dto.TableMappings))
	for i := range dto.TableMappings {
		val := &JobTableMapping{}
		err := val.FromDto(dto.TableMappings[i])
		if err != nil {
			return err
		}
		j.TableMappings[i] = val
	}
	return nil
}

type JobTableMapping struct {
	Table                  string
	ColumnMappings         []*JobColumnMapping
	InitSchemaBeforeInsert bool
	Truncate               TruncateTableConfig
}

func (j *JobTableMapping) ToDto() *mgmtv1alpha1.JobTableMapping {
	jtm := &mgmtv1alpha1.JobTableMapping{
		Table:                  j.Table,
		InitSchemaBeforeInsert: j.InitSchemaBeforeInsert,
		Truncate:               j.Truncate.ToDto(),
		ColumnMappings:         make([]*mgmtv1alpha1.JobColumnMapping, len(j.ColumnMappings)),
	}
	for i := range j.ColumnMappings {
		jtm.ColumnMappings[i] = j.ColumnMappings[i].ToDto()
	}
	return jtm
}
func (j *JobTableMapping) FromDto(dto *mgmtv1alpha1.JobTableMapping) error {
	j.Table = dto.Table
	j.InitSchemaBeforeInsert = dto.InitSchemaBeforeInsert
	j.ColumnMappings = make([]*JobColumnMapping, len(dto.ColumnMappings))
	for i := range dto.ColumnMappings {
		val := &JobColumnMapping{}
		err := val.FromDto(dto.ColumnMappings[i])
		if err != nil {
			return err
		}
		j.ColumnMappings[i] = val
	}
	j.Truncate = TruncateTableConfig{}
	err := j.Truncate.FromDto(dto.Truncate)
	if err != nil {
		return err
	}
	return nil
}

type TruncateTableConfig struct {
	TruncateBeforeInsert bool
	Cascade              bool
}

func (j *TruncateTableConfig) ToDto() *mgmtv1alpha1.TruncateTableConfig {
	return &mgmtv1alpha1.TruncateTableConfig{
		TruncateBeforeInsert: j.TruncateBeforeInsert,
		Cascade:              j.Cascade,
	}
}
func (j *TruncateTableConfig) FromDto(dto *mgmtv1alpha1.TruncateTableConfig) error {
	j.TruncateBeforeInsert = dto.TruncateBeforeInsert
	j.Cascade = dto.Cascade
	return nil
}

type JobColumnMapping struct {
	Column      string
	Transformer string
	Exclude     bool
}

func (jm *JobColumnMapping) ToDto() *mgmtv1alpha1.JobColumnMapping {
	return &mgmtv1alpha1.JobColumnMapping{
		Column:      jm.Column,
		Transformer: jm.Transformer,
		Exclude:     jm.Exclude,
	}
}

func (jm *JobColumnMapping) FromDto(dto *mgmtv1alpha1.JobColumnMapping) error {
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
					HaltOnNewColumnAddition: &j.SqlOptions.HaltOnNewColumnAddition,
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
			HaltOnNewColumnAddition: *config.SqlOptions.HaltOnNewColumnAddition,
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
type SqlDestinationOptions struct{}
type AwsS3DestinationOptions struct{}

func (j *JobDestinationOptions) ToDto() *mgmtv1alpha1.JobDestinationOptions {
	if j.SqlOptions != nil {
		return &mgmtv1alpha1.JobDestinationOptions{
			Config: &mgmtv1alpha1.JobDestinationOptions_SqlOptions{
				SqlOptions: &mgmtv1alpha1.SqlDestinationConnectionOptions{},
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
	switch dto.Config.(type) {
	case *mgmtv1alpha1.JobDestinationOptions_SqlOptions:
		j.SqlOptions = &SqlDestinationOptions{}
	case *mgmtv1alpha1.JobDestinationOptions_AwsS3Options:
		j.AwsS3Options = &AwsS3DestinationOptions{}
	default:
		return fmt.Errorf("invalid config")
	}
	return nil
}
