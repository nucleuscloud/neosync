# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: mgmt/v1alpha1/api_key.proto
# Protobuf Python Version: 5.29.3
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    5,
    29,
    3,
    '',
    'mgmt/v1alpha1/api_key.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from buf.validate import validate_pb2 as buf_dot_validate_dot_validate__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1bmgmt/v1alpha1/api_key.proto\x12\rmgmt.v1alpha1\x1a\x1b\x62uf/validate/validate.proto\x1a\x1fgoogle/protobuf/timestamp.proto\"\xc3\x01\n\x1a\x43reateAccountApiKeyRequest\x12\'\n\naccount_id\x18\x01 \x01(\tB\x08\xbaH\x05r\x03\xb0\x01\x01R\taccountId\x12-\n\x04name\x18\x02 \x01(\tB\x19\xbaH\x16r\x14\x32\x12^[a-z0-9-]{3,100}$R\x04name\x12M\n\nexpires_at\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x12\xbaH\x0f\xb2\x01\t@\x01J\x05\x08\x80\xe7\x84\x0f\xc8\x01\x01R\texpiresAt\"T\n\x1b\x43reateAccountApiKeyResponse\x12\x35\n\x07\x61pi_key\x18\x01 \x01(\x0b\x32\x1c.mgmt.v1alpha1.AccountApiKeyR\x06\x61piKey\"\x94\x03\n\rAccountApiKey\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\x12\x1d\n\naccount_id\x18\x03 \x01(\tR\taccountId\x12\"\n\rcreated_by_id\x18\x04 \x01(\tR\x0b\x63reatedById\x12\x39\n\ncreated_at\x18\x05 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tcreatedAt\x12\"\n\rupdated_by_id\x18\x06 \x01(\tR\x0bupdatedById\x12\x39\n\nupdated_at\x18\x07 \x01(\x0b\x32\x1a.google.protobuf.TimestampR\tupdatedAt\x12 \n\tkey_value\x18\x08 \x01(\tH\x00R\x08keyValue\x88\x01\x01\x12\x17\n\x07user_id\x18\t \x01(\tR\x06userId\x12\x39\n\nexpires_at\x18\n \x01(\x0b\x32\x1a.google.protobuf.TimestampR\texpiresAtB\x0c\n\n_key_value\"C\n\x18GetAccountApiKeysRequest\x12\'\n\naccount_id\x18\x01 \x01(\tB\x08\xbaH\x05r\x03\xb0\x01\x01R\taccountId\"T\n\x19GetAccountApiKeysResponse\x12\x37\n\x08\x61pi_keys\x18\x01 \x03(\x0b\x32\x1c.mgmt.v1alpha1.AccountApiKeyR\x07\x61piKeys\"3\n\x17GetAccountApiKeyRequest\x12\x18\n\x02id\x18\x01 \x01(\tB\x08\xbaH\x05r\x03\xb0\x01\x01R\x02id\"Q\n\x18GetAccountApiKeyResponse\x12\x35\n\x07\x61pi_key\x18\x01 \x01(\x0b\x32\x1c.mgmt.v1alpha1.AccountApiKeyR\x06\x61piKey\"\x89\x01\n\x1eRegenerateAccountApiKeyRequest\x12\x18\n\x02id\x18\x01 \x01(\tB\x08\xbaH\x05r\x03\xb0\x01\x01R\x02id\x12M\n\nexpires_at\x18\x02 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x12\xbaH\x0f\xb2\x01\t@\x01J\x05\x08\x80\xe7\x84\x0f\xc8\x01\x01R\texpiresAt\"X\n\x1fRegenerateAccountApiKeyResponse\x12\x35\n\x07\x61pi_key\x18\x01 \x01(\x0b\x32\x1c.mgmt.v1alpha1.AccountApiKeyR\x06\x61piKey\"6\n\x1a\x44\x65leteAccountApiKeyRequest\x12\x18\n\x02id\x18\x01 \x01(\tB\x08\xbaH\x05r\x03\xb0\x01\x01R\x02id\"\x1d\n\x1b\x44\x65leteAccountApiKeyResponse2\xc2\x04\n\rApiKeyService\x12k\n\x11GetAccountApiKeys\x12\'.mgmt.v1alpha1.GetAccountApiKeysRequest\x1a(.mgmt.v1alpha1.GetAccountApiKeysResponse\"\x03\x90\x02\x01\x12h\n\x10GetAccountApiKey\x12&.mgmt.v1alpha1.GetAccountApiKeyRequest\x1a\'.mgmt.v1alpha1.GetAccountApiKeyResponse\"\x03\x90\x02\x01\x12n\n\x13\x43reateAccountApiKey\x12).mgmt.v1alpha1.CreateAccountApiKeyRequest\x1a*.mgmt.v1alpha1.CreateAccountApiKeyResponse\"\x00\x12z\n\x17RegenerateAccountApiKey\x12-.mgmt.v1alpha1.RegenerateAccountApiKeyRequest\x1a..mgmt.v1alpha1.RegenerateAccountApiKeyResponse\"\x00\x12n\n\x13\x44\x65leteAccountApiKey\x12).mgmt.v1alpha1.DeleteAccountApiKeyRequest\x1a*.mgmt.v1alpha1.DeleteAccountApiKeyResponse\"\x00\x42\xc7\x01\n\x11\x63om.mgmt.v1alpha1B\x0b\x41piKeyProtoP\x01ZPgithub.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1;mgmtv1alpha1\xa2\x02\x03MXX\xaa\x02\rMgmt.V1alpha1\xca\x02\rMgmt\\V1alpha1\xe2\x02\x19Mgmt\\V1alpha1\\GPBMetadata\xea\x02\x0eMgmt::V1alpha1b\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'mgmt.v1alpha1.api_key_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'\n\021com.mgmt.v1alpha1B\013ApiKeyProtoP\001ZPgithub.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1;mgmtv1alpha1\242\002\003MXX\252\002\rMgmt.V1alpha1\312\002\rMgmt\\V1alpha1\342\002\031Mgmt\\V1alpha1\\GPBMetadata\352\002\016Mgmt::V1alpha1'
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['account_id']._loaded_options = None
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['account_id']._serialized_options = b'\272H\005r\003\260\001\001'
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['name']._loaded_options = None
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['name']._serialized_options = b'\272H\026r\0242\022^[a-z0-9-]{3,100}$'
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['expires_at']._loaded_options = None
  _globals['_CREATEACCOUNTAPIKEYREQUEST'].fields_by_name['expires_at']._serialized_options = b'\272H\017\262\001\t@\001J\005\010\200\347\204\017\310\001\001'
  _globals['_GETACCOUNTAPIKEYSREQUEST'].fields_by_name['account_id']._loaded_options = None
  _globals['_GETACCOUNTAPIKEYSREQUEST'].fields_by_name['account_id']._serialized_options = b'\272H\005r\003\260\001\001'
  _globals['_GETACCOUNTAPIKEYREQUEST'].fields_by_name['id']._loaded_options = None
  _globals['_GETACCOUNTAPIKEYREQUEST'].fields_by_name['id']._serialized_options = b'\272H\005r\003\260\001\001'
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST'].fields_by_name['id']._loaded_options = None
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST'].fields_by_name['id']._serialized_options = b'\272H\005r\003\260\001\001'
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST'].fields_by_name['expires_at']._loaded_options = None
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST'].fields_by_name['expires_at']._serialized_options = b'\272H\017\262\001\t@\001J\005\010\200\347\204\017\310\001\001'
  _globals['_DELETEACCOUNTAPIKEYREQUEST'].fields_by_name['id']._loaded_options = None
  _globals['_DELETEACCOUNTAPIKEYREQUEST'].fields_by_name['id']._serialized_options = b'\272H\005r\003\260\001\001'
  _globals['_APIKEYSERVICE'].methods_by_name['GetAccountApiKeys']._loaded_options = None
  _globals['_APIKEYSERVICE'].methods_by_name['GetAccountApiKeys']._serialized_options = b'\220\002\001'
  _globals['_APIKEYSERVICE'].methods_by_name['GetAccountApiKey']._loaded_options = None
  _globals['_APIKEYSERVICE'].methods_by_name['GetAccountApiKey']._serialized_options = b'\220\002\001'
  _globals['_CREATEACCOUNTAPIKEYREQUEST']._serialized_start=109
  _globals['_CREATEACCOUNTAPIKEYREQUEST']._serialized_end=304
  _globals['_CREATEACCOUNTAPIKEYRESPONSE']._serialized_start=306
  _globals['_CREATEACCOUNTAPIKEYRESPONSE']._serialized_end=390
  _globals['_ACCOUNTAPIKEY']._serialized_start=393
  _globals['_ACCOUNTAPIKEY']._serialized_end=797
  _globals['_GETACCOUNTAPIKEYSREQUEST']._serialized_start=799
  _globals['_GETACCOUNTAPIKEYSREQUEST']._serialized_end=866
  _globals['_GETACCOUNTAPIKEYSRESPONSE']._serialized_start=868
  _globals['_GETACCOUNTAPIKEYSRESPONSE']._serialized_end=952
  _globals['_GETACCOUNTAPIKEYREQUEST']._serialized_start=954
  _globals['_GETACCOUNTAPIKEYREQUEST']._serialized_end=1005
  _globals['_GETACCOUNTAPIKEYRESPONSE']._serialized_start=1007
  _globals['_GETACCOUNTAPIKEYRESPONSE']._serialized_end=1088
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST']._serialized_start=1091
  _globals['_REGENERATEACCOUNTAPIKEYREQUEST']._serialized_end=1228
  _globals['_REGENERATEACCOUNTAPIKEYRESPONSE']._serialized_start=1230
  _globals['_REGENERATEACCOUNTAPIKEYRESPONSE']._serialized_end=1318
  _globals['_DELETEACCOUNTAPIKEYREQUEST']._serialized_start=1320
  _globals['_DELETEACCOUNTAPIKEYREQUEST']._serialized_end=1374
  _globals['_DELETEACCOUNTAPIKEYRESPONSE']._serialized_start=1376
  _globals['_DELETEACCOUNTAPIKEYRESPONSE']._serialized_end=1405
  _globals['_APIKEYSERVICE']._serialized_start=1408
  _globals['_APIKEYSERVICE']._serialized_end=1986
# @@protoc_insertion_point(module_scope)
