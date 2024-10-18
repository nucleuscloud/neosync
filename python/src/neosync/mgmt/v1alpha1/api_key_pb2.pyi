from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CreateAccountApiKeyRequest(_message.Message):
    __slots__ = ("account_id", "name", "expires_at")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    name: str
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, account_id: _Optional[str] = ..., name: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class CreateAccountApiKeyResponse(_message.Message):
    __slots__ = ("api_key",)
    API_KEY_FIELD_NUMBER: _ClassVar[int]
    api_key: AccountApiKey
    def __init__(self, api_key: _Optional[_Union[AccountApiKey, _Mapping]] = ...) -> None: ...

class AccountApiKey(_message.Message):
    __slots__ = ("id", "name", "account_id", "created_by_id", "created_at", "updated_by_id", "updated_at", "key_value", "user_id", "expires_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_BY_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    KEY_VALUE_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    account_id: str
    created_by_id: str
    created_at: _timestamp_pb2.Timestamp
    updated_by_id: str
    updated_at: _timestamp_pb2.Timestamp
    key_value: str
    user_id: str
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., account_id: _Optional[str] = ..., created_by_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_by_id: _Optional[str] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., key_value: _Optional[str] = ..., user_id: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class GetAccountApiKeysRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountApiKeysResponse(_message.Message):
    __slots__ = ("api_keys",)
    API_KEYS_FIELD_NUMBER: _ClassVar[int]
    api_keys: _containers.RepeatedCompositeFieldContainer[AccountApiKey]
    def __init__(self, api_keys: _Optional[_Iterable[_Union[AccountApiKey, _Mapping]]] = ...) -> None: ...

class GetAccountApiKeyRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class GetAccountApiKeyResponse(_message.Message):
    __slots__ = ("api_key",)
    API_KEY_FIELD_NUMBER: _ClassVar[int]
    api_key: AccountApiKey
    def __init__(self, api_key: _Optional[_Union[AccountApiKey, _Mapping]] = ...) -> None: ...

class RegenerateAccountApiKeyRequest(_message.Message):
    __slots__ = ("id", "expires_at")
    ID_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    id: str
    expires_at: _timestamp_pb2.Timestamp
    def __init__(self, id: _Optional[str] = ..., expires_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class RegenerateAccountApiKeyResponse(_message.Message):
    __slots__ = ("api_key",)
    API_KEY_FIELD_NUMBER: _ClassVar[int]
    api_key: AccountApiKey
    def __init__(self, api_key: _Optional[_Union[AccountApiKey, _Mapping]] = ...) -> None: ...

class DeleteAccountApiKeyRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class DeleteAccountApiKeyResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
