from buf.validate import expression_pb2 as _expression_pb2
from buf.validate.priv import private_pb2 as _private_pb2
from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf import duration_pb2 as _duration_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Ignore(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    IGNORE_UNSPECIFIED: _ClassVar[Ignore]
    IGNORE_IF_UNPOPULATED: _ClassVar[Ignore]
    IGNORE_IF_DEFAULT_VALUE: _ClassVar[Ignore]
    IGNORE_ALWAYS: _ClassVar[Ignore]
    IGNORE_EMPTY: _ClassVar[Ignore]
    IGNORE_DEFAULT: _ClassVar[Ignore]

class KnownRegex(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    KNOWN_REGEX_UNSPECIFIED: _ClassVar[KnownRegex]
    KNOWN_REGEX_HTTP_HEADER_NAME: _ClassVar[KnownRegex]
    KNOWN_REGEX_HTTP_HEADER_VALUE: _ClassVar[KnownRegex]
IGNORE_UNSPECIFIED: Ignore
IGNORE_IF_UNPOPULATED: Ignore
IGNORE_IF_DEFAULT_VALUE: Ignore
IGNORE_ALWAYS: Ignore
IGNORE_EMPTY: Ignore
IGNORE_DEFAULT: Ignore
KNOWN_REGEX_UNSPECIFIED: KnownRegex
KNOWN_REGEX_HTTP_HEADER_NAME: KnownRegex
KNOWN_REGEX_HTTP_HEADER_VALUE: KnownRegex
MESSAGE_FIELD_NUMBER: _ClassVar[int]
message: _descriptor.FieldDescriptor
ONEOF_FIELD_NUMBER: _ClassVar[int]
oneof: _descriptor.FieldDescriptor
FIELD_FIELD_NUMBER: _ClassVar[int]
field: _descriptor.FieldDescriptor

class MessageConstraints(_message.Message):
    __slots__ = ("disabled", "cel")
    DISABLED_FIELD_NUMBER: _ClassVar[int]
    CEL_FIELD_NUMBER: _ClassVar[int]
    disabled: bool
    cel: _containers.RepeatedCompositeFieldContainer[_expression_pb2.Constraint]
    def __init__(self, disabled: bool = ..., cel: _Optional[_Iterable[_Union[_expression_pb2.Constraint, _Mapping]]] = ...) -> None: ...

class OneofConstraints(_message.Message):
    __slots__ = ("required",)
    REQUIRED_FIELD_NUMBER: _ClassVar[int]
    required: bool
    def __init__(self, required: bool = ...) -> None: ...

class FieldConstraints(_message.Message):
    __slots__ = ("cel", "required", "ignore", "float", "double", "int32", "int64", "uint32", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64", "bool", "string", "bytes", "enum", "repeated", "map", "any", "duration", "timestamp", "skipped", "ignore_empty")
    CEL_FIELD_NUMBER: _ClassVar[int]
    REQUIRED_FIELD_NUMBER: _ClassVar[int]
    IGNORE_FIELD_NUMBER: _ClassVar[int]
    FLOAT_FIELD_NUMBER: _ClassVar[int]
    DOUBLE_FIELD_NUMBER: _ClassVar[int]
    INT32_FIELD_NUMBER: _ClassVar[int]
    INT64_FIELD_NUMBER: _ClassVar[int]
    UINT32_FIELD_NUMBER: _ClassVar[int]
    UINT64_FIELD_NUMBER: _ClassVar[int]
    SINT32_FIELD_NUMBER: _ClassVar[int]
    SINT64_FIELD_NUMBER: _ClassVar[int]
    FIXED32_FIELD_NUMBER: _ClassVar[int]
    FIXED64_FIELD_NUMBER: _ClassVar[int]
    SFIXED32_FIELD_NUMBER: _ClassVar[int]
    SFIXED64_FIELD_NUMBER: _ClassVar[int]
    BOOL_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    BYTES_FIELD_NUMBER: _ClassVar[int]
    ENUM_FIELD_NUMBER: _ClassVar[int]
    REPEATED_FIELD_NUMBER: _ClassVar[int]
    MAP_FIELD_NUMBER: _ClassVar[int]
    ANY_FIELD_NUMBER: _ClassVar[int]
    DURATION_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    SKIPPED_FIELD_NUMBER: _ClassVar[int]
    IGNORE_EMPTY_FIELD_NUMBER: _ClassVar[int]
    cel: _containers.RepeatedCompositeFieldContainer[_expression_pb2.Constraint]
    required: bool
    ignore: Ignore
    float: FloatRules
    double: DoubleRules
    int32: Int32Rules
    int64: Int64Rules
    uint32: UInt32Rules
    uint64: UInt64Rules
    sint32: SInt32Rules
    sint64: SInt64Rules
    fixed32: Fixed32Rules
    fixed64: Fixed64Rules
    sfixed32: SFixed32Rules
    sfixed64: SFixed64Rules
    bool: BoolRules
    string: StringRules
    bytes: BytesRules
    enum: EnumRules
    repeated: RepeatedRules
    map: MapRules
    any: AnyRules
    duration: DurationRules
    timestamp: TimestampRules
    skipped: bool
    ignore_empty: bool
    def __init__(self, cel: _Optional[_Iterable[_Union[_expression_pb2.Constraint, _Mapping]]] = ..., required: bool = ..., ignore: _Optional[_Union[Ignore, str]] = ..., float: _Optional[_Union[FloatRules, _Mapping]] = ..., double: _Optional[_Union[DoubleRules, _Mapping]] = ..., int32: _Optional[_Union[Int32Rules, _Mapping]] = ..., int64: _Optional[_Union[Int64Rules, _Mapping]] = ..., uint32: _Optional[_Union[UInt32Rules, _Mapping]] = ..., uint64: _Optional[_Union[UInt64Rules, _Mapping]] = ..., sint32: _Optional[_Union[SInt32Rules, _Mapping]] = ..., sint64: _Optional[_Union[SInt64Rules, _Mapping]] = ..., fixed32: _Optional[_Union[Fixed32Rules, _Mapping]] = ..., fixed64: _Optional[_Union[Fixed64Rules, _Mapping]] = ..., sfixed32: _Optional[_Union[SFixed32Rules, _Mapping]] = ..., sfixed64: _Optional[_Union[SFixed64Rules, _Mapping]] = ..., bool: _Optional[_Union[BoolRules, _Mapping]] = ..., string: _Optional[_Union[StringRules, _Mapping]] = ..., bytes: _Optional[_Union[BytesRules, _Mapping]] = ..., enum: _Optional[_Union[EnumRules, _Mapping]] = ..., repeated: _Optional[_Union[RepeatedRules, _Mapping]] = ..., map: _Optional[_Union[MapRules, _Mapping]] = ..., any: _Optional[_Union[AnyRules, _Mapping]] = ..., duration: _Optional[_Union[DurationRules, _Mapping]] = ..., timestamp: _Optional[_Union[TimestampRules, _Mapping]] = ..., skipped: bool = ..., ignore_empty: bool = ...) -> None: ...

class FloatRules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in", "finite")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    FINITE_FIELD_NUMBER: _ClassVar[int]
    const: float
    lt: float
    lte: float
    gt: float
    gte: float
    not_in: _containers.RepeatedScalarFieldContainer[float]
    finite: bool
    def __init__(self, const: _Optional[float] = ..., lt: _Optional[float] = ..., lte: _Optional[float] = ..., gt: _Optional[float] = ..., gte: _Optional[float] = ..., not_in: _Optional[_Iterable[float]] = ..., finite: bool = ..., **kwargs) -> None: ...

class DoubleRules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in", "finite")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    FINITE_FIELD_NUMBER: _ClassVar[int]
    const: float
    lt: float
    lte: float
    gt: float
    gte: float
    not_in: _containers.RepeatedScalarFieldContainer[float]
    finite: bool
    def __init__(self, const: _Optional[float] = ..., lt: _Optional[float] = ..., lte: _Optional[float] = ..., gt: _Optional[float] = ..., gte: _Optional[float] = ..., not_in: _Optional[_Iterable[float]] = ..., finite: bool = ..., **kwargs) -> None: ...

class Int32Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class Int64Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class UInt32Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class UInt64Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class SInt32Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class SInt64Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class Fixed32Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class Fixed64Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class SFixed32Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class SFixed64Rules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    lt: int
    lte: int
    gt: int
    gte: int
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., lt: _Optional[int] = ..., lte: _Optional[int] = ..., gt: _Optional[int] = ..., gte: _Optional[int] = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class BoolRules(_message.Message):
    __slots__ = ("const",)
    CONST_FIELD_NUMBER: _ClassVar[int]
    const: bool
    def __init__(self, const: bool = ...) -> None: ...

class StringRules(_message.Message):
    __slots__ = ("const", "len", "min_len", "max_len", "len_bytes", "min_bytes", "max_bytes", "pattern", "prefix", "suffix", "contains", "not_contains", "not_in", "email", "hostname", "ip", "ipv4", "ipv6", "uri", "uri_ref", "address", "uuid", "tuuid", "ip_with_prefixlen", "ipv4_with_prefixlen", "ipv6_with_prefixlen", "ip_prefix", "ipv4_prefix", "ipv6_prefix", "host_and_port", "well_known_regex", "strict")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LEN_FIELD_NUMBER: _ClassVar[int]
    MIN_LEN_FIELD_NUMBER: _ClassVar[int]
    MAX_LEN_FIELD_NUMBER: _ClassVar[int]
    LEN_BYTES_FIELD_NUMBER: _ClassVar[int]
    MIN_BYTES_FIELD_NUMBER: _ClassVar[int]
    MAX_BYTES_FIELD_NUMBER: _ClassVar[int]
    PATTERN_FIELD_NUMBER: _ClassVar[int]
    PREFIX_FIELD_NUMBER: _ClassVar[int]
    SUFFIX_FIELD_NUMBER: _ClassVar[int]
    CONTAINS_FIELD_NUMBER: _ClassVar[int]
    NOT_CONTAINS_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    HOSTNAME_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    IPV4_FIELD_NUMBER: _ClassVar[int]
    IPV6_FIELD_NUMBER: _ClassVar[int]
    URI_FIELD_NUMBER: _ClassVar[int]
    URI_REF_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    UUID_FIELD_NUMBER: _ClassVar[int]
    TUUID_FIELD_NUMBER: _ClassVar[int]
    IP_WITH_PREFIXLEN_FIELD_NUMBER: _ClassVar[int]
    IPV4_WITH_PREFIXLEN_FIELD_NUMBER: _ClassVar[int]
    IPV6_WITH_PREFIXLEN_FIELD_NUMBER: _ClassVar[int]
    IP_PREFIX_FIELD_NUMBER: _ClassVar[int]
    IPV4_PREFIX_FIELD_NUMBER: _ClassVar[int]
    IPV6_PREFIX_FIELD_NUMBER: _ClassVar[int]
    HOST_AND_PORT_FIELD_NUMBER: _ClassVar[int]
    WELL_KNOWN_REGEX_FIELD_NUMBER: _ClassVar[int]
    STRICT_FIELD_NUMBER: _ClassVar[int]
    const: str
    len: int
    min_len: int
    max_len: int
    len_bytes: int
    min_bytes: int
    max_bytes: int
    pattern: str
    prefix: str
    suffix: str
    contains: str
    not_contains: str
    not_in: _containers.RepeatedScalarFieldContainer[str]
    email: bool
    hostname: bool
    ip: bool
    ipv4: bool
    ipv6: bool
    uri: bool
    uri_ref: bool
    address: bool
    uuid: bool
    tuuid: bool
    ip_with_prefixlen: bool
    ipv4_with_prefixlen: bool
    ipv6_with_prefixlen: bool
    ip_prefix: bool
    ipv4_prefix: bool
    ipv6_prefix: bool
    host_and_port: bool
    well_known_regex: KnownRegex
    strict: bool
    def __init__(self, const: _Optional[str] = ..., len: _Optional[int] = ..., min_len: _Optional[int] = ..., max_len: _Optional[int] = ..., len_bytes: _Optional[int] = ..., min_bytes: _Optional[int] = ..., max_bytes: _Optional[int] = ..., pattern: _Optional[str] = ..., prefix: _Optional[str] = ..., suffix: _Optional[str] = ..., contains: _Optional[str] = ..., not_contains: _Optional[str] = ..., not_in: _Optional[_Iterable[str]] = ..., email: bool = ..., hostname: bool = ..., ip: bool = ..., ipv4: bool = ..., ipv6: bool = ..., uri: bool = ..., uri_ref: bool = ..., address: bool = ..., uuid: bool = ..., tuuid: bool = ..., ip_with_prefixlen: bool = ..., ipv4_with_prefixlen: bool = ..., ipv6_with_prefixlen: bool = ..., ip_prefix: bool = ..., ipv4_prefix: bool = ..., ipv6_prefix: bool = ..., host_and_port: bool = ..., well_known_regex: _Optional[_Union[KnownRegex, str]] = ..., strict: bool = ..., **kwargs) -> None: ...

class BytesRules(_message.Message):
    __slots__ = ("const", "len", "min_len", "max_len", "pattern", "prefix", "suffix", "contains", "not_in", "ip", "ipv4", "ipv6")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LEN_FIELD_NUMBER: _ClassVar[int]
    MIN_LEN_FIELD_NUMBER: _ClassVar[int]
    MAX_LEN_FIELD_NUMBER: _ClassVar[int]
    PATTERN_FIELD_NUMBER: _ClassVar[int]
    PREFIX_FIELD_NUMBER: _ClassVar[int]
    SUFFIX_FIELD_NUMBER: _ClassVar[int]
    CONTAINS_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    IP_FIELD_NUMBER: _ClassVar[int]
    IPV4_FIELD_NUMBER: _ClassVar[int]
    IPV6_FIELD_NUMBER: _ClassVar[int]
    const: bytes
    len: int
    min_len: int
    max_len: int
    pattern: str
    prefix: bytes
    suffix: bytes
    contains: bytes
    not_in: _containers.RepeatedScalarFieldContainer[bytes]
    ip: bool
    ipv4: bool
    ipv6: bool
    def __init__(self, const: _Optional[bytes] = ..., len: _Optional[int] = ..., min_len: _Optional[int] = ..., max_len: _Optional[int] = ..., pattern: _Optional[str] = ..., prefix: _Optional[bytes] = ..., suffix: _Optional[bytes] = ..., contains: _Optional[bytes] = ..., not_in: _Optional[_Iterable[bytes]] = ..., ip: bool = ..., ipv4: bool = ..., ipv6: bool = ..., **kwargs) -> None: ...

class EnumRules(_message.Message):
    __slots__ = ("const", "defined_only", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    DEFINED_ONLY_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: int
    defined_only: bool
    not_in: _containers.RepeatedScalarFieldContainer[int]
    def __init__(self, const: _Optional[int] = ..., defined_only: bool = ..., not_in: _Optional[_Iterable[int]] = ..., **kwargs) -> None: ...

class RepeatedRules(_message.Message):
    __slots__ = ("min_items", "max_items", "unique", "items")
    MIN_ITEMS_FIELD_NUMBER: _ClassVar[int]
    MAX_ITEMS_FIELD_NUMBER: _ClassVar[int]
    UNIQUE_FIELD_NUMBER: _ClassVar[int]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    min_items: int
    max_items: int
    unique: bool
    items: FieldConstraints
    def __init__(self, min_items: _Optional[int] = ..., max_items: _Optional[int] = ..., unique: bool = ..., items: _Optional[_Union[FieldConstraints, _Mapping]] = ...) -> None: ...

class MapRules(_message.Message):
    __slots__ = ("min_pairs", "max_pairs", "keys", "values")
    MIN_PAIRS_FIELD_NUMBER: _ClassVar[int]
    MAX_PAIRS_FIELD_NUMBER: _ClassVar[int]
    KEYS_FIELD_NUMBER: _ClassVar[int]
    VALUES_FIELD_NUMBER: _ClassVar[int]
    min_pairs: int
    max_pairs: int
    keys: FieldConstraints
    values: FieldConstraints
    def __init__(self, min_pairs: _Optional[int] = ..., max_pairs: _Optional[int] = ..., keys: _Optional[_Union[FieldConstraints, _Mapping]] = ..., values: _Optional[_Union[FieldConstraints, _Mapping]] = ...) -> None: ...

class AnyRules(_message.Message):
    __slots__ = ("not_in",)
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    not_in: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, not_in: _Optional[_Iterable[str]] = ..., **kwargs) -> None: ...

class DurationRules(_message.Message):
    __slots__ = ("const", "lt", "lte", "gt", "gte", "not_in")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    IN_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    const: _duration_pb2.Duration
    lt: _duration_pb2.Duration
    lte: _duration_pb2.Duration
    gt: _duration_pb2.Duration
    gte: _duration_pb2.Duration
    not_in: _containers.RepeatedCompositeFieldContainer[_duration_pb2.Duration]
    def __init__(self, const: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., lt: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., lte: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., gt: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., gte: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ..., not_in: _Optional[_Iterable[_Union[_duration_pb2.Duration, _Mapping]]] = ..., **kwargs) -> None: ...

class TimestampRules(_message.Message):
    __slots__ = ("const", "lt", "lte", "lt_now", "gt", "gte", "gt_now", "within")
    CONST_FIELD_NUMBER: _ClassVar[int]
    LT_FIELD_NUMBER: _ClassVar[int]
    LTE_FIELD_NUMBER: _ClassVar[int]
    LT_NOW_FIELD_NUMBER: _ClassVar[int]
    GT_FIELD_NUMBER: _ClassVar[int]
    GTE_FIELD_NUMBER: _ClassVar[int]
    GT_NOW_FIELD_NUMBER: _ClassVar[int]
    WITHIN_FIELD_NUMBER: _ClassVar[int]
    const: _timestamp_pb2.Timestamp
    lt: _timestamp_pb2.Timestamp
    lte: _timestamp_pb2.Timestamp
    lt_now: bool
    gt: _timestamp_pb2.Timestamp
    gte: _timestamp_pb2.Timestamp
    gt_now: bool
    within: _duration_pb2.Duration
    def __init__(self, const: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., lt: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., lte: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., lt_now: bool = ..., gt: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., gte: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., gt_now: bool = ..., within: _Optional[_Union[_duration_pb2.Duration, _Mapping]] = ...) -> None: ...
