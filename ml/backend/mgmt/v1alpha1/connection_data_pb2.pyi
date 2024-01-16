from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PostgresStreamConfig(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class MysqlStreamConfig(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AwsS3StreamConfig(_message.Message):
    __slots__ = ("job_id", "job_run_id")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    job_run_id: str
    def __init__(self, job_id: _Optional[str] = ..., job_run_id: _Optional[str] = ...) -> None: ...

class ConnectionStreamConfig(_message.Message):
    __slots__ = ("pg_config", "aws_s3_config", "mysql_config")
    PG_CONFIG_FIELD_NUMBER: _ClassVar[int]
    AWS_S3_CONFIG_FIELD_NUMBER: _ClassVar[int]
    MYSQL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    pg_config: PostgresStreamConfig
    aws_s3_config: AwsS3StreamConfig
    mysql_config: MysqlStreamConfig
    def __init__(self, pg_config: _Optional[_Union[PostgresStreamConfig, _Mapping]] = ..., aws_s3_config: _Optional[_Union[AwsS3StreamConfig, _Mapping]] = ..., mysql_config: _Optional[_Union[MysqlStreamConfig, _Mapping]] = ...) -> None: ...

class GetConnectionDataStreamRequest(_message.Message):
    __slots__ = ("connection_id", "stream_config", "schema", "table")
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    STREAM_CONFIG_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    stream_config: ConnectionStreamConfig
    schema: str
    table: str
    def __init__(self, connection_id: _Optional[str] = ..., stream_config: _Optional[_Union[ConnectionStreamConfig, _Mapping]] = ..., schema: _Optional[str] = ..., table: _Optional[str] = ...) -> None: ...

class GetConnectionDataStreamResponse(_message.Message):
    __slots__ = ("row",)
    class RowEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: bytes
        def __init__(self, key: _Optional[str] = ..., value: _Optional[bytes] = ...) -> None: ...
    ROW_FIELD_NUMBER: _ClassVar[int]
    row: _containers.ScalarMap[str, bytes]
    def __init__(self, row: _Optional[_Mapping[str, bytes]] = ...) -> None: ...

class PostgresSchemaConfig(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class MysqlSchemaConfig(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class AwsS3SchemaConfig(_message.Message):
    __slots__ = ("job_id", "job_run_id")
    JOB_ID_FIELD_NUMBER: _ClassVar[int]
    JOB_RUN_ID_FIELD_NUMBER: _ClassVar[int]
    job_id: str
    job_run_id: str
    def __init__(self, job_id: _Optional[str] = ..., job_run_id: _Optional[str] = ...) -> None: ...

class ConnectionSchemaConfig(_message.Message):
    __slots__ = ("pg_config", "aws_s3_config", "mysql_config")
    PG_CONFIG_FIELD_NUMBER: _ClassVar[int]
    AWS_S3_CONFIG_FIELD_NUMBER: _ClassVar[int]
    MYSQL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    pg_config: PostgresSchemaConfig
    aws_s3_config: AwsS3SchemaConfig
    mysql_config: MysqlSchemaConfig
    def __init__(self, pg_config: _Optional[_Union[PostgresSchemaConfig, _Mapping]] = ..., aws_s3_config: _Optional[_Union[AwsS3SchemaConfig, _Mapping]] = ..., mysql_config: _Optional[_Union[MysqlSchemaConfig, _Mapping]] = ...) -> None: ...

class DatabaseColumn(_message.Message):
    __slots__ = ("schema", "table", "column", "data_type")
    SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TABLE_FIELD_NUMBER: _ClassVar[int]
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    DATA_TYPE_FIELD_NUMBER: _ClassVar[int]
    schema: str
    table: str
    column: str
    data_type: str
    def __init__(self, schema: _Optional[str] = ..., table: _Optional[str] = ..., column: _Optional[str] = ..., data_type: _Optional[str] = ...) -> None: ...

class GetConnectionSchemaRequest(_message.Message):
    __slots__ = ("connection_id", "schema_config")
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    SCHEMA_CONFIG_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    schema_config: ConnectionSchemaConfig
    def __init__(self, connection_id: _Optional[str] = ..., schema_config: _Optional[_Union[ConnectionSchemaConfig, _Mapping]] = ...) -> None: ...

class GetConnectionSchemaResponse(_message.Message):
    __slots__ = ("schemas",)
    SCHEMAS_FIELD_NUMBER: _ClassVar[int]
    schemas: _containers.RepeatedCompositeFieldContainer[DatabaseColumn]
    def __init__(self, schemas: _Optional[_Iterable[_Union[DatabaseColumn, _Mapping]]] = ...) -> None: ...

class GetConnectionForeignConstraintsRequest(_message.Message):
    __slots__ = ("connection_id",)
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    def __init__(self, connection_id: _Optional[str] = ...) -> None: ...

class ForeignKey(_message.Message):
    __slots__ = ("table", "column")
    TABLE_FIELD_NUMBER: _ClassVar[int]
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    table: str
    column: str
    def __init__(self, table: _Optional[str] = ..., column: _Optional[str] = ...) -> None: ...

class ForeignConstraint(_message.Message):
    __slots__ = ("column", "is_nullable", "foreign_key")
    COLUMN_FIELD_NUMBER: _ClassVar[int]
    IS_NULLABLE_FIELD_NUMBER: _ClassVar[int]
    FOREIGN_KEY_FIELD_NUMBER: _ClassVar[int]
    column: str
    is_nullable: bool
    foreign_key: ForeignKey
    def __init__(self, column: _Optional[str] = ..., is_nullable: bool = ..., foreign_key: _Optional[_Union[ForeignKey, _Mapping]] = ...) -> None: ...

class ForeignConstraintTables(_message.Message):
    __slots__ = ("constraints",)
    CONSTRAINTS_FIELD_NUMBER: _ClassVar[int]
    constraints: _containers.RepeatedCompositeFieldContainer[ForeignConstraint]
    def __init__(self, constraints: _Optional[_Iterable[_Union[ForeignConstraint, _Mapping]]] = ...) -> None: ...

class GetConnectionForeignConstraintsResponse(_message.Message):
    __slots__ = ("table_constraints",)
    class TableConstraintsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: ForeignConstraintTables
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[ForeignConstraintTables, _Mapping]] = ...) -> None: ...
    TABLE_CONSTRAINTS_FIELD_NUMBER: _ClassVar[int]
    table_constraints: _containers.MessageMap[str, ForeignConstraintTables]
    def __init__(self, table_constraints: _Optional[_Mapping[str, ForeignConstraintTables]] = ...) -> None: ...

class InitStatementOptions(_message.Message):
    __slots__ = ("init_schema", "truncate_before_insert", "truncate_cascade")
    INIT_SCHEMA_FIELD_NUMBER: _ClassVar[int]
    TRUNCATE_BEFORE_INSERT_FIELD_NUMBER: _ClassVar[int]
    TRUNCATE_CASCADE_FIELD_NUMBER: _ClassVar[int]
    init_schema: bool
    truncate_before_insert: bool
    truncate_cascade: bool
    def __init__(self, init_schema: bool = ..., truncate_before_insert: bool = ..., truncate_cascade: bool = ...) -> None: ...

class GetConnectionInitStatementsRequest(_message.Message):
    __slots__ = ("connection_id", "options")
    CONNECTION_ID_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    connection_id: str
    options: InitStatementOptions
    def __init__(self, connection_id: _Optional[str] = ..., options: _Optional[_Union[InitStatementOptions, _Mapping]] = ...) -> None: ...

class GetConnectionInitStatementsResponse(_message.Message):
    __slots__ = ("table_init_statements",)
    class TableInitStatementsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    TABLE_INIT_STATEMENTS_FIELD_NUMBER: _ClassVar[int]
    table_init_statements: _containers.ScalarMap[str, str]
    def __init__(self, table_init_statements: _Optional[_Mapping[str, str]] = ...) -> None: ...
