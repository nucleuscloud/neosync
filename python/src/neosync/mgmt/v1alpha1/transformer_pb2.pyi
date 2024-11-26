from buf.validate import validate_pb2 as _validate_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TransformerSource(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRANSFORMER_SOURCE_UNSPECIFIED: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_PASSTHROUGH: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_DEFAULT: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_EMAIL: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_EMAIL: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_BOOL: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_CITY: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_FIRST_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_FLOAT64: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_FULL_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_GENDER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_INT64: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_RANDOM_INT64: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_LAST_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_SHA256HASH: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_SSN: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_STATE: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_STRING: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_USERNAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_UUID: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_ZIPCODE: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_FLOAT64: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_INT64: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_STRING: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_NULL: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_CATEGORICAL: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_USER_DEFINED: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_COUNTRY: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_TRANSFORM_PII_TEXT: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_BUSINESS_NAME: _ClassVar[TransformerSource]
    TRANSFORMER_SOURCE_GENERATE_IP_ADDRESS: _ClassVar[TransformerSource]

class TransformerDataType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRANSFORMER_DATA_TYPE_UNSPECIFIED: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_STRING: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_INT64: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_BOOLEAN: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_FLOAT64: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_NULL: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_ANY: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_TIME: _ClassVar[TransformerDataType]
    TRANSFORMER_DATA_TYPE_UUID: _ClassVar[TransformerDataType]

class SupportedJobType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SUPPORTED_JOB_TYPE_UNSPECIFIED: _ClassVar[SupportedJobType]
    SUPPORTED_JOB_TYPE_SYNC: _ClassVar[SupportedJobType]
    SUPPORTED_JOB_TYPE_GENERATE: _ClassVar[SupportedJobType]

class GenerateEmailType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    GENERATE_EMAIL_TYPE_UNSPECIFIED: _ClassVar[GenerateEmailType]
    GENERATE_EMAIL_TYPE_UUID_V4: _ClassVar[GenerateEmailType]
    GENERATE_EMAIL_TYPE_FULLNAME: _ClassVar[GenerateEmailType]

class InvalidEmailAction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    INVALID_EMAIL_ACTION_UNSPECIFIED: _ClassVar[InvalidEmailAction]
    INVALID_EMAIL_ACTION_REJECT: _ClassVar[InvalidEmailAction]
    INVALID_EMAIL_ACTION_NULL: _ClassVar[InvalidEmailAction]
    INVALID_EMAIL_ACTION_PASSTHROUGH: _ClassVar[InvalidEmailAction]
    INVALID_EMAIL_ACTION_GENERATE: _ClassVar[InvalidEmailAction]

class GenerateIpAddressType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    GENERATE_IP_ADDRESS_TYPE_UNSPECIFIED: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_A: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_B: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_C: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_LINK_LOCAL: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_MULTICAST: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V4_LOOPBACK: _ClassVar[GenerateIpAddressType]
    GENERATE_IP_ADDRESS_TYPE_V6: _ClassVar[GenerateIpAddressType]
TRANSFORMER_SOURCE_UNSPECIFIED: TransformerSource
TRANSFORMER_SOURCE_PASSTHROUGH: TransformerSource
TRANSFORMER_SOURCE_GENERATE_DEFAULT: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT: TransformerSource
TRANSFORMER_SOURCE_GENERATE_EMAIL: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_EMAIL: TransformerSource
TRANSFORMER_SOURCE_GENERATE_BOOL: TransformerSource
TRANSFORMER_SOURCE_GENERATE_CARD_NUMBER: TransformerSource
TRANSFORMER_SOURCE_GENERATE_CITY: TransformerSource
TRANSFORMER_SOURCE_GENERATE_E164_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_GENERATE_FIRST_NAME: TransformerSource
TRANSFORMER_SOURCE_GENERATE_FLOAT64: TransformerSource
TRANSFORMER_SOURCE_GENERATE_FULL_ADDRESS: TransformerSource
TRANSFORMER_SOURCE_GENERATE_FULL_NAME: TransformerSource
TRANSFORMER_SOURCE_GENERATE_GENDER: TransformerSource
TRANSFORMER_SOURCE_GENERATE_INT64_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_GENERATE_INT64: TransformerSource
TRANSFORMER_SOURCE_GENERATE_RANDOM_INT64: TransformerSource
TRANSFORMER_SOURCE_GENERATE_LAST_NAME: TransformerSource
TRANSFORMER_SOURCE_GENERATE_SHA256HASH: TransformerSource
TRANSFORMER_SOURCE_GENERATE_SSN: TransformerSource
TRANSFORMER_SOURCE_GENERATE_STATE: TransformerSource
TRANSFORMER_SOURCE_GENERATE_STREET_ADDRESS: TransformerSource
TRANSFORMER_SOURCE_GENERATE_STRING_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_GENERATE_STRING: TransformerSource
TRANSFORMER_SOURCE_GENERATE_RANDOM_STRING: TransformerSource
TRANSFORMER_SOURCE_GENERATE_UNIXTIMESTAMP: TransformerSource
TRANSFORMER_SOURCE_GENERATE_USERNAME: TransformerSource
TRANSFORMER_SOURCE_GENERATE_UTCTIMESTAMP: TransformerSource
TRANSFORMER_SOURCE_GENERATE_UUID: TransformerSource
TRANSFORMER_SOURCE_GENERATE_ZIPCODE: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_E164_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_FIRST_NAME: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_FLOAT64: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_FULL_NAME: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_INT64_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_INT64: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_LAST_NAME: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_PHONE_NUMBER: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_STRING: TransformerSource
TRANSFORMER_SOURCE_GENERATE_NULL: TransformerSource
TRANSFORMER_SOURCE_GENERATE_CATEGORICAL: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_CHARACTER_SCRAMBLE: TransformerSource
TRANSFORMER_SOURCE_USER_DEFINED: TransformerSource
TRANSFORMER_SOURCE_GENERATE_JAVASCRIPT: TransformerSource
TRANSFORMER_SOURCE_GENERATE_COUNTRY: TransformerSource
TRANSFORMER_SOURCE_TRANSFORM_PII_TEXT: TransformerSource
TRANSFORMER_SOURCE_GENERATE_BUSINESS_NAME: TransformerSource
TRANSFORMER_SOURCE_GENERATE_IP_ADDRESS: TransformerSource
TRANSFORMER_DATA_TYPE_UNSPECIFIED: TransformerDataType
TRANSFORMER_DATA_TYPE_STRING: TransformerDataType
TRANSFORMER_DATA_TYPE_INT64: TransformerDataType
TRANSFORMER_DATA_TYPE_BOOLEAN: TransformerDataType
TRANSFORMER_DATA_TYPE_FLOAT64: TransformerDataType
TRANSFORMER_DATA_TYPE_NULL: TransformerDataType
TRANSFORMER_DATA_TYPE_ANY: TransformerDataType
TRANSFORMER_DATA_TYPE_TIME: TransformerDataType
TRANSFORMER_DATA_TYPE_UUID: TransformerDataType
SUPPORTED_JOB_TYPE_UNSPECIFIED: SupportedJobType
SUPPORTED_JOB_TYPE_SYNC: SupportedJobType
SUPPORTED_JOB_TYPE_GENERATE: SupportedJobType
GENERATE_EMAIL_TYPE_UNSPECIFIED: GenerateEmailType
GENERATE_EMAIL_TYPE_UUID_V4: GenerateEmailType
GENERATE_EMAIL_TYPE_FULLNAME: GenerateEmailType
INVALID_EMAIL_ACTION_UNSPECIFIED: InvalidEmailAction
INVALID_EMAIL_ACTION_REJECT: InvalidEmailAction
INVALID_EMAIL_ACTION_NULL: InvalidEmailAction
INVALID_EMAIL_ACTION_PASSTHROUGH: InvalidEmailAction
INVALID_EMAIL_ACTION_GENERATE: InvalidEmailAction
GENERATE_IP_ADDRESS_TYPE_UNSPECIFIED: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_PUBLIC: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_A: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_B: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_PRIVATE_C: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_LINK_LOCAL: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_MULTICAST: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V4_LOOPBACK: GenerateIpAddressType
GENERATE_IP_ADDRESS_TYPE_V6: GenerateIpAddressType

class GetSystemTransformersRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetSystemTransformersResponse(_message.Message):
    __slots__ = ("transformers",)
    TRANSFORMERS_FIELD_NUMBER: _ClassVar[int]
    transformers: _containers.RepeatedCompositeFieldContainer[SystemTransformer]
    def __init__(self, transformers: _Optional[_Iterable[_Union[SystemTransformer, _Mapping]]] = ...) -> None: ...

class GetSystemTransformerBySourceRequest(_message.Message):
    __slots__ = ("source",)
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    source: TransformerSource
    def __init__(self, source: _Optional[_Union[TransformerSource, str]] = ...) -> None: ...

class GetSystemTransformerBySourceResponse(_message.Message):
    __slots__ = ("transformer",)
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    transformer: SystemTransformer
    def __init__(self, transformer: _Optional[_Union[SystemTransformer, _Mapping]] = ...) -> None: ...

class GetUserDefinedTransformersRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetUserDefinedTransformersResponse(_message.Message):
    __slots__ = ("transformers",)
    TRANSFORMERS_FIELD_NUMBER: _ClassVar[int]
    transformers: _containers.RepeatedCompositeFieldContainer[UserDefinedTransformer]
    def __init__(self, transformers: _Optional[_Iterable[_Union[UserDefinedTransformer, _Mapping]]] = ...) -> None: ...

class GetUserDefinedTransformerByIdRequest(_message.Message):
    __slots__ = ("transformer_id",)
    TRANSFORMER_ID_FIELD_NUMBER: _ClassVar[int]
    transformer_id: str
    def __init__(self, transformer_id: _Optional[str] = ...) -> None: ...

class GetUserDefinedTransformerByIdResponse(_message.Message):
    __slots__ = ("transformer",)
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    transformer: UserDefinedTransformer
    def __init__(self, transformer: _Optional[_Union[UserDefinedTransformer, _Mapping]] = ...) -> None: ...

class CreateUserDefinedTransformerRequest(_message.Message):
    __slots__ = ("account_id", "name", "description", "type", "source", "transformer_config")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    name: str
    description: str
    type: str
    source: TransformerSource
    transformer_config: TransformerConfig
    def __init__(self, account_id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., type: _Optional[str] = ..., source: _Optional[_Union[TransformerSource, str]] = ..., transformer_config: _Optional[_Union[TransformerConfig, _Mapping]] = ...) -> None: ...

class CreateUserDefinedTransformerResponse(_message.Message):
    __slots__ = ("transformer",)
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    transformer: UserDefinedTransformer
    def __init__(self, transformer: _Optional[_Union[UserDefinedTransformer, _Mapping]] = ...) -> None: ...

class DeleteUserDefinedTransformerRequest(_message.Message):
    __slots__ = ("transformer_id",)
    TRANSFORMER_ID_FIELD_NUMBER: _ClassVar[int]
    transformer_id: str
    def __init__(self, transformer_id: _Optional[str] = ...) -> None: ...

class DeleteUserDefinedTransformerResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class UpdateUserDefinedTransformerRequest(_message.Message):
    __slots__ = ("transformer_id", "name", "description", "transformer_config")
    TRANSFORMER_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    transformer_id: str
    name: str
    description: str
    transformer_config: TransformerConfig
    def __init__(self, transformer_id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., transformer_config: _Optional[_Union[TransformerConfig, _Mapping]] = ...) -> None: ...

class UpdateUserDefinedTransformerResponse(_message.Message):
    __slots__ = ("transformer",)
    TRANSFORMER_FIELD_NUMBER: _ClassVar[int]
    transformer: UserDefinedTransformer
    def __init__(self, transformer: _Optional[_Union[UserDefinedTransformer, _Mapping]] = ...) -> None: ...

class IsTransformerNameAvailableRequest(_message.Message):
    __slots__ = ("account_id", "transformer_name")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    TRANSFORMER_NAME_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    transformer_name: str
    def __init__(self, account_id: _Optional[str] = ..., transformer_name: _Optional[str] = ...) -> None: ...

class IsTransformerNameAvailableResponse(_message.Message):
    __slots__ = ("is_available",)
    IS_AVAILABLE_FIELD_NUMBER: _ClassVar[int]
    is_available: bool
    def __init__(self, is_available: bool = ...) -> None: ...

class UserDefinedTransformer(_message.Message):
    __slots__ = ("id", "name", "description", "data_type", "source", "config", "created_at", "updated_at", "account_id", "data_types")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    DATA_TYPE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    DATA_TYPES_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    description: str
    data_type: TransformerDataType
    source: TransformerSource
    config: TransformerConfig
    created_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    account_id: str
    data_types: _containers.RepeatedScalarFieldContainer[TransformerDataType]
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., data_type: _Optional[_Union[TransformerDataType, str]] = ..., source: _Optional[_Union[TransformerSource, str]] = ..., config: _Optional[_Union[TransformerConfig, _Mapping]] = ..., created_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., account_id: _Optional[str] = ..., data_types: _Optional[_Iterable[_Union[TransformerDataType, str]]] = ...) -> None: ...

class SystemTransformer(_message.Message):
    __slots__ = ("name", "description", "data_type", "source", "config", "data_types", "supported_job_types")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    DATA_TYPE_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    CONFIG_FIELD_NUMBER: _ClassVar[int]
    DATA_TYPES_FIELD_NUMBER: _ClassVar[int]
    SUPPORTED_JOB_TYPES_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    data_type: TransformerDataType
    source: TransformerSource
    config: TransformerConfig
    data_types: _containers.RepeatedScalarFieldContainer[TransformerDataType]
    supported_job_types: _containers.RepeatedScalarFieldContainer[SupportedJobType]
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., data_type: _Optional[_Union[TransformerDataType, str]] = ..., source: _Optional[_Union[TransformerSource, str]] = ..., config: _Optional[_Union[TransformerConfig, _Mapping]] = ..., data_types: _Optional[_Iterable[_Union[TransformerDataType, str]]] = ..., supported_job_types: _Optional[_Iterable[_Union[SupportedJobType, str]]] = ...) -> None: ...

class TransformerConfig(_message.Message):
    __slots__ = ("generate_email_config", "transform_email_config", "generate_bool_config", "generate_card_number_config", "generate_city_config", "generate_e164_phone_number_config", "generate_first_name_config", "generate_float64_config", "generate_full_address_config", "generate_full_name_config", "generate_gender_config", "generate_int64_phone_number_config", "generate_int64_config", "generate_last_name_config", "generate_sha256hash_config", "generate_ssn_config", "generate_state_config", "generate_street_address_config", "generate_string_phone_number_config", "generate_string_config", "generate_unixtimestamp_config", "generate_username_config", "generate_utctimestamp_config", "generate_uuid_config", "generate_zipcode_config", "transform_e164_phone_number_config", "transform_first_name_config", "transform_float64_config", "transform_full_name_config", "transform_int64_phone_number_config", "transform_int64_config", "transform_last_name_config", "transform_phone_number_config", "transform_string_config", "passthrough_config", "nullconfig", "user_defined_transformer_config", "generate_default_config", "transform_javascript_config", "generate_categorical_config", "transform_character_scramble_config", "generate_javascript_config", "generate_country_config", "transform_pii_text_config", "generate_business_name_config", "generate_ip_address_config")
    GENERATE_EMAIL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_EMAIL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_BOOL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_CARD_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_CITY_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_E164_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_FIRST_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_FLOAT64_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_FULL_ADDRESS_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_FULL_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_GENDER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_INT64_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_INT64_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_LAST_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_SHA256HASH_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_SSN_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_STATE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_STREET_ADDRESS_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_STRING_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_STRING_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_UNIXTIMESTAMP_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_USERNAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_UTCTIMESTAMP_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_UUID_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_ZIPCODE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_E164_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_FIRST_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_FLOAT64_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_FULL_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_INT64_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_INT64_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_LAST_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_PHONE_NUMBER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_STRING_CONFIG_FIELD_NUMBER: _ClassVar[int]
    PASSTHROUGH_CONFIG_FIELD_NUMBER: _ClassVar[int]
    NULLCONFIG_FIELD_NUMBER: _ClassVar[int]
    USER_DEFINED_TRANSFORMER_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_DEFAULT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_JAVASCRIPT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_CATEGORICAL_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_CHARACTER_SCRAMBLE_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_JAVASCRIPT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_COUNTRY_CONFIG_FIELD_NUMBER: _ClassVar[int]
    TRANSFORM_PII_TEXT_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_BUSINESS_NAME_CONFIG_FIELD_NUMBER: _ClassVar[int]
    GENERATE_IP_ADDRESS_CONFIG_FIELD_NUMBER: _ClassVar[int]
    generate_email_config: GenerateEmail
    transform_email_config: TransformEmail
    generate_bool_config: GenerateBool
    generate_card_number_config: GenerateCardNumber
    generate_city_config: GenerateCity
    generate_e164_phone_number_config: GenerateE164PhoneNumber
    generate_first_name_config: GenerateFirstName
    generate_float64_config: GenerateFloat64
    generate_full_address_config: GenerateFullAddress
    generate_full_name_config: GenerateFullName
    generate_gender_config: GenerateGender
    generate_int64_phone_number_config: GenerateInt64PhoneNumber
    generate_int64_config: GenerateInt64
    generate_last_name_config: GenerateLastName
    generate_sha256hash_config: GenerateSha256Hash
    generate_ssn_config: GenerateSSN
    generate_state_config: GenerateState
    generate_street_address_config: GenerateStreetAddress
    generate_string_phone_number_config: GenerateStringPhoneNumber
    generate_string_config: GenerateString
    generate_unixtimestamp_config: GenerateUnixTimestamp
    generate_username_config: GenerateUsername
    generate_utctimestamp_config: GenerateUtcTimestamp
    generate_uuid_config: GenerateUuid
    generate_zipcode_config: GenerateZipcode
    transform_e164_phone_number_config: TransformE164PhoneNumber
    transform_first_name_config: TransformFirstName
    transform_float64_config: TransformFloat64
    transform_full_name_config: TransformFullName
    transform_int64_phone_number_config: TransformInt64PhoneNumber
    transform_int64_config: TransformInt64
    transform_last_name_config: TransformLastName
    transform_phone_number_config: TransformPhoneNumber
    transform_string_config: TransformString
    passthrough_config: Passthrough
    nullconfig: Null
    user_defined_transformer_config: UserDefinedTransformerConfig
    generate_default_config: GenerateDefault
    transform_javascript_config: TransformJavascript
    generate_categorical_config: GenerateCategorical
    transform_character_scramble_config: TransformCharacterScramble
    generate_javascript_config: GenerateJavascript
    generate_country_config: GenerateCountry
    transform_pii_text_config: TransformPiiText
    generate_business_name_config: GenerateBusinessName
    generate_ip_address_config: GenerateIpAddress
    def __init__(self, generate_email_config: _Optional[_Union[GenerateEmail, _Mapping]] = ..., transform_email_config: _Optional[_Union[TransformEmail, _Mapping]] = ..., generate_bool_config: _Optional[_Union[GenerateBool, _Mapping]] = ..., generate_card_number_config: _Optional[_Union[GenerateCardNumber, _Mapping]] = ..., generate_city_config: _Optional[_Union[GenerateCity, _Mapping]] = ..., generate_e164_phone_number_config: _Optional[_Union[GenerateE164PhoneNumber, _Mapping]] = ..., generate_first_name_config: _Optional[_Union[GenerateFirstName, _Mapping]] = ..., generate_float64_config: _Optional[_Union[GenerateFloat64, _Mapping]] = ..., generate_full_address_config: _Optional[_Union[GenerateFullAddress, _Mapping]] = ..., generate_full_name_config: _Optional[_Union[GenerateFullName, _Mapping]] = ..., generate_gender_config: _Optional[_Union[GenerateGender, _Mapping]] = ..., generate_int64_phone_number_config: _Optional[_Union[GenerateInt64PhoneNumber, _Mapping]] = ..., generate_int64_config: _Optional[_Union[GenerateInt64, _Mapping]] = ..., generate_last_name_config: _Optional[_Union[GenerateLastName, _Mapping]] = ..., generate_sha256hash_config: _Optional[_Union[GenerateSha256Hash, _Mapping]] = ..., generate_ssn_config: _Optional[_Union[GenerateSSN, _Mapping]] = ..., generate_state_config: _Optional[_Union[GenerateState, _Mapping]] = ..., generate_street_address_config: _Optional[_Union[GenerateStreetAddress, _Mapping]] = ..., generate_string_phone_number_config: _Optional[_Union[GenerateStringPhoneNumber, _Mapping]] = ..., generate_string_config: _Optional[_Union[GenerateString, _Mapping]] = ..., generate_unixtimestamp_config: _Optional[_Union[GenerateUnixTimestamp, _Mapping]] = ..., generate_username_config: _Optional[_Union[GenerateUsername, _Mapping]] = ..., generate_utctimestamp_config: _Optional[_Union[GenerateUtcTimestamp, _Mapping]] = ..., generate_uuid_config: _Optional[_Union[GenerateUuid, _Mapping]] = ..., generate_zipcode_config: _Optional[_Union[GenerateZipcode, _Mapping]] = ..., transform_e164_phone_number_config: _Optional[_Union[TransformE164PhoneNumber, _Mapping]] = ..., transform_first_name_config: _Optional[_Union[TransformFirstName, _Mapping]] = ..., transform_float64_config: _Optional[_Union[TransformFloat64, _Mapping]] = ..., transform_full_name_config: _Optional[_Union[TransformFullName, _Mapping]] = ..., transform_int64_phone_number_config: _Optional[_Union[TransformInt64PhoneNumber, _Mapping]] = ..., transform_int64_config: _Optional[_Union[TransformInt64, _Mapping]] = ..., transform_last_name_config: _Optional[_Union[TransformLastName, _Mapping]] = ..., transform_phone_number_config: _Optional[_Union[TransformPhoneNumber, _Mapping]] = ..., transform_string_config: _Optional[_Union[TransformString, _Mapping]] = ..., passthrough_config: _Optional[_Union[Passthrough, _Mapping]] = ..., nullconfig: _Optional[_Union[Null, _Mapping]] = ..., user_defined_transformer_config: _Optional[_Union[UserDefinedTransformerConfig, _Mapping]] = ..., generate_default_config: _Optional[_Union[GenerateDefault, _Mapping]] = ..., transform_javascript_config: _Optional[_Union[TransformJavascript, _Mapping]] = ..., generate_categorical_config: _Optional[_Union[GenerateCategorical, _Mapping]] = ..., transform_character_scramble_config: _Optional[_Union[TransformCharacterScramble, _Mapping]] = ..., generate_javascript_config: _Optional[_Union[GenerateJavascript, _Mapping]] = ..., generate_country_config: _Optional[_Union[GenerateCountry, _Mapping]] = ..., transform_pii_text_config: _Optional[_Union[TransformPiiText, _Mapping]] = ..., generate_business_name_config: _Optional[_Union[GenerateBusinessName, _Mapping]] = ..., generate_ip_address_config: _Optional[_Union[GenerateIpAddress, _Mapping]] = ...) -> None: ...

class TransformPiiText(_message.Message):
    __slots__ = ("score_threshold", "default_anonymizer", "deny_recognizers", "allowed_entities", "allowed_phrases", "language")
    SCORE_THRESHOLD_FIELD_NUMBER: _ClassVar[int]
    DEFAULT_ANONYMIZER_FIELD_NUMBER: _ClassVar[int]
    DENY_RECOGNIZERS_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_ENTITIES_FIELD_NUMBER: _ClassVar[int]
    ALLOWED_PHRASES_FIELD_NUMBER: _ClassVar[int]
    LANGUAGE_FIELD_NUMBER: _ClassVar[int]
    score_threshold: float
    default_anonymizer: PiiAnonymizer
    deny_recognizers: _containers.RepeatedCompositeFieldContainer[PiiDenyRecognizer]
    allowed_entities: _containers.RepeatedScalarFieldContainer[str]
    allowed_phrases: _containers.RepeatedScalarFieldContainer[str]
    language: str
    def __init__(self, score_threshold: _Optional[float] = ..., default_anonymizer: _Optional[_Union[PiiAnonymizer, _Mapping]] = ..., deny_recognizers: _Optional[_Iterable[_Union[PiiDenyRecognizer, _Mapping]]] = ..., allowed_entities: _Optional[_Iterable[str]] = ..., allowed_phrases: _Optional[_Iterable[str]] = ..., language: _Optional[str] = ...) -> None: ...

class PiiDenyRecognizer(_message.Message):
    __slots__ = ("name", "deny_words")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DENY_WORDS_FIELD_NUMBER: _ClassVar[int]
    name: str
    deny_words: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, name: _Optional[str] = ..., deny_words: _Optional[_Iterable[str]] = ...) -> None: ...

class PiiAnonymizer(_message.Message):
    __slots__ = ("replace", "redact", "mask", "hash")
    class Replace(_message.Message):
        __slots__ = ("value",)
        VALUE_FIELD_NUMBER: _ClassVar[int]
        value: str
        def __init__(self, value: _Optional[str] = ...) -> None: ...
    class Redact(_message.Message):
        __slots__ = ()
        def __init__(self) -> None: ...
    class Mask(_message.Message):
        __slots__ = ("masking_char", "chars_to_mask", "from_end")
        MASKING_CHAR_FIELD_NUMBER: _ClassVar[int]
        CHARS_TO_MASK_FIELD_NUMBER: _ClassVar[int]
        FROM_END_FIELD_NUMBER: _ClassVar[int]
        masking_char: str
        chars_to_mask: int
        from_end: bool
        def __init__(self, masking_char: _Optional[str] = ..., chars_to_mask: _Optional[int] = ..., from_end: bool = ...) -> None: ...
    class Hash(_message.Message):
        __slots__ = ("algo",)
        class HashType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
            __slots__ = ()
            HASH_TYPE_UNSPECIFIED: _ClassVar[PiiAnonymizer.Hash.HashType]
            HASH_TYPE_MD5: _ClassVar[PiiAnonymizer.Hash.HashType]
            HASH_TYPE_SHA256: _ClassVar[PiiAnonymizer.Hash.HashType]
            HASH_TYPE_SHA512: _ClassVar[PiiAnonymizer.Hash.HashType]
        HASH_TYPE_UNSPECIFIED: PiiAnonymizer.Hash.HashType
        HASH_TYPE_MD5: PiiAnonymizer.Hash.HashType
        HASH_TYPE_SHA256: PiiAnonymizer.Hash.HashType
        HASH_TYPE_SHA512: PiiAnonymizer.Hash.HashType
        ALGO_FIELD_NUMBER: _ClassVar[int]
        algo: PiiAnonymizer.Hash.HashType
        def __init__(self, algo: _Optional[_Union[PiiAnonymizer.Hash.HashType, str]] = ...) -> None: ...
    REPLACE_FIELD_NUMBER: _ClassVar[int]
    REDACT_FIELD_NUMBER: _ClassVar[int]
    MASK_FIELD_NUMBER: _ClassVar[int]
    HASH_FIELD_NUMBER: _ClassVar[int]
    replace: PiiAnonymizer.Replace
    redact: PiiAnonymizer.Redact
    mask: PiiAnonymizer.Mask
    hash: PiiAnonymizer.Hash
    def __init__(self, replace: _Optional[_Union[PiiAnonymizer.Replace, _Mapping]] = ..., redact: _Optional[_Union[PiiAnonymizer.Redact, _Mapping]] = ..., mask: _Optional[_Union[PiiAnonymizer.Mask, _Mapping]] = ..., hash: _Optional[_Union[PiiAnonymizer.Hash, _Mapping]] = ...) -> None: ...

class GenerateEmail(_message.Message):
    __slots__ = ("email_type",)
    EMAIL_TYPE_FIELD_NUMBER: _ClassVar[int]
    email_type: GenerateEmailType
    def __init__(self, email_type: _Optional[_Union[GenerateEmailType, str]] = ...) -> None: ...

class TransformEmail(_message.Message):
    __slots__ = ("preserve_domain", "preserve_length", "excluded_domains", "email_type", "invalid_email_action")
    PRESERVE_DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    EXCLUDED_DOMAINS_FIELD_NUMBER: _ClassVar[int]
    EMAIL_TYPE_FIELD_NUMBER: _ClassVar[int]
    INVALID_EMAIL_ACTION_FIELD_NUMBER: _ClassVar[int]
    preserve_domain: bool
    preserve_length: bool
    excluded_domains: _containers.RepeatedScalarFieldContainer[str]
    email_type: GenerateEmailType
    invalid_email_action: InvalidEmailAction
    def __init__(self, preserve_domain: bool = ..., preserve_length: bool = ..., excluded_domains: _Optional[_Iterable[str]] = ..., email_type: _Optional[_Union[GenerateEmailType, str]] = ..., invalid_email_action: _Optional[_Union[InvalidEmailAction, str]] = ...) -> None: ...

class GenerateBool(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateCardNumber(_message.Message):
    __slots__ = ("valid_luhn",)
    VALID_LUHN_FIELD_NUMBER: _ClassVar[int]
    valid_luhn: bool
    def __init__(self, valid_luhn: bool = ...) -> None: ...

class GenerateCity(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateDefault(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateE164PhoneNumber(_message.Message):
    __slots__ = ("min", "max")
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    min: int
    max: int
    def __init__(self, min: _Optional[int] = ..., max: _Optional[int] = ...) -> None: ...

class GenerateFirstName(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateFloat64(_message.Message):
    __slots__ = ("randomize_sign", "min", "max", "precision")
    RANDOMIZE_SIGN_FIELD_NUMBER: _ClassVar[int]
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    PRECISION_FIELD_NUMBER: _ClassVar[int]
    randomize_sign: bool
    min: float
    max: float
    precision: int
    def __init__(self, randomize_sign: bool = ..., min: _Optional[float] = ..., max: _Optional[float] = ..., precision: _Optional[int] = ...) -> None: ...

class GenerateFullAddress(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateFullName(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateGender(_message.Message):
    __slots__ = ("abbreviate",)
    ABBREVIATE_FIELD_NUMBER: _ClassVar[int]
    abbreviate: bool
    def __init__(self, abbreviate: bool = ...) -> None: ...

class GenerateInt64PhoneNumber(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateInt64(_message.Message):
    __slots__ = ("randomize_sign", "min", "max")
    RANDOMIZE_SIGN_FIELD_NUMBER: _ClassVar[int]
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    randomize_sign: bool
    min: int
    max: int
    def __init__(self, randomize_sign: bool = ..., min: _Optional[int] = ..., max: _Optional[int] = ...) -> None: ...

class GenerateLastName(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateSha256Hash(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateSSN(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateState(_message.Message):
    __slots__ = ("generate_full_name",)
    GENERATE_FULL_NAME_FIELD_NUMBER: _ClassVar[int]
    generate_full_name: bool
    def __init__(self, generate_full_name: bool = ...) -> None: ...

class GenerateStreetAddress(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateStringPhoneNumber(_message.Message):
    __slots__ = ("min", "max")
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    min: int
    max: int
    def __init__(self, min: _Optional[int] = ..., max: _Optional[int] = ...) -> None: ...

class GenerateString(_message.Message):
    __slots__ = ("min", "max")
    MIN_FIELD_NUMBER: _ClassVar[int]
    MAX_FIELD_NUMBER: _ClassVar[int]
    min: int
    max: int
    def __init__(self, min: _Optional[int] = ..., max: _Optional[int] = ...) -> None: ...

class GenerateUnixTimestamp(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateUsername(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateUtcTimestamp(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateUuid(_message.Message):
    __slots__ = ("include_hyphens",)
    INCLUDE_HYPHENS_FIELD_NUMBER: _ClassVar[int]
    include_hyphens: bool
    def __init__(self, include_hyphens: bool = ...) -> None: ...

class GenerateZipcode(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class TransformE164PhoneNumber(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformFirstName(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformFloat64(_message.Message):
    __slots__ = ("randomization_range_min", "randomization_range_max")
    RANDOMIZATION_RANGE_MIN_FIELD_NUMBER: _ClassVar[int]
    RANDOMIZATION_RANGE_MAX_FIELD_NUMBER: _ClassVar[int]
    randomization_range_min: float
    randomization_range_max: float
    def __init__(self, randomization_range_min: _Optional[float] = ..., randomization_range_max: _Optional[float] = ...) -> None: ...

class TransformFullName(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformInt64PhoneNumber(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformInt64(_message.Message):
    __slots__ = ("randomization_range_min", "randomization_range_max")
    RANDOMIZATION_RANGE_MIN_FIELD_NUMBER: _ClassVar[int]
    RANDOMIZATION_RANGE_MAX_FIELD_NUMBER: _ClassVar[int]
    randomization_range_min: int
    randomization_range_max: int
    def __init__(self, randomization_range_min: _Optional[int] = ..., randomization_range_max: _Optional[int] = ...) -> None: ...

class TransformLastName(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformPhoneNumber(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class TransformString(_message.Message):
    __slots__ = ("preserve_length",)
    PRESERVE_LENGTH_FIELD_NUMBER: _ClassVar[int]
    preserve_length: bool
    def __init__(self, preserve_length: bool = ...) -> None: ...

class Passthrough(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Null(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class TransformJavascript(_message.Message):
    __slots__ = ("code",)
    CODE_FIELD_NUMBER: _ClassVar[int]
    code: str
    def __init__(self, code: _Optional[str] = ...) -> None: ...

class UserDefinedTransformerConfig(_message.Message):
    __slots__ = ("id",)
    ID_FIELD_NUMBER: _ClassVar[int]
    id: str
    def __init__(self, id: _Optional[str] = ...) -> None: ...

class ValidateUserJavascriptCodeRequest(_message.Message):
    __slots__ = ("account_id", "code")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    CODE_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    code: str
    def __init__(self, account_id: _Optional[str] = ..., code: _Optional[str] = ...) -> None: ...

class ValidateUserJavascriptCodeResponse(_message.Message):
    __slots__ = ("valid",)
    VALID_FIELD_NUMBER: _ClassVar[int]
    valid: bool
    def __init__(self, valid: bool = ...) -> None: ...

class GenerateCategorical(_message.Message):
    __slots__ = ("categories",)
    CATEGORIES_FIELD_NUMBER: _ClassVar[int]
    categories: str
    def __init__(self, categories: _Optional[str] = ...) -> None: ...

class TransformCharacterScramble(_message.Message):
    __slots__ = ("user_provided_regex",)
    USER_PROVIDED_REGEX_FIELD_NUMBER: _ClassVar[int]
    user_provided_regex: str
    def __init__(self, user_provided_regex: _Optional[str] = ...) -> None: ...

class GenerateJavascript(_message.Message):
    __slots__ = ("code",)
    CODE_FIELD_NUMBER: _ClassVar[int]
    code: str
    def __init__(self, code: _Optional[str] = ...) -> None: ...

class ValidateUserRegexCodeRequest(_message.Message):
    __slots__ = ("account_id", "user_provided_regex")
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_PROVIDED_REGEX_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    user_provided_regex: str
    def __init__(self, account_id: _Optional[str] = ..., user_provided_regex: _Optional[str] = ...) -> None: ...

class ValidateUserRegexCodeResponse(_message.Message):
    __slots__ = ("valid",)
    VALID_FIELD_NUMBER: _ClassVar[int]
    valid: bool
    def __init__(self, valid: bool = ...) -> None: ...

class GenerateCountry(_message.Message):
    __slots__ = ("generate_full_name",)
    GENERATE_FULL_NAME_FIELD_NUMBER: _ClassVar[int]
    generate_full_name: bool
    def __init__(self, generate_full_name: bool = ...) -> None: ...

class GetTransformPiiEntitiesRequest(_message.Message):
    __slots__ = ("account_id",)
    ACCOUNT_ID_FIELD_NUMBER: _ClassVar[int]
    account_id: str
    def __init__(self, account_id: _Optional[str] = ...) -> None: ...

class GetTransformPiiEntitiesResponse(_message.Message):
    __slots__ = ("entities",)
    ENTITIES_FIELD_NUMBER: _ClassVar[int]
    entities: _containers.RepeatedScalarFieldContainer[str]
    def __init__(self, entities: _Optional[_Iterable[str]] = ...) -> None: ...

class GenerateBusinessName(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GenerateIpAddress(_message.Message):
    __slots__ = ("ip_type",)
    IP_TYPE_FIELD_NUMBER: _ClassVar[int]
    ip_type: GenerateIpAddressType
    def __init__(self, ip_type: _Optional[_Union[GenerateIpAddressType, str]] = ...) -> None: ...
