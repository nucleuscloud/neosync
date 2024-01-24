from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Constraint(_message.Message):
    __slots__ = ("id", "message", "expression")
    ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    id: str
    message: str
    expression: str
    def __init__(self, id: _Optional[str] = ..., message: _Optional[str] = ..., expression: _Optional[str] = ...) -> None: ...

class Violations(_message.Message):
    __slots__ = ("violations",)
    VIOLATIONS_FIELD_NUMBER: _ClassVar[int]
    violations: _containers.RepeatedCompositeFieldContainer[Violation]
    def __init__(self, violations: _Optional[_Iterable[_Union[Violation, _Mapping]]] = ...) -> None: ...

class Violation(_message.Message):
    __slots__ = ("field_path", "constraint_id", "message", "for_key")
    FIELD_PATH_FIELD_NUMBER: _ClassVar[int]
    CONSTRAINT_ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    FOR_KEY_FIELD_NUMBER: _ClassVar[int]
    field_path: str
    constraint_id: str
    message: str
    for_key: bool
    def __init__(self, field_path: _Optional[str] = ..., constraint_id: _Optional[str] = ..., message: _Optional[str] = ..., for_key: bool = ...) -> None: ...
