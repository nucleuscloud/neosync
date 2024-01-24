from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor
FIELD_FIELD_NUMBER: _ClassVar[int]
field: _descriptor.FieldDescriptor

class FieldConstraints(_message.Message):
    __slots__ = ("cel",)
    CEL_FIELD_NUMBER: _ClassVar[int]
    cel: _containers.RepeatedCompositeFieldContainer[Constraint]
    def __init__(self, cel: _Optional[_Iterable[_Union[Constraint, _Mapping]]] = ...) -> None: ...

class Constraint(_message.Message):
    __slots__ = ("id", "message", "expression")
    ID_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    id: str
    message: str
    expression: str
    def __init__(self, id: _Optional[str] = ..., message: _Optional[str] = ..., expression: _Optional[str] = ...) -> None: ...
