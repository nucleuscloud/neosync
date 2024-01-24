from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from mgmt.v1alpha1 import transformer_pb2 as _transformer_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class JobStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    JOB_STATUS_UNSPECIFIED: _ClassVar[JobStatus]
    JOB_STATUS_ENABLED: _ClassVar[JobStatus]
    JOB_STATUS_PAUSED: _ClassVar[JobStatus]
    JOB_STATUS_DISABLED: _ClassVar[JobStatus]

class ActivityStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACTIVITY_STATUS_UNSPECIFIED: _ClassVar[ActivityStatus]
    ACTIVITY_STATUS_SCHEDULED: _ClassVar[ActivityStatus]
    ACTIVITY_STATUS_STARTED: _ClassVar[ActivityStatus]
    ACTIVITY_STATUS_CANCELED: _ClassVar[ActivityStatus]
    ACTIVITY_STATUS_FAILED: _ClassVar[ActivityStatus]

class JobRunStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    JOB_RUN_STATUS_UNSPECIFIED: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_PENDING: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_RUNNING: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_COMPLETE: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_ERROR: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_CANCELED: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_TERMINATED: _ClassVar[JobRunStatus]
    JOB_RUN_STATUS_FAILED: _ClassVar[JobRunStatus]
JOB_STATUS_UNSPECIFIED: JobStatus
JOB_STATUS_ENABLED: JobStatus
JOB_STATUS_PAUSED: JobStatus
JOB_STATUS_DISABLED: JobStatus
ACTIVITY_STATUS_UNSPECIFIED: ActivityStatus
ACTIVITY_STATUS_SCHEDULED: ActivityStatus
ACTIVITY_STATUS_STARTED: ActivityStatus
ACTIVITY_STATUS_CANCELED: ActivityStatus
ACTIVITY_STATUS_FAILED: ActivityStatus
JOB_RUN_STATUS_UNSPECIFIED: JobRunStatus
JOB_RUN_STATUS_PENDING: JobRunStatus
JOB_RUN_STATUS_RUNNING: JobRunStatus
JOB_RUN_STATUS_COMPLETE: JobRunStatus
JOB_RUN_STATUS_ERROR: JobRunStatus
JOB_RUN_STATUS_CANCELED: JobRunStatus
JOB_RUN_STATUS_TERMINATED: JobRunStatus
JOB_RUN_STATUS_FAILED: JobRunStatus

class GetJobsRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetJobsResponse(_message.Message):
    __slots__ = ("jobs",)
    JOBS_FIELD_NUMBER: _ClassVar[int]
    jobs: _containers.RepeatedCompositeFieldContainer[Job]
    def __init__(self, jobs: _Optional[_Iterable[_Union[Job, _Mapping]]] = ...) -> None: ...

class JobSource(_message.Message):
    __slots__ = ("options",)
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    options: JobSourceOptions
    def __init__(self, options: _Optional[_Union[JobSourceOptions, _Mapping]] = ...) -> None: ...

class JobSourceOptions(_message.Message):
    __slots__ = ("postgres", "aws_s3", "mysql", "generate")
    POSTGRES_FIELD_NUMBER: _ClassVar[int]
    AWS_S3_FIELD_NUMBER: _ClassVar[int]
    MYSQL_FIELD_NUMBER: _ClassVar[int]
    GENERATE_FIELD_NUMBER: _ClassVar[int]
    postgres: PostgresSourceConnectionOptions
    aws_s3: AwsS3SourceConnectionOptions
    mysql: MysqlSourceConnectionOptions
    generate: GenerateSourceOptions
    def __init__(self, postgres: _Optional[_Union[PostgresSourceConnectionOptions, _Mapping]] = ..., aws_s3: _Optional[_Union[AwsS3SourceConnectionOptions, _Mapping]] = ..., mysql: _Optional[_Union[MysqlSourceConnectionOptions, _Mapping]] = ..., generate: _Optional[_Union[GenerateSourceOptions, _Mapping]] = ...) -> None: ...

class CreateJobDestination(_message.Message):
    __slots__ = ("connection_id", "options")
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    options: JobDestinationOptions
    def __init__(self, connection_id: _Optional[str] = ..., options: _Optional[_Union[JobDestinationOptions, _Mapping]] = ...) -> None: ...

class JobDestination(_message.Message):
    __slots__ = ("connection_id", "options", "id")
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    options: JobDestinationOptions
    id: str
    def __init__(self, connection_id: _Optional[str] = ..., options: _Optional[_Union[JobDestinationOptions, _Mapping]] = ..., id: _Optional[str] = ...) -> None: ...

class GenerateSourceOptions(_message.Message):
    __slots__ = ("schemas", "fk_source_connection_id")
    SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    FK_SOURCE_CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    schemas: _containers.RepeatedCompositeFieldContainer[GenerateSourceSchemaOption]
    fk_source_connection_id: str
    def __init__(self, schemas: _Optional[_Iterable[_Union[GenerateSourceSchemaOption, _Mapping]]] = ..., fk_source_connection_id: _Optional[str] = ...) -> None: ...

class GenerateSourceSchemaOption(_message.Message):
    __slots__ = ("schema", "tables")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLES_FIELD_NUMBER: _ClassVar[int]
    schema: str
    tables: _containers.RepeatedCompositeFieldContainer[GenerateSourceTableOption]
    def __init__(self, schema: _Optional[str] = ..., tables: _Optional[_Iterable[_Union[GenerateSourceTableOption, _Mapping]]] = ...) -> None: ...

class GenerateSourceTableOption(_message.Message):
    __slots__ = ("table", "row_count")
    TABLE_FIELD_NUMBER: _ClassVar[int]
    ROW_COUNT_FIELD_NUMBER: _ClassVar[int]
    table: str
    row_count: int
    def __init__(self, table: _Optional[str] = ..., row_count: _Optional[int] = ...) -> None: ...

class PostgresSourceConnectionOptions(_message.Message):
    __slots__ = ("halt_on_new_column_addition", "schemas", "connection_id")
    HALT_ON_NEW_COLUMN_ADDITION_FIELD_NUMBER: _ClassVar[int]
    SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    halt_on_new_column_addition: bool
    schemas: _containers.RepeatedCompositeFieldContainer[PostgresSourceSchemaOption]
    connection_id: str
    def __init__(self, halt_on_new_column_addition: bool = ..., schemas: _Optional[_Iterable[_Union[PostgresSourceSchemaOption, _Mapping]]] = ..., connection_id: _Optional[str] = ...) -> None: ...

class PostgresSourceSchemaOption(_message.Message):
    __slots__ = ("schema", "tables")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLES_FIELD_NUMBER: _ClassVar[int]
    schema: str
    tables: _containers.RepeatedCompositeFieldContainer[PostgresSourceTableOption]
    def __init__(self, schema: _Optional[str] = ..., tables: _Optional[_Iterable[_Union[PostgresSourceTableOption, _Mapping]]] = ...) -> None: ...

class PostgresSourceTableOption(_message.Message):
    __slots__ = ("table", "where_clause")
    TABLE_FIELD_NUMBER: _ClassVar[int]
    WHERE_CLAUSE_FIELD_NUMBER: _ClassVar[int]
    table: str
    where_clause: str
    def __init__(self, table: _Optional[str] = ..., where_clause: _Optional[str] = ...) -> None: ...

class MysqlSourceConnectionOptions(_message.Message):
    __slots__ = ("halt_on_new_column_addition", "schemas", "connection_id")
    HALT_ON_NEW_COLUMN_ADDITION_FIELD_NUMBER: _ClassVar[int]
    SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    halt_on_new_column_addition: bool
    schemas: _containers.RepeatedCompositeFieldContainer[MysqlSourceSchemaOption]
    connection_id: str
    def __init__(self, halt_on_new_column_addition: bool = ..., schemas: _Optional[_Iterable[_Union[MysqlSourceSchemaOption, _Mapping]]] = ..., connection_id: _Optional[str] = ...) -> None: ...

class MysqlSourceSchemaOption(_message.Message):
    __slots__ = ("schema", "tables")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLES_FIELD_NUMBER: _ClassVar[int]
    schema: str
    tables: _containers.RepeatedCompositeFieldContainer[MysqlSourceTableOption]
    def __init__(self, schema: _Optional[str] = ..., tables: _Optional[_Iterable[_Union[MysqlSourceTableOption, _Mapping]]] = ...) -> None: ...

class MysqlSourceTableOption(_message.Message):
    __slots__ = ("table", "where_clause")
    TABLE_FIELD_NUMBER: _ClassVar[int]
    WHERE_CLAUSE_FIELD_NUMBER: _ClassVar[int]
    table: str
    where_clause: str
    def __init__(self, table: _Optional[str] = ..., where_clause: _Optional[str] = ...) -> None: ...

class AwsS3SourceConnectionOptions(_message.Message):
    __slots__ = ("connection_id",)
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class JobDestinationOptions(_message.Message):
    __slots__ = ("postgres_options", "aws_s3_options", "mysql_options")
    POSTGRES_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    AWS_S3_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    MYSQL_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    postgres_options: PostgresDestinationConnectionOptions
    aws_s3_options: AwsS3DestinationConnectionOptions
    mysql_options: MysqlDestinationConnectionOptions
    def __init__(self, postgres_options: _Optional[_Union[PostgresDestinationConnectionOptions, _Mapping]] = ..., aws_s3_options: _Optional[_Union[AwsS3DestinationConnectionOptions, _Mapping]] = ..., mysql_options: _Optional[_Union[MysqlDestinationConnectionOptions, _Mapping]] = ...) -> None: ...

class PostgresDestinationConnectionOptions(_message.Message):
    __slots__ = ("truncate_table", "init_table_schema")
    TRUNCATE_TABLE_FIELD_NUMBER: _ClassVar[int]
    INIT_TABLE_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    truncate_table: PostgresTruncateTableConfig
    init_table_schema: bool
    def __init__(self, truncate_table: _Optional[_Union[PostgresTruncateTableConfig, _Mapping]] = ..., init_table_schema: bool = ...) -> None: ...

class PostgresTruncateTableConfig(_message.Message):
    __slots__ = ("truncate_before_insert", "cascade")
    TRUNCATE_BEFORE_INSERT_FIELD_NUMBER: _ClassVar[int]
    CASCADE_FIELD_NUMBER: _ClassVar[int]
    truncate_before_insert: bool
    cascade: bool
    def __init__(self, truncate_before_insert: bool = ..., cascade: bool = ...) -> None: ...

class MysqlDestinationConnectionOptions(_message.Message):
    __slots__ = ("truncate_table", "init_table_schema")
    TRUNCATE_TABLE_FIELD_NUMBER: _ClassVar[int]
    INIT_TABLE_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    truncate_table: MysqlTruncateTableConfig
    init_table_schema: bool
    def __init__(self, truncate_table: _Optional[_Union[MysqlTruncateTableConfig, _Mapping]] = ..., init_table_schema: bool = ...) -> None: ...

class MysqlTruncateTableConfig(_message.Message):
    __slots__ = ("truncate_before_insert",)
    TRUNCATE_BEFORE_INSERT_FIELD_NUMBER: _ClassVar[int]
    truncate_before_insert: bool
    def __init__(self, truncate_before_insert: bool = ...) -> None: ...

class AwsS3DestinationConnectionOptions(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CreateJobRequest(_message.Message):
    __slots__ = ("account_id", "job_name", "cron_schedule", "mappings", "source", "destinations", "initiate_job_run")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_NAME_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    DESTINATIONS_FIELD_NUMBER: _ClassVar[int]
    INITIATE_JOB_RUN_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    job_name: str
    cron_schedule: str
    mappings: _containers.RepeatedCompositeFieldContainer[JobMapping]
    source: JobSource
    destinations: _containers.RepeatedCompositeFieldContainer[CreateJobDestination]
    initiate_job_run: bool
    def __init__(self, account_id: _Optional[str] = ..., job_name: _Optional[str] = ..., cron_schedule: _Optional[str] = ..., mappings: _Optional[_Iterable[_Union[JobMapping, _Mapping]]] = ..., source: _Optional[_Union[JobSource, _Mapping]] = ..., destinations: _Optional[_Iterable[_Union[CreateJobDestination, _Mapping]]] = ..., initiate_job_run: bool = ...) -> None: ...

class CreateJobResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class JobMappingTransformer(_message.Message):
    __slots__ = ("source", "config")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    source: str
    config: _transformer_pb2.TransformerConfig
    def __init__(self, source: _Optional[str] = ..., config: _Optional[_Union[_transformer_pb2.TransformerConfig, _Mapping]] = ...) -> None: ...

class JobMapping(_message.Message):
    __slots__ = ("schema", "table", "column", "transformer")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    schema: str
    table: str
    column: str
    transformer: JobMappingTransformer
    def __init__(self, schema: _Optional[str] = ..., table: _Optional[str] = ..., column: _Optional[str] = ..., transformer: _Optional[_Union[JobMappingTransformer, _Mapping]] = ...) -> None: ...

class GetJobRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class GetJobResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class UpdateJobScheduleRequest(_message.Message):
    __slots__ = ("id", "cron_schedule")
    ID_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    id: str
    cron_schedule: str
    def __init__(self, id: _Optional[str] = ..., cron_schedule: _Optional[str] = ...) -> None: ...

class UpdateJobScheduleResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class PauseJobRequest(_message.Message):
    __slots__ = ("id", "pause", "note")
    ID_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    NOTE_FIELD_NUMBER: _ClassVar[int]
    id: str
    pause: bool
    note: str
    def __init__(self, id: _Optional[str] = ..., pause: bool = ..., note: _Optional[str] = ...) -> None: ...

class PauseJobResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class UpdateJobSourceConnectionRequest(_message.Message):
    __slots__ = ("id", "source", "mappings")
    ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    id: str
    source: JobSource
    mappings: _containers.RepeatedCompositeFieldContainer[JobMapping]
    def __init__(self, id: _Optional[str] = ..., source: _Optional[_Union[JobSource, _Mapping]] = ..., mappings: _Optional[_Iterable[_Union[JobMapping, _Mapping]]] = ...) -> None: ...

class UpdateJobSourceConnectionResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class PostgresSourceSchemaSubset(_message.Message):
    __slots__ = ("postgres_schemas",)
    POSTGRES_SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    postgres_schemas: _containers.RepeatedCompositeFieldContainer[PostgresSourceSchemaOption]
    def __init__(self, postgres_schemas: _Optional[_Iterable[_Union[PostgresSourceSchemaOption, _Mapping]]] = ...) -> None: ...

class MysqlSourceSchemaSubset(_message.Message):
    __slots__ = ("mysql_schemas",)
    MYSQL_SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    mysql_schemas: _containers.RepeatedCompositeFieldContainer[MysqlSourceSchemaOption]
    def __init__(self, mysql_schemas: _Optional[_Iterable[_Union[MysqlSourceSchemaOption, _Mapping]]] = ...) -> None: ...

class JobSourceSqlSubetSchemas(_message.Message):
    __slots__ = ("postgres_subset", "mysql_subset")
    POSTGRES_SUBSET_FIELD_NUMBER: _ClassVar[int]
    MYSQL_SUBSET_FIELD_NUMBER: _ClassVar[int]
    postgres_subset: PostgresSourceSchemaSubset
    mysql_subset: MysqlSourceSchemaSubset
    def __init__(self, postgres_subset: _Optional[_Union[PostgresSourceSchemaSubset, _Mapping]] = ..., mysql_subset: _Optional[_Union[MysqlSourceSchemaSubset, _Mapping]] = ...) -> None: ...

class SetJobSourceSqlConnectionSubsetsRequest(_message.Message):
    __slots__ = ("id", "schemas")
    ID_FIELD_NUMBER: _ClassVar[int]
    SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    id: str
    schemas: JobSourceSqlSubetSchemas
    def __init__(self, id: _Optional[str] = ..., schemas: _Optional[_Union[JobSourceSqlSubetSchemas, _Mapping]] = ...) -> None: ...

class SetJobSourceSqlConnectionSubsetsResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class UpdateJobDestinationConnectionRequest(_message.Message):
    __slots__ = ("job_id", "connection_id", "options", "destination_id")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    DESTINATION_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    connection_id: str
    options: JobDestinationOptions
    destination_id: str
    def __init__(self, job_id: _Optional[str] = ..., connection_id: _Optional[str] = ..., options: _Optional[_Union[JobDestinationOptions, _Mapping]] = ..., destination_id: _Optional[str] = ...) -> None: ...

class UpdateJobDestinationConnectionResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class DeleteJobDestinationConnectionRequest(_message.Message):
    __slots__ = ("destination_id",)
    DESTINATION_ID_FIELD_NUMBER: _ClassVar[int]
    destination_id: str
    def __init__(self, destination_id: _Optional[str] = ...) -> None: ...

class DeleteJobDestinationConnectionResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CreateJobDestinationConnectionsRequest(_message.Message):
    __slots__ = ("job_id", "destinations")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    DESTINATIONS_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    destinations: _containers.RepeatedCompositeFieldContainer[CreateJobDestination]
    def __init__(self, job_id: _Optional[str] = ..., destinations: _Optional[_Iterable[_Union[CreateJobDestination, _Mapping]]] = ...) -> None: ...

class CreateJobDestinationConnectionsResponse(_message.Message):
    __slots__ = ("job",)
    JOB_FIELD_NUMBER: _ClassVar[int]
    job: Job
    def __init__(self, job: _Optional[_Union[Job, _Mapping]] = ...) -> None: ...

class DeleteJobRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class DeleteJobResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class IsJobNameAvailableRequest(_message.Message):
    __slots__ = ("name", "account_id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    account_id: str
    def __init__(self, name: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class IsJobNameAvailableResponse(_message.Message):
    __slots__ = ("is_available",)
    IS_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    is_available: bool
    def __init__(self, is_available: bool = ...) -> None: ...

class GetJobRunsRequest(_message.Message):
    __slots__ = ("job_id", "account_id")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    account_id: str
    def __init__(self, job_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class GetJobRunsResponse(_message.Message):
    __slots__ = ("job_runs",)
    JOB_RUNS_FIELD_NUMBER: _ClassVar[int]
    job_runs: _containers.RepeatedCompositeFieldContainer[JobRun]
    def __init__(self, job_runs: _Optional[_Iterable[_Union[JobRun, _Mapping]]] = ...) -> None: ...

class GetJobRunRequest(_message.Message):
    __slots__ = ("job_run_id", "account_id")
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_run_id: str
    account_id: str
    def __init__(self, job_run_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class GetJobRunResponse(_message.Message):
    __slots__ = ("job_run",)
    JOB_RUN_FIELD_NUMBER: _ClassVar[int]
    job_run: JobRun
    def __init__(self, job_run: _Optional[_Union[JobRun, _Mapping]] = ...) -> None: ...

class CreateJobRunRequest(_message.Message):
    __slots__ = ("job_id",)
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    def __init__(self, job_id: _Optional[str] = ...) -> None: ...

class CreateJobRunResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CancelJobRunRequest(_message.Message):
    __slots__ = ("job_run_id", "account_id")
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_run_id: str
    account_id: str
    def __init__(self, job_run_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class CancelJobRunResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Job(_message.Message):
    __slots__ = ("id", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "name", "source", "destinations", "mappings", "cron_schedule", "account_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    DESTINATIONS_FIELD_NUMBER: _ClassVar[int]
    MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    CRON_SCHEDULE_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    created_by_user_id: str
    created_at: _timestamp_pb2.Timestamp
    updated_by_user_id: str
    updated_at: _timestamp_pb2.Timestamp
    name: str
    source: JobSource
    destinations: _containers.RepeatedCompositeFieldContainer[JobDestination]
    mappings: _containers.RepeatedCompositeFieldContainer[JobMapping]
    cron_schedule: str
    account_id: str
    def __init__(self, id: _Optional[str] = ..., created_by_user_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_by_user_id: _Optional[str] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., name: _Optional[str] = ..., source: _Optional[_Union[JobSource, _Mapping]] = ..., destinations: _Optional[_Iterable[_Union[JobDestination, _Mapping]]] = ..., mappings: _Optional[_Iterable[_Union[JobMapping, _Mapping]]] = ..., cron_schedule: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class JobRecentRun(_message.Message):
    __slots__ = ("start_time", "job_run_id")
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    start_time: _timestamp_pb2.Timestamp
    job_run_id: str
    def __init__(self, start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., job_run_id: _Optional[str] = ...) -> None: ...

class GetJobRecentRunsRequest(_message.Message):
    __slots__ = ("job_id",)
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    def __init__(self, job_id: _Optional[str] = ...) -> None: ...

class GetJobRecentRunsResponse(_message.Message):
    __slots__ = ("recent_runs",)
    RECENT_RUNS_FIELD_NUMBER: _ClassVar[int]
    recent_runs: _containers.RepeatedCompositeFieldContainer[JobRecentRun]
    def __init__(self, recent_runs: _Optional[_Iterable[_Union[JobRecentRun, _Mapping]]] = ...) -> None: ...

class JobNextRuns(_message.Message):
    __slots__ = ("next_run_times",)
    NEXT_RUN_TIMES_FIELD_NUMBER: _ClassVar[int]
    next_run_times: _containers.RepeatedCompositeFieldContainer[_timestamp_pb2.Timestamp]
    def __init__(self, next_run_times: _Optional[_Iterable[_Union[_timestamp_pb2.Timestamp, _Mapping]]] = ...) -> None: ...

class GetJobNextRunsRequest(_message.Message):
    __slots__ = ("job_id",)
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    def __init__(self, job_id: _Optional[str] = ...) -> None: ...

class GetJobNextRunsResponse(_message.Message):
    __slots__ = ("next_runs",)
    NEXT_RUNS_FIELD_NUMBER: _ClassVar[int]
    next_runs: JobNextRuns
    def __init__(self, next_runs: _Optional[_Union[JobNextRuns, _Mapping]] = ...) -> None: ...

class GetJobStatusRequest(_message.Message):
    __slots__ = ("job_id",)
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    def __init__(self, job_id: _Optional[str] = ...) -> None: ...

class GetJobStatusResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: JobStatus
    def __init__(self, status: _Optional[_Union[JobStatus, str]] = ...) -> None: ...

class JobStatusRecord(_message.Message):
    __slots__ = ("job_id", "status")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    status: JobStatus
    def __init__(self, job_id: _Optional[str] = ..., status: _Optional[_Union[JobStatus, str]] = ...) -> None: ...

class GetJobStatusesRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetJobStatusesResponse(_message.Message):
    __slots__ = ("statuses",)
    STATUSES_FIELD_NUMBER: _ClassVar[int]
    statuses: _containers.RepeatedCompositeFieldContainer[JobStatusRecord]
    def __init__(self, statuses: _Optional[_Iterable[_Union[JobStatusRecord, _Mapping]]] = ...) -> None: ...

class ActivityFailure(_message.Message):
    __slots__ = ("message",)
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    message: str
    def __init__(self, message: _Optional[str] = ...) -> None: ...

class PendingActivity(_message.Message):
    __slots__ = ("status", "activity_name", "last_failure")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_NAME_FIELD_NUMBER: _ClassVar[int]
    LAST_FAILURE_FIELD_NUMBER: _ClassVar[int]
    status: ActivityStatus
    activity_name: str
    last_failure: ActivityFailure
    def __init__(self, status: _Optional[_Union[ActivityStatus, str]] = ..., activity_name: _Optional[str] = ..., last_failure: _Optional[_Union[ActivityFailure, _Mapping]] = ...) -> None: ...

class JobRun(_message.Message):
    __slots__ = ("id", "job_id", "name", "status", "started_at", "completed_at", "pending_activities")
    ID_FIELD_NUMBER: _ClassVar[int]
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    COMPLETED_AT_FIELD_NUMBER: _ClassVar[int]
    PENDING_ACTIVITIES_FIELD_NUMBER: _ClassVar[int]
    id: str
    job_id: str
    name: str
    status: JobRunStatus
    started_at: _timestamp_pb2.Timestamp
    completed_at: _timestamp_pb2.Timestamp
    pending_activities: _containers.RepeatedCompositeFieldContainer[PendingActivity]
    def __init__(self, id: _Optional[str] = ..., job_id: _Optional[str] = ..., name: _Optional[str] = ..., status: _Optional[_Union[JobRunStatus, str]] = ..., started_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., completed_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., pending_activities: _Optional[_Iterable[_Union[PendingActivity, _Mapping]]] = ...) -> None: ...

class JobRunEventTaskError(_message.Message):
    __slots__ = ("message", "retry_state")
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    RETRY_STATE_FIELD_NUMBER: _ClassVar[int]
    message: str
    retry_state: str
    def __init__(self, message: _Optional[str] = ..., retry_state: _Optional[str] = ...) -> None: ...

class JobRunEventTask(_message.Message):
    __slots__ = ("id", "type", "event_time", "error")
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    EVENT_TIME_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    id: int
    type: str
    event_time: _timestamp_pb2.Timestamp
    error: JobRunEventTaskError
    def __init__(self, id: _Optional[int] = ..., type: _Optional[str] = ..., event_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., error: _Optional[_Union[JobRunEventTaskError, _Mapping]] = ...) -> None: ...

class JobRunSyncMetadata(_message.Message):
    __slots__ = ("schema", "table")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    schema: str
    table: str
    def __init__(self, schema: _Optional[str] = ..., table: _Optional[str] = ...) -> None: ...

class JobRunEventMetadata(_message.Message):
    __slots__ = ("sync_metadata",)
    SYNC_METADATA_FIELD_NUMBER: _ClassVar[int]
    sync_metadata: JobRunSyncMetadata
    def __init__(self, sync_metadata: _Optional[_Union[JobRunSyncMetadata, _Mapping]] = ...) -> None: ...

class JobRunEvent(_message.Message):
    __slots__ = ("id", "type", "start_time", "close_time", "metadata", "tasks")
    ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    CLOSE_TIME_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    TASKS_FIELD_NUMBER: _ClassVar[int]
    id: int
    type: str
    start_time: _timestamp_pb2.Timestamp
    close_time: _timestamp_pb2.Timestamp
    metadata: JobRunEventMetadata
    tasks: _containers.RepeatedCompositeFieldContainer[JobRunEventTask]
    def __init__(self, id: _Optional[int] = ..., type: _Optional[str] = ..., start_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., close_time: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., metadata: _Optional[_Union[JobRunEventMetadata, _Mapping]] = ..., tasks: _Optional[_Iterable[_Union[JobRunEventTask, _Mapping]]] = ...) -> None: ...

class GetJobRunEventsRequest(_message.Message):
    __slots__ = ("job_run_id", "account_id")
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_run_id: str
    account_id: str
    def __init__(self, job_run_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class GetJobRunEventsResponse(_message.Message):
    __slots__ = ("events", "is_run_complete")
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    IS_RUN_COMPLETE_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[JobRunEvent]
    is_run_complete: bool
    def __init__(self, events: _Optional[_Iterable[_Union[JobRunEvent, _Mapping]]] = ..., is_run_complete: bool = ...) -> None: ...

class DeleteJobRunRequest(_message.Message):
    __slots__ = ("job_run_id", "account_id")
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_run_id: str
    account_id: str
    def __init__(self, job_run_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class DeleteJobRunResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class TerminateJobRunRequest(_message.Message):
    __slots__ = ("job_run_id", "account_id")
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    job_run_id: str
    account_id: str
    def __init__(self, job_run_id: _Optional[str] = ..., account_id: _Optional[str] = ...) -> None: ...

class TerminateJobRunResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
