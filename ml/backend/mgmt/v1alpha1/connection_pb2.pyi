from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetConnectionsRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetConnectionsResponse(_message.Message):
    __slots__ = ("connections",)
    CONNECTIONS_FIELD_NUMBER: _ClassVar[int]
    connections: _containers.RepeatedCompositeFieldContainer[Connection]
    def __init__(self, connections: _Optional[_Iterable[_Union[Connection, _Mapping]]] = ...) -> None: ...

class GetConnectionRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class GetConnectionResponse(_message.Message):
    __slots__ = ("connection",)
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: Connection
    def __init__(self, connection: _Optional[_Union[Connection, _Mapping]] = ...) -> None: ...

class CreateConnectionRequest(_message.Message):
    __slots__ = ("account_id", "name", "connection_config")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    name: str
    connection_config: ConnectionConfig
    def __init__(self, account_id: _Optional[str] = ..., name: _Optional[str] = ..., connection_config: _Optional[_Union[ConnectionConfig, _Mapping]] = ...) -> None: ...

class CreateConnectionResponse(_message.Message):
    __slots__ = ("connection",)
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: Connection
    def __init__(self, connection: _Optional[_Union[Connection, _Mapping]] = ...) -> None: ...

class UpdateConnectionRequest(_message.Message):
    __slots__ = ("id", "name", "connection_config")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    connection_config: ConnectionConfig
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., connection_config: _Optional[_Union[ConnectionConfig, _Mapping]] = ...) -> None: ...

class UpdateConnectionResponse(_message.Message):
    __slots__ = ("connection",)
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    connection: Connection
    def __init__(self, connection: _Optional[_Union[Connection, _Mapping]] = ...) -> None: ...

class DeleteConnectionRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class DeleteConnectionResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CheckConnectionConfigRequest(_message.Message):
    __slots__ = ("connection_config",)
    CONNECTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    connection_config: ConnectionConfig
    def __init__(self, connection_config: _Optional[_Union[ConnectionConfig, _Mapping]] = ...) -> None: ...

class CheckConnectionConfigResponse(_message.Message):
    __slots__ = ("is_connected", "connection_error")
    IS_CONNECTED_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_ERROR_FIELD_NUMBER: _ClassVar[int]
    is_connected: bool
    connection_error: str
    def __init__(self, is_connected: bool = ..., connection_error: _Optional[str] = ...) -> None: ...

class Connection(_message.Message):
    __slots__ = ("id", "name", "connection_config", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "account_id")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_CONFIG_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    connection_config: ConnectionConfig
    created_by_user_id: str
    created_at: _timestamp_pb2.Timestamp
    updated_by_user_id: str
    updated_at: _timestamp_pb2.Timestamp
    account_id: str
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., connection_config: _Optional[_Union[ConnectionConfig, _Mapping]] = ..., created_by_user_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_by_user_id: _Optional[str] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., account_id: _Optional[str] = ...) -> None: ...

class ConnectionConfig(_message.Message):
    __slots__ = ("pg_config", "aws_s3_config", "mysql_config", "local_dir_config")
    PG_CONFIG_FIELD_NUMBER: _ClassVar[int]
    AWS_S3_CONFIG_FIELD_NUMBER: _ClassVar[int]
    MYSQL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    LOCAL_DIR_CONFIG_FIELD_NUMBER: _ClassVar[int]
    pg_config: PostgresConnectionConfig
    aws_s3_config: AwsS3ConnectionConfig
    mysql_config: MysqlConnectionConfig
    local_dir_config: LocalDirectoryConnectionConfig
    def __init__(self, pg_config: _Optional[_Union[PostgresConnectionConfig, _Mapping]] = ..., aws_s3_config: _Optional[_Union[AwsS3ConnectionConfig, _Mapping]] = ..., mysql_config: _Optional[_Union[MysqlConnectionConfig, _Mapping]] = ..., local_dir_config: _Optional[_Union[LocalDirectoryConnectionConfig, _Mapping]] = ...) -> None: ...

class LocalDirectoryConnectionConfig(_message.Message):
    __slots__ = ("path",)
    PATH_FIELD_NUMBER: _ClassVar[int]
    path: str
    def __init__(self, path: _Optional[str] = ...) -> None: ...

class PostgresConnectionConfig(_message.Message):
    __slots__ = ("url", "connection")
    URL_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    url: str
    connection: PostgresConnection
    def __init__(self, url: _Optional[str] = ..., connection: _Optional[_Union[PostgresConnection, _Mapping]] = ...) -> None: ...

class PostgresConnection(_message.Message):
    __slots__ = ("host", "port", "name", "user", "ssl_mode")
    HOST_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    USER_FIELD_NUMBER: _ClassVar[int]
    PASS_FIELD_NUMBER: _ClassVar[int]
    SSL_MODE_FIELD_NUMBER: _ClassVar[int]
    host: str
    port: int
    name: str
    user: str
    ssl_mode: str
    def __init__(self, host: _Optional[str] = ..., port: _Optional[int] = ..., name: _Optional[str] = ..., user: _Optional[str] = ..., ssl_mode: _Optional[str] = ..., **kwargs) -> None: ...

class MysqlConnection(_message.Message):
    __slots__ = ("user", "protocol", "host", "port", "name")
    USER_FIELD_NUMBER: _ClassVar[int]
    PASS_FIELD_NUMBER: _ClassVar[int]
    PROTOCOL_FIELD_NUMBER: _ClassVar[int]
    HOST_FIELD_NUMBER: _ClassVar[int]
    PORT_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    user: str
    protocol: str
    host: str
    port: int
    name: str
    def __init__(self, user: _Optional[str] = ..., protocol: _Optional[str] = ..., host: _Optional[str] = ..., port: _Optional[int] = ..., name: _Optional[str] = ..., **kwargs) -> None: ...

class MysqlConnectionConfig(_message.Message):
    __slots__ = ("url", "connection")
    URL_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_FIELD_NUMBER: _ClassVar[int]
    url: str
    connection: MysqlConnection
    def __init__(self, url: _Optional[str] = ..., connection: _Optional[_Union[MysqlConnection, _Mapping]] = ...) -> None: ...

class AwsS3ConnectionConfig(_message.Message):
    __slots__ = ("bucket_arn", "path_prefix", "credentials", "region", "endpoint", "bucket")
    BUCKET_ARN_FIELD_NUMBER: _ClassVar[int]
    PATH_PREFIX_FIELD_NUMBER: _ClassVar[int]
    CREDENTIALS_FIELD_NUMBER: _ClassVar[int]
    REGION_FIELD_NUMBER: _ClassVar[int]
    ENDPOINT_FIELD_NUMBER: _ClassVar[int]
    BUCKET_FIELD_NUMBER: _ClassVar[int]
    bucket_arn: str
    path_prefix: str
    credentials: AwsS3Credentials
    region: str
    endpoint: str
    bucket: str
    def __init__(self, bucket_arn: _Optional[str] = ..., path_prefix: _Optional[str] = ..., credentials: _Optional[_Union[AwsS3Credentials, _Mapping]] = ..., region: _Optional[str] = ..., endpoint: _Optional[str] = ..., bucket: _Optional[str] = ...) -> None: ...

class AwsS3Credentials(_message.Message):
    __slots__ = ("profile", "access_key_id", "secret_access_key", "session_token", "from_ec2_role", "role_arn", "role_external_id")
    PROFILE_FIELD_NUMBER: _ClassVar[int]
    ACCESS_KEY_ID_FIELD_NUMBER: _ClassVar[int]
    SECRET_ACCESS_KEY_FIELD_NUMBER: _ClassVar[int]
    SESSION_TOKEN_FIELD_NUMBER: _ClassVar[int]
    FROM_EC2_ROLE_FIELD_NUMBER: _ClassVar[int]
    ROLE_ARN_FIELD_NUMBER: _ClassVar[int]
    ROLE_EXTERNAL_ID_FIELD_NUMBER: _ClassVar[int]
    profile: str
    access_key_id: str
    secret_access_key: str
    session_token: str
    from_ec2_role: bool
    role_arn: str
    role_external_id: str
    def __init__(self, profile: _Optional[str] = ..., access_key_id: _Optional[str] = ..., secret_access_key: _Optional[str] = ..., session_token: _Optional[str] = ..., from_ec2_role: bool = ..., role_arn: _Optional[str] = ..., role_external_id: _Optional[str] = ...) -> None: ...

class IsConnectionNameAvailableRequest(_message.Message):
    __slots__ = ("account_id", "connection_name")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CONNECTION_NAME_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    connection_name: str
    def __init__(self, account_id: _Optional[str] = ..., connection_name: _Optional[str] = ...) -> None: ...

class IsConnectionNameAvailableResponse(_message.Message):
    __slots__ = ("is_available",)
    IS_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    is_available: bool
    def __init__(self, is_available: bool = ...) -> None: ...

class CheckSqlQueryRequest(_message.Message):
    __slots__ = ("id", "query")
    ID_FIELD_NUMBER: _ClassVar[int]
    QUERY_FIELD_NUMBER: _ClassVar[int]
    id: str
    query: str
    def __init__(self, id: _Optional[str] = ..., query: _Optional[str] = ...) -> None: ...

class CheckSqlQueryResponse(_message.Message):
    __slots__ = ("is_valid", "erorr_message")
    IS_VALID_FIELD_NUMBER: _ClassVar[int]
    ERORR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    is_valid: bool
    erorr_message: str
    def __init__(self, is_valid: bool = ..., erorr_message: _Optional[str] = ...) -> None: ...
