from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AccountHookEvent(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACCOUNT_HOOK_EVENT_UNSPECIFIED: _ClassVar[AccountHookEvent]
    ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED: _ClassVar[AccountHookEvent]
    ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED: _ClassVar[AccountHookEvent]
    ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED: _ClassVar[AccountHookEvent]
ACCOUNT_HOOK_EVENT_UNSPECIFIED: AccountHookEvent
ACCOUNT_HOOK_EVENT_JOB_RUN_CREATED: AccountHookEvent
ACCOUNT_HOOK_EVENT_JOB_RUN_FAILED: AccountHookEvent
ACCOUNT_HOOK_EVENT_JOB_RUN_SUCCEEDED: AccountHookEvent

class AccountHook(_message.Message):
    __slots__ = ("id", "name", "description", "account_id", "events", "config", "created_by_user_id", "created_at", "updated_by_user_id", "updated_at", "enabled")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    CREATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_BY_USER_ID_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    description: str
    account_id: str
    events: _containers.RepeatedScalarFieldContainer[AccountHookEvent]
    config: AccountHookConfig
    created_by_user_id: str
    created_at: _timestamp_pb2.Timestamp
    updated_by_user_id: str
    updated_at: _timestamp_pb2.Timestamp
    enabled: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., account_id: _Optional[str] = ..., events: _Optional[_Iterable[_Union[AccountHookEvent, str]]] = ..., config: _Optional[_Union[AccountHookConfig, _Mapping]] = ..., created_by_user_id: _Optional[str] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_by_user_id: _Optional[str] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., enabled: bool = ...) -> None: ...

class NewAccountHook(_message.Message):
    __slots__ = ("name", "description", "events", "config", "enabled")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    events: _containers.RepeatedScalarFieldContainer[AccountHookEvent]
    config: AccountHookConfig
    enabled: bool
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., events: _Optional[_Iterable[_Union[AccountHookEvent, str]]] = ..., config: _Optional[_Union[AccountHookConfig, _Mapping]] = ..., enabled: bool = ...) -> None: ...

class AccountHookConfig(_message.Message):
    __slots__ = ("webhook", "slack")
    class WebHook(_message.Message):
        __slots__ = ("url", "secret", "disable_ssl_verification")
        URL_FIELD_NUMBER: _ClassVar[int]
        SECRET_FIELD_NUMBER: _ClassVar[int]
        DISABLE_SSL_VERIFICATION_FIELD_NUMBER: _ClassVar[int]
        url: str
        secret: str
        disable_ssl_verification: bool
        def __init__(self, url: _Optional[str] = ..., secret: _Optional[str] = ..., disable_ssl_verification: bool = ...) -> None: ...
    class SlackHook(_message.Message):
        __slots__ = ("channel",)
        CHANNEL_FIELD_NUMBER: _ClassVar[int]
        channel: str
        def __init__(self, channel: _Optional[str] = ...) -> None: ...
    WEBHOOK_FIELD_NUMBER: _ClassVar[int]
    SLACK_FIELD_NUMBER: _ClassVar[int]
    webhook: AccountHookConfig.WebHook
    slack: AccountHookConfig.SlackHook
    def __init__(self, webhook: _Optional[_Union[AccountHookConfig.WebHook, _Mapping]] = ..., slack: _Optional[_Union[AccountHookConfig.SlackHook, _Mapping]] = ...) -> None: ...

class GetAccountHooksRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetAccountHooksResponse(_message.Message):
    __slots__ = ("hooks",)
    HOOKS_FIELD_NUMBER: _ClassVar[int]
    hooks: _containers.RepeatedCompositeFieldContainer[AccountHook]
    def __init__(self, hooks: _Optional[_Iterable[_Union[AccountHook, _Mapping]]] = ...) -> None: ...

class GetAccountHookRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class GetAccountHookResponse(_message.Message):
    __slots__ = ("hook",)
    HOOK_FIELD_NUMBER: _ClassVar[int]
    hook: AccountHook
    def __init__(self, hook: _Optional[_Union[AccountHook, _Mapping]] = ...) -> None: ...

class CreateAccountHookRequest(_message.Message):
    __slots__ = ("account_id", "hook")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    HOOK_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    hook: NewAccountHook
    def __init__(self, account_id: _Optional[str] = ..., hook: _Optional[_Union[NewAccountHook, _Mapping]] = ...) -> None: ...

class CreateAccountHookResponse(_message.Message):
    __slots__ = ("hook",)
    HOOK_FIELD_NUMBER: _ClassVar[int]
    hook: AccountHook
    def __init__(self, hook: _Optional[_Union[AccountHook, _Mapping]] = ...) -> None: ...

class UpdateAccountHookRequest(_message.Message):
    __slots__ = ("id", "name", "description", "events", "config", "enabled")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    description: str
    events: _containers.RepeatedScalarFieldContainer[AccountHookEvent]
    config: AccountHookConfig
    enabled: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., events: _Optional[_Iterable[_Union[AccountHookEvent, str]]] = ..., config: _Optional[_Union[AccountHookConfig, _Mapping]] = ..., enabled: bool = ...) -> None: ...

class UpdateAccountHookResponse(_message.Message):
    __slots__ = ("hook",)
    HOOK_FIELD_NUMBER: _ClassVar[int]
    hook: AccountHook
    def __init__(self, hook: _Optional[_Union[AccountHook, _Mapping]] = ...) -> None: ...

class DeleteAccountHookRequest(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class DeleteAccountHookResponse(_message.Message):
    __slots__ = ("hook",)
    HOOK_FIELD_NUMBER: _ClassVar[int]
    hook: AccountHook
    def __init__(self, hook: _Optional[_Union[AccountHook, _Mapping]] = ...) -> None: ...

class IsAccountHookNameAvailableRequest(_message.Message):
    __slots__ = ("account_id", "name")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    name: str
    def __init__(self, account_id: _Optional[str] = ..., name: _Optional[str] = ...) -> None: ...

class IsAccountHookNameAvailableResponse(_message.Message):
    __slots__ = ("is_available",)
    IS_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    is_available: bool
    def __init__(self, is_available: bool = ...) -> None: ...

class SetAccountHookEnabledRequest(_message.Message):
    __slots__ = ("id", "enabled")
    ID_FIELD_NUMBER: _ClassVar[int]
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    id: str
    enabled: bool
    def __init__(self, id: _Optional[str] = ..., enabled: bool = ...) -> None: ...

class SetAccountHookEnabledResponse(_message.Message):
    __slots__ = ("hook",)
    HOOK_FIELD_NUMBER: _ClassVar[int]
    hook: AccountHook
    def __init__(self, hook: _Optional[_Union[AccountHook, _Mapping]] = ...) -> None: ...

class GetActiveAccountHooksByEventRequest(_message.Message):
    __slots__ = ("account_id", "event")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    event: AccountHookEvent
    def __init__(self, account_id: _Optional[str] = ..., event: _Optional[_Union[AccountHookEvent, str]] = ...) -> None: ...

class GetActiveAccountHooksByEventResponse(_message.Message):
    __slots__ = ("hooks",)
    HOOKS_FIELD_NUMBER: _ClassVar[int]
    hooks: _containers.RepeatedCompositeFieldContainer[AccountHook]
    def __init__(self, hooks: _Optional[_Iterable[_Union[AccountHook, _Mapping]]] = ...) -> None: ...

class GetSlackConnectionUrlRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetSlackConnectionUrlResponse(_message.Message):
    __slots__ = ("url",)
    URL_FIELD_NUMBER: _ClassVar[int]
    url: str
    def __init__(self, url: _Optional[str] = ...) -> None: ...

class HandleSlackOAuthCallbackRequest(_message.Message):
    __slots__ = ("account_id", "state", "code")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    state: str
    code: str
    def __init__(self, account_id: _Optional[str] = ..., state: _Optional[str] = ..., code: _Optional[str] = ...) -> None: ...

class HandleSlackOAuthCallbackResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class TestSlackConnectionRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class TestSlackConnectionResponse(_message.Message):
    __slots__ = ("has_configuration", "test_response", "error")
    class Response(_message.Message):
        __slots__ = ("url", "team")
        URL_FIELD_NUMBER: _ClassVar[int]
        TEAM_FIELD_NUMBER: _ClassVar[int]
        url: str
        team: str
        def __init__(self, url: _Optional[str] = ..., team: _Optional[str] = ...) -> None: ...
    HAS_CONFIGURATION_FIELD_NUMBER: _ClassVar[int]
    TEST_RESPONSE_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    has_configuration: bool
    test_response: TestSlackConnectionResponse.Response
    error: str
    def __init__(self, has_configuration: bool = ..., test_response: _Optional[_Union[TestSlackConnectionResponse.Response, _Mapping]] = ..., error: _Optional[str] = ...) -> None: ...

class SendSlackMessageRequest(_message.Message):
    __slots__ = ("account_hook_id", "event")
    ACCOUNT_HOOK_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    account_hook_id: str
    event: bytes
    def __init__(self, account_hook_id: _Optional[str] = ..., event: _Optional[bytes] = ...) -> None: ...

class SendSlackMessageResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
