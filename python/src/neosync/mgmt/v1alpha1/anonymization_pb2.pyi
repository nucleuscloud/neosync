from buf.validate import validate_pb2 as _validate_pb2
from mgmt.v1alpha1 import transformer_pb2 as _transformer_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AnonymizeManyRequest(_message.Message):
    __slots__ = ("input_data", "transformer_mappings", "default_transformers", "halt_on_failure", "account_id")
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_TRANSFORMERS_FIELD_NUMBER: _ClassVar[int]
    HALT_ON_FAILURE_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    input_data: _containers.RepeatedScalarFieldContainer[str]
    transformer_mappings: _containers.RepeatedCompositeFieldContainer[TransformerMapping]
    default_transformers: DefaultTransformersConfig
    halt_on_failure: bool
    account_id: str
    def __init__(self, input_data: _Optional[_Iterable[str]] = ..., transformer_mappings: _Optional[_Iterable[_Union[TransformerMapping, _Mapping]]] = ..., default_transformers: _Optional[_Union[DefaultTransformersConfig, _Mapping]] = ..., halt_on_failure: bool = ..., account_id: _Optional[str] = ...) -> None: ...

class AnonymizeManyResponse(_message.Message):
    __slots__ = ("output_data", "errors")
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    ERRORS_FIELD_NUMBER: _ClassVar[int]
    output_data: _containers.RepeatedScalarFieldContainer[str]
    errors: _containers.RepeatedCompositeFieldContainer[AnonymizeManyErrors]
    def __init__(self, output_data: _Optional[_Iterable[str]] = ..., errors: _Optional[_Iterable[_Union[AnonymizeManyErrors, _Mapping]]] = ...) -> None: ...

class TransformerMapping(_message.Message):
    __slots__ = ("expression", "transformer")
    EXPRESSION_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    expression: str
    transformer: _transformer_pb2.TransformerConfig
    def __init__(self, expression: _Optional[str] = ..., transformer: _Optional[_Union[_transformer_pb2.TransformerConfig, _Mapping]] = ...) -> None: ...

class DefaultTransformersConfig(_message.Message):
    __slots__ = ("boolean", "number", "string")
    BOOLEAN_FIELD_NUMBER: _ClassVar[int]
    NUMBER_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    boolean: _transformer_pb2.TransformerConfig
    number: _transformer_pb2.TransformerConfig
    string: _transformer_pb2.TransformerConfig
    def __init__(self, boolean: _Optional[_Union[_transformer_pb2.TransformerConfig, _Mapping]] = ..., number: _Optional[_Union[_transformer_pb2.TransformerConfig, _Mapping]] = ..., string: _Optional[_Union[_transformer_pb2.TransformerConfig, _Mapping]] = ...) -> None: ...

class AnonymizeManyErrors(_message.Message):
    __slots__ = ("input_index", "error_message")
    INPUT_INDEX_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    input_index: int
    error_message: str
    def __init__(self, input_index: _Optional[int] = ..., error_message: _Optional[str] = ...) -> None: ...

class AnonymizeSingleRequest(_message.Message):
    __slots__ = ("input_data", "transformer_mappings", "default_transformers", "account_id")
    INPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_MAPPINGS_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_TRANSFORMERS_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    input_data: str
    transformer_mappings: _containers.RepeatedCompositeFieldContainer[TransformerMapping]
    default_transformers: DefaultTransformersConfig
    account_id: str
    def __init__(self, input_data: _Optional[str] = ..., transformer_mappings: _Optional[_Iterable[_Union[TransformerMapping, _Mapping]]] = ..., default_transformers: _Optional[_Union[DefaultTransformersConfig, _Mapping]] = ..., account_id: _Optional[str] = ...) -> None: ...

class AnonymizeSingleResponse(_message.Message):
    __slots__ = ("output_data",)
    OUTPUT_DATA_FIELD_NUMBER: _ClassVar[int]
    output_data: str
    def __init__(self, output_data: _Optional[str] = ...) -> None: ...
