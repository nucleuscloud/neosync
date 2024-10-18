from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class LoginCliRequest(_message.Message):
    __slots__ = ("code", "redirect_uri")
    CODE_FIELD_NUMBER: _ClassVar[int]
    REDIRECT_URI_FIELD_NUMBER: _ClassVar[int]
    code: str
    redirect_uri: str
    def __init__(self, code: _Optional[str] = ..., redirect_uri: _Optional[str] = ...) -> None: ...

class LoginCliResponse(_message.Message):
    __slots__ = ("access_token",)
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    access_token: AccessToken
    def __init__(self, access_token: _Optional[_Union[AccessToken, _Mapping]] = ...) -> None: ...

class GetAuthStatusRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetAuthStatusResponse(_message.Message):
    __slots__ = ("is_enabled",)
    IS_ENABLED_FIELD_NUMBER: _ClassVar[int]
    is_enabled: bool
    def __init__(self, is_enabled: bool = ...) -> None: ...

class AccessToken(_message.Message):
    __slots__ = ("access_token", "refresh_token", "expires_in", "scope", "id_token", "token_type")
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_IN_FIELD_NUMBER: _ClassVar[int]
    SCOPE_FIELD_NUMBER: _ClassVar[int]
    ID_TOKEN_FIELD_NUMBER: _ClassVar[int]
    TOKEN_TYPE_FIELD_NUMBER: _ClassVar[int]
    access_token: str
    refresh_token: str
    expires_in: int
    scope: str
    id_token: str
    token_type: str
    def __init__(self, access_token: _Optional[str] = ..., refresh_token: _Optional[str] = ..., expires_in: _Optional[int] = ..., scope: _Optional[str] = ..., id_token: _Optional[str] = ..., token_type: _Optional[str] = ...) -> None: ...

class GetAuthorizeUrlRequest(_message.Message):
    __slots__ = ("state", "redirect_uri", "scope")
    STATE_FIELD_NUMBER: _ClassVar[int]
    REDIRECT_URI_FIELD_NUMBER: _ClassVar[int]
    SCOPE_FIELD_NUMBER: _ClassVar[int]
    state: str
    redirect_uri: str
    scope: str
    def __init__(self, state: _Optional[str] = ..., redirect_uri: _Optional[str] = ..., scope: _Optional[str] = ...) -> None: ...

class GetAuthorizeUrlResponse(_message.Message):
    __slots__ = ("url",)
    URL_FIELD_NUMBER: _ClassVar[int]
    url: str
    def __init__(self, url: _Optional[str] = ...) -> None: ...

class GetCliIssuerRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetCliIssuerResponse(_message.Message):
    __slots__ = ("issuer_url", "audience")
    ISSUER_URL_FIELD_NUMBER: _ClassVar[int]
    AUDIENCE_FIELD_NUMBER: _ClassVar[int]
    issuer_url: str
    audience: str
    def __init__(self, issuer_url: _Optional[str] = ..., audience: _Optional[str] = ...) -> None: ...

class RefreshCliRequest(_message.Message):
    __slots__ = ("refresh_token",)
    REFRESH_TOKEN_FIELD_NUMBER: _ClassVar[int]
    refresh_token: str
    def __init__(self, refresh_token: _Optional[str] = ...) -> None: ...

class RefreshCliResponse(_message.Message):
    __slots__ = ("access_token",)
    ACCESS_TOKEN_FIELD_NUMBER: _ClassVar[int]
    access_token: AccessToken
    def __init__(self, access_token: _Optional[_Union[AccessToken, _Mapping]] = ...) -> None: ...

class CheckTokenRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class CheckTokenResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
