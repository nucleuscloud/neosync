// src/index.ts
import { Code, ConnectError } from "@connectrpc/connect";

// src/client/mgmt/v1alpha1/api_key_pb.ts
import { Message, proto3, Timestamp } from "@bufbuild/protobuf";
var CreateAccountApiKeyRequest = class _CreateAccountApiKeyRequest extends Message {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * Validate between now and one year: now < x < 365 days
   *
   * @generated from field: google.protobuf.Timestamp expires_at = 3;
   */
  expiresAt;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.CreateAccountApiKeyRequest";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "expires_at", kind: "message", T: Timestamp }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateAccountApiKeyRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateAccountApiKeyRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateAccountApiKeyRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_CreateAccountApiKeyRequest, a, b);
  }
};
var CreateAccountApiKeyResponse = class _CreateAccountApiKeyResponse extends Message {
  /**
   * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
   */
  apiKey;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.CreateAccountApiKeyResponse";
  static fields = proto3.util.newFieldList(() => [
    { no: 1, name: "api_key", kind: "message", T: AccountApiKey }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateAccountApiKeyResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateAccountApiKeyResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateAccountApiKeyResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_CreateAccountApiKeyResponse, a, b);
  }
};
var AccountApiKey = class _AccountApiKey extends Message {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * The friendly name of the API Key
   *
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string account_id = 3;
   */
  accountId = "";
  /**
   * @generated from field: string created_by_id = 4;
   */
  createdById = "";
  /**
   * @generated from field: google.protobuf.Timestamp created_at = 5;
   */
  createdAt;
  /**
   * @generated from field: string updated_by_id = 6;
   */
  updatedById = "";
  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 7;
   */
  updatedAt;
  /**
   * key_value is only returned on initial creation or when it is regenerated
   *
   * @generated from field: optional string key_value = 8;
   */
  keyValue;
  /**
   * @generated from field: string user_id = 9;
   */
  userId = "";
  /**
   * The timestamp of what the API key expires and will not longer be usable.
   *
   * @generated from field: google.protobuf.Timestamp expires_at = 10;
   */
  expiresAt;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.AccountApiKey";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "created_by_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 5, name: "created_at", kind: "message", T: Timestamp },
    {
      no: 6,
      name: "updated_by_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 7, name: "updated_at", kind: "message", T: Timestamp },
    { no: 8, name: "key_value", kind: "scalar", T: 9, opt: true },
    {
      no: 9,
      name: "user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 10, name: "expires_at", kind: "message", T: Timestamp }
  ]);
  static fromBinary(bytes, options) {
    return new _AccountApiKey().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AccountApiKey().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AccountApiKey().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_AccountApiKey, a, b);
  }
};
var GetAccountApiKeysRequest = class _GetAccountApiKeysRequest extends Message {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.GetAccountApiKeysRequest";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountApiKeysRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountApiKeysRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountApiKeysRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_GetAccountApiKeysRequest, a, b);
  }
};
var GetAccountApiKeysResponse = class _GetAccountApiKeysResponse extends Message {
  /**
   * @generated from field: repeated mgmt.v1alpha1.AccountApiKey api_keys = 1;
   */
  apiKeys = [];
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.GetAccountApiKeysResponse";
  static fields = proto3.util.newFieldList(() => [
    { no: 1, name: "api_keys", kind: "message", T: AccountApiKey, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountApiKeysResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountApiKeysResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountApiKeysResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_GetAccountApiKeysResponse, a, b);
  }
};
var GetAccountApiKeyRequest = class _GetAccountApiKeyRequest extends Message {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.GetAccountApiKeyRequest";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountApiKeyRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountApiKeyRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountApiKeyRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_GetAccountApiKeyRequest, a, b);
  }
};
var GetAccountApiKeyResponse = class _GetAccountApiKeyResponse extends Message {
  /**
   * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
   */
  apiKey;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.GetAccountApiKeyResponse";
  static fields = proto3.util.newFieldList(() => [
    { no: 1, name: "api_key", kind: "message", T: AccountApiKey }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountApiKeyResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountApiKeyResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountApiKeyResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_GetAccountApiKeyResponse, a, b);
  }
};
var RegenerateAccountApiKeyRequest = class _RegenerateAccountApiKeyRequest extends Message {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * Validate between now and one year: now < x < 365 days
   *
   * @generated from field: google.protobuf.Timestamp expires_at = 2;
   */
  expiresAt;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.RegenerateAccountApiKeyRequest";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "expires_at", kind: "message", T: Timestamp }
  ]);
  static fromBinary(bytes, options) {
    return new _RegenerateAccountApiKeyRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RegenerateAccountApiKeyRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RegenerateAccountApiKeyRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_RegenerateAccountApiKeyRequest, a, b);
  }
};
var RegenerateAccountApiKeyResponse = class _RegenerateAccountApiKeyResponse extends Message {
  /**
   * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
   */
  apiKey;
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.RegenerateAccountApiKeyResponse";
  static fields = proto3.util.newFieldList(() => [
    { no: 1, name: "api_key", kind: "message", T: AccountApiKey }
  ]);
  static fromBinary(bytes, options) {
    return new _RegenerateAccountApiKeyResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RegenerateAccountApiKeyResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RegenerateAccountApiKeyResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_RegenerateAccountApiKeyResponse, a, b);
  }
};
var DeleteAccountApiKeyRequest = class _DeleteAccountApiKeyRequest extends Message {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.DeleteAccountApiKeyRequest";
  static fields = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteAccountApiKeyRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteAccountApiKeyRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteAccountApiKeyRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_DeleteAccountApiKeyRequest, a, b);
  }
};
var DeleteAccountApiKeyResponse = class _DeleteAccountApiKeyResponse extends Message {
  constructor(data) {
    super();
    proto3.util.initPartial(data, this);
  }
  static runtime = proto3;
  static typeName = "mgmt.v1alpha1.DeleteAccountApiKeyResponse";
  static fields = proto3.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteAccountApiKeyResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteAccountApiKeyResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteAccountApiKeyResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto3.util.equals(_DeleteAccountApiKeyResponse, a, b);
  }
};

// src/client/mgmt/v1alpha1/api_key_connect.ts
import { MethodKind } from "@bufbuild/protobuf";
var ApiKeyService = {
  typeName: "mgmt.v1alpha1.ApiKeyService",
  methods: {
    /**
     * Retrieves a list of Account API Keys
     *
     * @generated from rpc mgmt.v1alpha1.ApiKeyService.GetAccountApiKeys
     */
    getAccountApiKeys: {
      name: "GetAccountApiKeys",
      I: GetAccountApiKeysRequest,
      O: GetAccountApiKeysResponse,
      kind: MethodKind.Unary
    },
    /**
     * Retrieves a single API Key
     *
     * @generated from rpc mgmt.v1alpha1.ApiKeyService.GetAccountApiKey
     */
    getAccountApiKey: {
      name: "GetAccountApiKey",
      I: GetAccountApiKeyRequest,
      O: GetAccountApiKeyResponse,
      kind: MethodKind.Unary
    },
    /**
     * Creates a single API Key
     * This method will return the decrypted contents of the API key
     *
     * @generated from rpc mgmt.v1alpha1.ApiKeyService.CreateAccountApiKey
     */
    createAccountApiKey: {
      name: "CreateAccountApiKey",
      I: CreateAccountApiKeyRequest,
      O: CreateAccountApiKeyResponse,
      kind: MethodKind.Unary
    },
    /**
     * Regenerates a single API Key with a new expiration time
     * This method will return the decrypted contents of the API key
     *
     * @generated from rpc mgmt.v1alpha1.ApiKeyService.RegenerateAccountApiKey
     */
    regenerateAccountApiKey: {
      name: "RegenerateAccountApiKey",
      I: RegenerateAccountApiKeyRequest,
      O: RegenerateAccountApiKeyResponse,
      kind: MethodKind.Unary
    },
    /**
     * Deletes an API Key from the system.
     *
     * @generated from rpc mgmt.v1alpha1.ApiKeyService.DeleteAccountApiKey
     */
    deleteAccountApiKey: {
      name: "DeleteAccountApiKey",
      I: DeleteAccountApiKeyRequest,
      O: DeleteAccountApiKeyResponse,
      kind: MethodKind.Unary
    }
  }
};

// src/client/mgmt/v1alpha1/connection_pb.ts
import { Message as Message2, proto3 as proto32, Timestamp as Timestamp2 } from "@bufbuild/protobuf";
var GetConnectionsRequest = class _GetConnectionsRequest extends Message2 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionsRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionsRequest, a, b);
  }
};
var GetConnectionsResponse = class _GetConnectionsResponse extends Message2 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.Connection connections = 1;
   */
  connections = [];
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionsResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "connections", kind: "message", T: Connection, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionsResponse, a, b);
  }
};
var GetConnectionRequest = class _GetConnectionRequest extends Message2 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionRequest, a, b);
  }
};
var GetConnectionResponse = class _GetConnectionResponse extends Message2 {
  /**
   * @generated from field: mgmt.v1alpha1.Connection connection = 1;
   */
  connection;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "connection", kind: "message", T: Connection }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionResponse, a, b);
  }
};
var CreateConnectionRequest = class _CreateConnectionRequest extends Message2 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * The friendly name of the connection
   *
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
   */
  connectionConfig;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CreateConnectionRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "connection_config", kind: "message", T: ConnectionConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CreateConnectionRequest, a, b);
  }
};
var CreateConnectionResponse = class _CreateConnectionResponse extends Message2 {
  /**
   * @generated from field: mgmt.v1alpha1.Connection connection = 1;
   */
  connection;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CreateConnectionResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "connection", kind: "message", T: Connection }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CreateConnectionResponse, a, b);
  }
};
var UpdateConnectionRequest = class _UpdateConnectionRequest extends Message2 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
   */
  connectionConfig;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.UpdateConnectionRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "connection_config", kind: "message", T: ConnectionConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_UpdateConnectionRequest, a, b);
  }
};
var UpdateConnectionResponse = class _UpdateConnectionResponse extends Message2 {
  /**
   * @generated from field: mgmt.v1alpha1.Connection connection = 1;
   */
  connection;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.UpdateConnectionResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "connection", kind: "message", T: Connection }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_UpdateConnectionResponse, a, b);
  }
};
var DeleteConnectionRequest = class _DeleteConnectionRequest extends Message2 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.DeleteConnectionRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_DeleteConnectionRequest, a, b);
  }
};
var DeleteConnectionResponse = class _DeleteConnectionResponse extends Message2 {
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.DeleteConnectionResponse";
  static fields = proto32.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_DeleteConnectionResponse, a, b);
  }
};
var CheckConnectionConfigRequest = class _CheckConnectionConfigRequest extends Message2 {
  /**
   * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 1;
   */
  connectionConfig;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CheckConnectionConfigRequest";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "connection_config", kind: "message", T: ConnectionConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _CheckConnectionConfigRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CheckConnectionConfigRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CheckConnectionConfigRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CheckConnectionConfigRequest, a, b);
  }
};
var CheckConnectionConfigResponse = class _CheckConnectionConfigResponse extends Message2 {
  /**
   * Whether or not the API was able to ping the connection
   *
   * @generated from field: bool is_connected = 1;
   */
  isConnected = false;
  /**
   * This is the error that was received if the API was unable to connect
   *
   * @generated from field: optional string connection_error = 2;
   */
  connectionError;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CheckConnectionConfigResponse";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "is_connected",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 2, name: "connection_error", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _CheckConnectionConfigResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CheckConnectionConfigResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CheckConnectionConfigResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CheckConnectionConfigResponse, a, b);
  }
};
var Connection = class _Connection extends Message2 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
   */
  connectionConfig;
  /**
   * @generated from field: string created_by_user_id = 4;
   */
  createdByUserId = "";
  /**
   * @generated from field: google.protobuf.Timestamp created_at = 5;
   */
  createdAt;
  /**
   * @generated from field: string updated_by_user_id = 6;
   */
  updatedByUserId = "";
  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 7;
   */
  updatedAt;
  /**
   * @generated from field: string account_id = 8;
   */
  accountId = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.Connection";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "connection_config", kind: "message", T: ConnectionConfig },
    {
      no: 4,
      name: "created_by_user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 5, name: "created_at", kind: "message", T: Timestamp2 },
    {
      no: 6,
      name: "updated_by_user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 7, name: "updated_at", kind: "message", T: Timestamp2 },
    {
      no: 8,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _Connection().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _Connection().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _Connection().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_Connection, a, b);
  }
};
var ConnectionConfig = class _ConnectionConfig extends Message2 {
  /**
   * @generated from oneof mgmt.v1alpha1.ConnectionConfig.config
   */
  config = { case: void 0 };
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.ConnectionConfig";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "pg_config", kind: "message", T: PostgresConnectionConfig, oneof: "config" },
    { no: 2, name: "aws_s3_config", kind: "message", T: AwsS3ConnectionConfig, oneof: "config" },
    { no: 3, name: "mysql_config", kind: "message", T: MysqlConnectionConfig, oneof: "config" }
  ]);
  static fromBinary(bytes, options) {
    return new _ConnectionConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _ConnectionConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _ConnectionConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_ConnectionConfig, a, b);
  }
};
var PostgresConnectionConfig = class _PostgresConnectionConfig extends Message2 {
  /**
   * May provide either a raw string url, or a structured version
   *
   * @generated from oneof mgmt.v1alpha1.PostgresConnectionConfig.connection_config
   */
  connectionConfig = { case: void 0 };
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.PostgresConnectionConfig";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "url", kind: "scalar", T: 9, oneof: "connection_config" },
    { no: 2, name: "connection", kind: "message", T: PostgresConnection, oneof: "connection_config" }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresConnectionConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresConnectionConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresConnectionConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_PostgresConnectionConfig, a, b);
  }
};
var PostgresConnection = class _PostgresConnection extends Message2 {
  /**
   * @generated from field: string host = 1;
   */
  host = "";
  /**
   * @generated from field: int32 port = 2;
   */
  port = 0;
  /**
   * @generated from field: string name = 3;
   */
  name = "";
  /**
   * @generated from field: string user = 4;
   */
  user = "";
  /**
   * @generated from field: string pass = 5;
   */
  pass = "";
  /**
   * @generated from field: optional string ssl_mode = 6;
   */
  sslMode;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.PostgresConnection";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "host",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "port",
      kind: "scalar",
      T: 5
      /* ScalarType.INT32 */
    },
    {
      no: 3,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "user",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "pass",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 6, name: "ssl_mode", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresConnection().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresConnection().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresConnection().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_PostgresConnection, a, b);
  }
};
var MysqlConnection = class _MysqlConnection extends Message2 {
  /**
   * @generated from field: string user = 1;
   */
  user = "";
  /**
   * @generated from field: string pass = 2;
   */
  pass = "";
  /**
   * @generated from field: string protocol = 3;
   */
  protocol = "";
  /**
   * @generated from field: string host = 4;
   */
  host = "";
  /**
   * @generated from field: int32 port = 5;
   */
  port = 0;
  /**
   * @generated from field: string name = 6;
   */
  name = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.MysqlConnection";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "user",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "pass",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "protocol",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "host",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "port",
      kind: "scalar",
      T: 5
      /* ScalarType.INT32 */
    },
    {
      no: 6,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlConnection().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlConnection().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlConnection().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_MysqlConnection, a, b);
  }
};
var MysqlConnectionConfig = class _MysqlConnectionConfig extends Message2 {
  /**
   * May provide either a raw string url, or a structured version
   *
   * @generated from oneof mgmt.v1alpha1.MysqlConnectionConfig.connection_config
   */
  connectionConfig = { case: void 0 };
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.MysqlConnectionConfig";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "url", kind: "scalar", T: 9, oneof: "connection_config" },
    { no: 2, name: "connection", kind: "message", T: MysqlConnection, oneof: "connection_config" }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlConnectionConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlConnectionConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlConnectionConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_MysqlConnectionConfig, a, b);
  }
};
var AwsS3ConnectionConfig = class _AwsS3ConnectionConfig extends Message2 {
  /**
   * @generated from field: string bucket_arn = 1;
   */
  bucketArn = "";
  /**
   * @generated from field: optional string path_prefix = 2;
   */
  pathPrefix;
  /**
   * @generated from field: optional mgmt.v1alpha1.AwsS3Credentials credentials = 3;
   */
  credentials;
  /**
   * @generated from field: optional string region = 4;
   */
  region;
  /**
   * @generated from field: optional string endpoint = 5;
   */
  endpoint;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.AwsS3ConnectionConfig";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "bucket_arn",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "path_prefix", kind: "scalar", T: 9, opt: true },
    { no: 3, name: "credentials", kind: "message", T: AwsS3Credentials, opt: true },
    { no: 4, name: "region", kind: "scalar", T: 9, opt: true },
    { no: 5, name: "endpoint", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _AwsS3ConnectionConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AwsS3ConnectionConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AwsS3ConnectionConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_AwsS3ConnectionConfig, a, b);
  }
};
var AwsS3Credentials = class _AwsS3Credentials extends Message2 {
  /**
   * @generated from field: optional string profile = 1;
   */
  profile;
  /**
   * @generated from field: optional string access_key_id = 2;
   */
  accessKeyId;
  /**
   * @generated from field: optional string secret_access_key = 3;
   */
  secretAccessKey;
  /**
   * @generated from field: optional string session_token = 4;
   */
  sessionToken;
  /**
   * @generated from field: optional bool from_ec2_role = 5;
   */
  fromEc2Role;
  /**
   * @generated from field: optional string role_arn = 6;
   */
  roleArn;
  /**
   * @generated from field: optional string role_external_id = 7;
   */
  roleExternalId;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.AwsS3Credentials";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "profile", kind: "scalar", T: 9, opt: true },
    { no: 2, name: "access_key_id", kind: "scalar", T: 9, opt: true },
    { no: 3, name: "secret_access_key", kind: "scalar", T: 9, opt: true },
    { no: 4, name: "session_token", kind: "scalar", T: 9, opt: true },
    { no: 5, name: "from_ec2_role", kind: "scalar", T: 8, opt: true },
    { no: 6, name: "role_arn", kind: "scalar", T: 9, opt: true },
    { no: 7, name: "role_external_id", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _AwsS3Credentials().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AwsS3Credentials().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AwsS3Credentials().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_AwsS3Credentials, a, b);
  }
};
var IsConnectionNameAvailableRequest = class _IsConnectionNameAvailableRequest extends Message2 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string connection_name = 2;
   */
  connectionName = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.IsConnectionNameAvailableRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "connection_name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsConnectionNameAvailableRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsConnectionNameAvailableRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsConnectionNameAvailableRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_IsConnectionNameAvailableRequest, a, b);
  }
};
var IsConnectionNameAvailableResponse = class _IsConnectionNameAvailableResponse extends Message2 {
  /**
   * @generated from field: bool is_available = 1;
   */
  isAvailable = false;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.IsConnectionNameAvailableResponse";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "is_available",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsConnectionNameAvailableResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsConnectionNameAvailableResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsConnectionNameAvailableResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_IsConnectionNameAvailableResponse, a, b);
  }
};
var DatabaseColumn = class _DatabaseColumn extends Message2 {
  /**
   * The database schema. Ex: public
   *
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * The name of the table in the schema
   *
   * @generated from field: string table = 2;
   */
  table = "";
  /**
   * The name of the column
   *
   * @generated from field: string column = 3;
   */
  column = "";
  /**
   * The datatype of the column
   *
   * @generated from field: string data_type = 4;
   */
  dataType = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.DatabaseColumn";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "column",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "data_type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DatabaseColumn().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DatabaseColumn().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DatabaseColumn().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_DatabaseColumn, a, b);
  }
};
var GetConnectionSchemaRequest = class _GetConnectionSchemaRequest extends Message2 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionSchemaRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionSchemaRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionSchemaRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionSchemaRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionSchemaRequest, a, b);
  }
};
var GetConnectionSchemaResponse = class _GetConnectionSchemaResponse extends Message2 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.DatabaseColumn schemas = 1;
   */
  schemas = [];
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionSchemaResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "schemas", kind: "message", T: DatabaseColumn, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionSchemaResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionSchemaResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionSchemaResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionSchemaResponse, a, b);
  }
};
var CheckSqlQueryRequest = class _CheckSqlQueryRequest extends Message2 {
  /**
   * The connection id that the query will be checked against
   *
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * The full query that will be run through a PREPARE statement
   *
   * @generated from field: string query = 2;
   */
  query = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CheckSqlQueryRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "query",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CheckSqlQueryRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CheckSqlQueryRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CheckSqlQueryRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CheckSqlQueryRequest, a, b);
  }
};
var CheckSqlQueryResponse = class _CheckSqlQueryResponse extends Message2 {
  /**
   * The query is run through PREPARE. Returns valid if it correctly compiled
   *
   * @generated from field: bool is_valid = 1;
   */
  isValid = false;
  /**
   * The error message returned by the sql client if the prepare did not return successfully
   *
   * @generated from field: optional string erorr_message = 2;
   */
  erorrMessage;
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.CheckSqlQueryResponse";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "is_valid",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 2, name: "erorr_message", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _CheckSqlQueryResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CheckSqlQueryResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CheckSqlQueryResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_CheckSqlQueryResponse, a, b);
  }
};
var GetConnectionDataStreamRequest = class _GetConnectionDataStreamRequest extends Message2 {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";
  /**
   * @generated from field: string schema = 2;
   */
  schema = "";
  /**
   * @generated from field: string table = 3;
   */
  table = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionDataStreamRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionDataStreamRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionDataStreamRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionDataStreamRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionDataStreamRequest, a, b);
  }
};
var GetConnectionDataStreamResponse = class _GetConnectionDataStreamResponse extends Message2 {
  /**
   * A map of column name to the bytes value of the data that was found for that column and row
   *
   * @generated from field: map<string, bytes> row = 1;
   */
  row = {};
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionDataStreamResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "row", kind: "map", K: 9, V: {
      kind: "scalar",
      T: 12
      /* ScalarType.BYTES */
    } }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionDataStreamResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionDataStreamResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionDataStreamResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionDataStreamResponse, a, b);
  }
};
var GetConnectionForeignConstraintsRequest = class _GetConnectionForeignConstraintsRequest extends Message2 {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionForeignConstraintsRequest";
  static fields = proto32.util.newFieldList(() => [
    {
      no: 1,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionForeignConstraintsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionForeignConstraintsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionForeignConstraintsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionForeignConstraintsRequest, a, b);
  }
};
var ForeignConstraintTables = class _ForeignConstraintTables extends Message2 {
  /**
   * @generated from field: repeated string tables = 1;
   */
  tables = [];
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.ForeignConstraintTables";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "tables", kind: "scalar", T: 9, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _ForeignConstraintTables().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _ForeignConstraintTables().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _ForeignConstraintTables().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_ForeignConstraintTables, a, b);
  }
};
var GetConnectionForeignConstraintsResponse = class _GetConnectionForeignConstraintsResponse extends Message2 {
  /**
   * the key here is <schema>.<table> and the list of tables that it depends on, also `<schema>.<table>` format.
   *
   * @generated from field: map<string, mgmt.v1alpha1.ForeignConstraintTables> table_constraints = 1;
   */
  tableConstraints = {};
  constructor(data) {
    super();
    proto32.util.initPartial(data, this);
  }
  static runtime = proto32;
  static typeName = "mgmt.v1alpha1.GetConnectionForeignConstraintsResponse";
  static fields = proto32.util.newFieldList(() => [
    { no: 1, name: "table_constraints", kind: "map", K: 9, V: { kind: "message", T: ForeignConstraintTables } }
  ]);
  static fromBinary(bytes, options) {
    return new _GetConnectionForeignConstraintsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetConnectionForeignConstraintsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetConnectionForeignConstraintsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto32.util.equals(_GetConnectionForeignConstraintsResponse, a, b);
  }
};

// src/client/mgmt/v1alpha1/connection_connect.ts
import { MethodKind as MethodKind2 } from "@bufbuild/protobuf";
var ConnectionService = {
  typeName: "mgmt.v1alpha1.ConnectionService",
  methods: {
    /**
     * Returns a list of connections associated with the account
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnections
     */
    getConnections: {
      name: "GetConnections",
      I: GetConnectionsRequest,
      O: GetConnectionsResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Returns a single connection
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnection
     */
    getConnection: {
      name: "GetConnection",
      I: GetConnectionRequest,
      O: GetConnectionResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Creates a new connection
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.CreateConnection
     */
    createConnection: {
      name: "CreateConnection",
      I: CreateConnectionRequest,
      O: CreateConnectionResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Updates an existing connection
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.UpdateConnection
     */
    updateConnection: {
      name: "UpdateConnection",
      I: UpdateConnectionRequest,
      O: UpdateConnectionResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Removes a connection from the system.
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.DeleteConnection
     */
    deleteConnection: {
      name: "DeleteConnection",
      I: DeleteConnectionRequest,
      O: DeleteConnectionResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Connections have friendly names, this method checks if the requested name is available in the system based on the account
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.IsConnectionNameAvailable
     */
    isConnectionNameAvailable: {
      name: "IsConnectionNameAvailable",
      I: IsConnectionNameAvailableRequest,
      O: IsConnectionNameAvailableResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Checks if the connection config is connectable by the backend.
     * Used mostly to verify that a connection is valid prior to creating a Connection object.
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.CheckConnectionConfig
     */
    checkConnectionConfig: {
      name: "CheckConnectionConfig",
      I: CheckConnectionConfigRequest,
      O: CheckConnectionConfigResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Returns the schema for a specific connection. Used mostly for SQL-based connections
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionSchema
     */
    getConnectionSchema: {
      name: "GetConnectionSchema",
      I: GetConnectionSchemaRequest,
      O: GetConnectionSchemaResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Checks a constructed SQL query against a sql-based connection to see if it's valid based on that connection's data schema
     * This is useful when constructing subsets to see if the WHERE clause is correct
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.CheckSqlQuery
     */
    checkSqlQuery: {
      name: "CheckSqlQuery",
      I: CheckSqlQueryRequest,
      O: CheckSqlQueryResponse,
      kind: MethodKind2.Unary
    },
    /**
     * Streaming endpoint that will stream the data available from the Connection to the client.
     * Used primarily by the CLI sync command.
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionDataStream
     */
    getConnectionDataStream: {
      name: "GetConnectionDataStream",
      I: GetConnectionDataStreamRequest,
      O: GetConnectionDataStreamResponse,
      kind: MethodKind2.ServerStreaming
    },
    /**
     * For a specific connection, returns the foreign key constraints. Mostly useful for SQL-based Connections.
     * Used primarily by the CLI sync command to determine stream order.
     *
     * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionForeignConstraints
     */
    getConnectionForeignConstraints: {
      name: "GetConnectionForeignConstraints",
      I: GetConnectionForeignConstraintsRequest,
      O: GetConnectionForeignConstraintsResponse,
      kind: MethodKind2.Unary
    }
  }
};

// src/client/mgmt/v1alpha1/job_pb.ts
import { Message as Message4, proto3 as proto34, protoInt64 as protoInt642, Timestamp as Timestamp4 } from "@bufbuild/protobuf";

// src/client/mgmt/v1alpha1/transformer_pb.ts
import { Message as Message3, proto3 as proto33, protoInt64, Timestamp as Timestamp3 } from "@bufbuild/protobuf";
var GetSystemTransformersRequest = class _GetSystemTransformersRequest extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetSystemTransformersRequest";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GetSystemTransformersRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetSystemTransformersRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetSystemTransformersRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetSystemTransformersRequest, a, b);
  }
};
var GetSystemTransformersResponse = class _GetSystemTransformersResponse extends Message3 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.SystemTransformer transformers = 1;
   */
  transformers = [];
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetSystemTransformersResponse";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "transformers", kind: "message", T: SystemTransformer, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetSystemTransformersResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetSystemTransformersResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetSystemTransformersResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetSystemTransformersResponse, a, b);
  }
};
var GetUserDefinedTransformersRequest = class _GetUserDefinedTransformersRequest extends Message3 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetUserDefinedTransformersRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserDefinedTransformersRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserDefinedTransformersRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserDefinedTransformersRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetUserDefinedTransformersRequest, a, b);
  }
};
var GetUserDefinedTransformersResponse = class _GetUserDefinedTransformersResponse extends Message3 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.UserDefinedTransformer transformers = 1;
   */
  transformers = [];
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetUserDefinedTransformersResponse";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "transformers", kind: "message", T: UserDefinedTransformer, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserDefinedTransformersResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserDefinedTransformersResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserDefinedTransformersResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetUserDefinedTransformersResponse, a, b);
  }
};
var GetUserDefinedTransformerByIdRequest = class _GetUserDefinedTransformerByIdRequest extends Message3 {
  /**
   * @generated from field: string transformer_id = 1;
   */
  transformerId = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetUserDefinedTransformerByIdRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "transformer_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserDefinedTransformerByIdRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserDefinedTransformerByIdRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserDefinedTransformerByIdRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetUserDefinedTransformerByIdRequest, a, b);
  }
};
var GetUserDefinedTransformerByIdResponse = class _GetUserDefinedTransformerByIdResponse extends Message3 {
  /**
   * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
   */
  transformer;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GetUserDefinedTransformerByIdResponse";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "transformer", kind: "message", T: UserDefinedTransformer }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserDefinedTransformerByIdResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserDefinedTransformerByIdResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserDefinedTransformerByIdResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GetUserDefinedTransformerByIdResponse, a, b);
  }
};
var CreateUserDefinedTransformerRequest = class _CreateUserDefinedTransformerRequest extends Message3 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string description = 3;
   */
  description = "";
  /**
   * @generated from field: string type = 4;
   */
  type = "";
  /**
   * @generated from field: string source = 5;
   */
  source = "";
  /**
   * @generated from field: mgmt.v1alpha1.TransformerConfig transformer_config = 6;
   */
  transformerConfig;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.CreateUserDefinedTransformerRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "description",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "source",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 6, name: "transformer_config", kind: "message", T: TransformerConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateUserDefinedTransformerRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateUserDefinedTransformerRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateUserDefinedTransformerRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_CreateUserDefinedTransformerRequest, a, b);
  }
};
var CreateUserDefinedTransformerResponse = class _CreateUserDefinedTransformerResponse extends Message3 {
  /**
   * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
   */
  transformer;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.CreateUserDefinedTransformerResponse";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "transformer", kind: "message", T: UserDefinedTransformer }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateUserDefinedTransformerResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateUserDefinedTransformerResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateUserDefinedTransformerResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_CreateUserDefinedTransformerResponse, a, b);
  }
};
var DeleteUserDefinedTransformerRequest = class _DeleteUserDefinedTransformerRequest extends Message3 {
  /**
   * @generated from field: string transformer_id = 1;
   */
  transformerId = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.DeleteUserDefinedTransformerRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "transformer_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteUserDefinedTransformerRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteUserDefinedTransformerRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteUserDefinedTransformerRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_DeleteUserDefinedTransformerRequest, a, b);
  }
};
var DeleteUserDefinedTransformerResponse = class _DeleteUserDefinedTransformerResponse extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.DeleteUserDefinedTransformerResponse";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteUserDefinedTransformerResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteUserDefinedTransformerResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteUserDefinedTransformerResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_DeleteUserDefinedTransformerResponse, a, b);
  }
};
var UpdateUserDefinedTransformerRequest = class _UpdateUserDefinedTransformerRequest extends Message3 {
  /**
   * @generated from field: string transformer_id = 1;
   */
  transformerId = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string description = 3;
   */
  description = "";
  /**
   * @generated from field: mgmt.v1alpha1.TransformerConfig transformer_config = 4;
   */
  transformerConfig;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.UpdateUserDefinedTransformerRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "transformer_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "description",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 4, name: "transformer_config", kind: "message", T: TransformerConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateUserDefinedTransformerRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateUserDefinedTransformerRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateUserDefinedTransformerRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_UpdateUserDefinedTransformerRequest, a, b);
  }
};
var UpdateUserDefinedTransformerResponse = class _UpdateUserDefinedTransformerResponse extends Message3 {
  /**
   * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
   */
  transformer;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.UpdateUserDefinedTransformerResponse";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "transformer", kind: "message", T: UserDefinedTransformer }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateUserDefinedTransformerResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateUserDefinedTransformerResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateUserDefinedTransformerResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_UpdateUserDefinedTransformerResponse, a, b);
  }
};
var IsTransformerNameAvailableRequest = class _IsTransformerNameAvailableRequest extends Message3 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string transformer_name = 2;
   */
  transformerName = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.IsTransformerNameAvailableRequest";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "transformer_name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsTransformerNameAvailableRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsTransformerNameAvailableRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsTransformerNameAvailableRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_IsTransformerNameAvailableRequest, a, b);
  }
};
var IsTransformerNameAvailableResponse = class _IsTransformerNameAvailableResponse extends Message3 {
  /**
   * @generated from field: bool is_available = 1;
   */
  isAvailable = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.IsTransformerNameAvailableResponse";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "is_available",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsTransformerNameAvailableResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsTransformerNameAvailableResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsTransformerNameAvailableResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_IsTransformerNameAvailableResponse, a, b);
  }
};
var UserDefinedTransformer = class _UserDefinedTransformer extends Message3 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string description = 3;
   */
  description = "";
  /**
   * @generated from field: string data_type = 5;
   */
  dataType = "";
  /**
   * @generated from field: string source = 6;
   */
  source = "";
  /**
   * @generated from field: mgmt.v1alpha1.TransformerConfig config = 7;
   */
  config;
  /**
   * @generated from field: google.protobuf.Timestamp created_at = 8;
   */
  createdAt;
  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 9;
   */
  updatedAt;
  /**
   * @generated from field: string account_id = 10;
   */
  accountId = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.UserDefinedTransformer";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "description",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "data_type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 6,
      name: "source",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 7, name: "config", kind: "message", T: TransformerConfig },
    { no: 8, name: "created_at", kind: "message", T: Timestamp3 },
    { no: 9, name: "updated_at", kind: "message", T: Timestamp3 },
    {
      no: 10,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _UserDefinedTransformer().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UserDefinedTransformer().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UserDefinedTransformer().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_UserDefinedTransformer, a, b);
  }
};
var SystemTransformer = class _SystemTransformer extends Message3 {
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string description = 3;
   */
  description = "";
  /**
   * @generated from field: string data_type = 5;
   */
  dataType = "";
  /**
   * @generated from field: string source = 6;
   */
  source = "";
  /**
   * @generated from field: mgmt.v1alpha1.TransformerConfig config = 7;
   */
  config;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.SystemTransformer";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "description",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "data_type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 6,
      name: "source",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 7, name: "config", kind: "message", T: TransformerConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _SystemTransformer().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SystemTransformer().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SystemTransformer().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_SystemTransformer, a, b);
  }
};
var TransformerConfig = class _TransformerConfig extends Message3 {
  /**
   * @generated from oneof mgmt.v1alpha1.TransformerConfig.config
   */
  config = { case: void 0 };
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformerConfig";
  static fields = proto33.util.newFieldList(() => [
    { no: 1, name: "generate_email_config", kind: "message", T: GenerateEmail, oneof: "config" },
    { no: 2, name: "generate_realistic_email_config", kind: "message", T: GenerateRealisticEmail, oneof: "config" },
    { no: 3, name: "transform_email_config", kind: "message", T: TransformEmail, oneof: "config" },
    { no: 4, name: "generate_bool_config", kind: "message", T: GenerateBool, oneof: "config" },
    { no: 5, name: "generate_card_number_config", kind: "message", T: GenerateCardNumber, oneof: "config" },
    { no: 6, name: "generate_city_config", kind: "message", T: GenerateCity, oneof: "config" },
    { no: 7, name: "generate_e164_number_config", kind: "message", T: GenerateE164Number, oneof: "config" },
    { no: 8, name: "generate_first_name_config", kind: "message", T: GenerateFirstName, oneof: "config" },
    { no: 9, name: "generate_float_config", kind: "message", T: GenerateFloat, oneof: "config" },
    { no: 10, name: "generate_full_address_config", kind: "message", T: GenerateFullAddress, oneof: "config" },
    { no: 11, name: "generate_full_name_config", kind: "message", T: GenerateFullName, oneof: "config" },
    { no: 12, name: "generate_gender_config", kind: "message", T: GenerateGender, oneof: "config" },
    { no: 13, name: "generate_int64_phone_config", kind: "message", T: GenerateInt64Phone, oneof: "config" },
    { no: 14, name: "generate_int_config", kind: "message", T: GenerateInt, oneof: "config" },
    { no: 15, name: "generate_last_name_config", kind: "message", T: GenerateLastName, oneof: "config" },
    { no: 16, name: "generate_sha256hash_config", kind: "message", T: GenerateSha256Hash, oneof: "config" },
    { no: 17, name: "generate_ssn_config", kind: "message", T: GenerateSSN, oneof: "config" },
    { no: 18, name: "generate_state_config", kind: "message", T: GenerateState, oneof: "config" },
    { no: 19, name: "generate_street_address_config", kind: "message", T: GenerateStreetAddress, oneof: "config" },
    { no: 20, name: "generate_string_phone_config", kind: "message", T: GenerateStringPhone, oneof: "config" },
    { no: 21, name: "generate_string_config", kind: "message", T: GenerateString, oneof: "config" },
    { no: 22, name: "generate_unixtimestamp_config", kind: "message", T: GenerateUnixTimestamp, oneof: "config" },
    { no: 23, name: "generate_username_config", kind: "message", T: GenerateUsername, oneof: "config" },
    { no: 24, name: "generate_utctimestamp_config", kind: "message", T: GenerateUtcTimestamp, oneof: "config" },
    { no: 25, name: "generate_uuid_config", kind: "message", T: GenerateUuid, oneof: "config" },
    { no: 26, name: "generate_zipcode_config", kind: "message", T: GenerateZipcode, oneof: "config" },
    { no: 27, name: "transform_e164_phone_config", kind: "message", T: TransformE164Phone, oneof: "config" },
    { no: 28, name: "transform_first_name_config", kind: "message", T: TransformFirstName, oneof: "config" },
    { no: 29, name: "transform_float_config", kind: "message", T: TransformFloat, oneof: "config" },
    { no: 30, name: "transform_full_name_config", kind: "message", T: TransformFullName, oneof: "config" },
    { no: 31, name: "transform_int_phone_config", kind: "message", T: TransformIntPhone, oneof: "config" },
    { no: 32, name: "transform_int_config", kind: "message", T: TransformInt, oneof: "config" },
    { no: 33, name: "transform_last_name_config", kind: "message", T: TransformLastName, oneof: "config" },
    { no: 34, name: "transform_phone_config", kind: "message", T: TransformPhone, oneof: "config" },
    { no: 35, name: "transform_string_config", kind: "message", T: TransformString, oneof: "config" },
    { no: 36, name: "passthrough_config", kind: "message", T: Passthrough, oneof: "config" },
    { no: 37, name: "nullconfig", kind: "message", T: Null, oneof: "config" },
    { no: 38, name: "user_defined_transformer_config", kind: "message", T: UserDefinedTransformerConfig, oneof: "config" },
    { no: 39, name: "generate_default_config", kind: "message", T: GenerateDefault, oneof: "config" }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformerConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformerConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformerConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformerConfig, a, b);
  }
};
var GenerateEmail = class _GenerateEmail extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateEmail";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateEmail().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateEmail().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateEmail().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateEmail, a, b);
  }
};
var GenerateRealisticEmail = class _GenerateRealisticEmail extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateRealisticEmail";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateRealisticEmail().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateRealisticEmail().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateRealisticEmail().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateRealisticEmail, a, b);
  }
};
var TransformEmail = class _TransformEmail extends Message3 {
  /**
   * @generated from field: bool preserve_domain = 1;
   */
  preserveDomain = false;
  /**
   * @generated from field: bool preserve_length = 2;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformEmail";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_domain",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    {
      no: 2,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformEmail().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformEmail().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformEmail().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformEmail, a, b);
  }
};
var GenerateBool = class _GenerateBool extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateBool";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateBool().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateBool().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateBool().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateBool, a, b);
  }
};
var GenerateCardNumber = class _GenerateCardNumber extends Message3 {
  /**
   * @generated from field: bool valid_luhn = 1;
   */
  validLuhn = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateCardNumber";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "valid_luhn",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateCardNumber().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateCardNumber().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateCardNumber().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateCardNumber, a, b);
  }
};
var GenerateCity = class _GenerateCity extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateCity";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateCity().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateCity().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateCity().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateCity, a, b);
  }
};
var GenerateDefault = class _GenerateDefault extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateDefault";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateDefault().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateDefault().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateDefault().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateDefault, a, b);
  }
};
var GenerateE164Number = class _GenerateE164Number extends Message3 {
  /**
   * @generated from field: int64 length = 1;
   */
  length = protoInt64.zero;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateE164Number";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "length",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateE164Number().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateE164Number().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateE164Number().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateE164Number, a, b);
  }
};
var GenerateFirstName = class _GenerateFirstName extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateFirstName";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateFirstName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateFirstName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateFirstName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateFirstName, a, b);
  }
};
var GenerateFloat = class _GenerateFloat extends Message3 {
  /**
   * @generated from field: string sign = 1;
   */
  sign = "";
  /**
   * @generated from field: int64 digits_before_decimal = 2;
   */
  digitsBeforeDecimal = protoInt64.zero;
  /**
   * @generated from field: int64 digits_after_decimal = 3;
   */
  digitsAfterDecimal = protoInt64.zero;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateFloat";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "sign",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "digits_before_decimal",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    },
    {
      no: 3,
      name: "digits_after_decimal",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateFloat().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateFloat().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateFloat().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateFloat, a, b);
  }
};
var GenerateFullAddress = class _GenerateFullAddress extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateFullAddress";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateFullAddress().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateFullAddress().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateFullAddress().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateFullAddress, a, b);
  }
};
var GenerateFullName = class _GenerateFullName extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateFullName";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateFullName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateFullName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateFullName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateFullName, a, b);
  }
};
var GenerateGender = class _GenerateGender extends Message3 {
  /**
   * @generated from field: bool abbreviate = 1;
   */
  abbreviate = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateGender";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "abbreviate",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateGender().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateGender().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateGender().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateGender, a, b);
  }
};
var GenerateInt64Phone = class _GenerateInt64Phone extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateInt64Phone";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateInt64Phone().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateInt64Phone().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateInt64Phone().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateInt64Phone, a, b);
  }
};
var GenerateInt = class _GenerateInt extends Message3 {
  /**
   * @generated from field: int64 length = 1;
   */
  length = protoInt64.zero;
  /**
   * @generated from field: string sign = 2;
   */
  sign = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateInt";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "length",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    },
    {
      no: 2,
      name: "sign",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateInt().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateInt().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateInt().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateInt, a, b);
  }
};
var GenerateLastName = class _GenerateLastName extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateLastName";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateLastName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateLastName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateLastName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateLastName, a, b);
  }
};
var GenerateSha256Hash = class _GenerateSha256Hash extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateSha256Hash";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateSha256Hash().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateSha256Hash().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateSha256Hash().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateSha256Hash, a, b);
  }
};
var GenerateSSN = class _GenerateSSN extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateSSN";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateSSN().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateSSN().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateSSN().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateSSN, a, b);
  }
};
var GenerateState = class _GenerateState extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateState";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateState().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateState().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateState().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateState, a, b);
  }
};
var GenerateStreetAddress = class _GenerateStreetAddress extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateStreetAddress";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateStreetAddress().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateStreetAddress().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateStreetAddress().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateStreetAddress, a, b);
  }
};
var GenerateStringPhone = class _GenerateStringPhone extends Message3 {
  /**
   * @generated from field: bool include_hyphens = 2;
   */
  includeHyphens = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateStringPhone";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 2,
      name: "include_hyphens",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateStringPhone().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateStringPhone().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateStringPhone().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateStringPhone, a, b);
  }
};
var GenerateString = class _GenerateString extends Message3 {
  /**
   * @generated from field: int64 length = 1;
   */
  length = protoInt64.zero;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateString";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "length",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateString().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateString().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateString().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateString, a, b);
  }
};
var GenerateUnixTimestamp = class _GenerateUnixTimestamp extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateUnixTimestamp";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateUnixTimestamp().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateUnixTimestamp().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateUnixTimestamp().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateUnixTimestamp, a, b);
  }
};
var GenerateUsername = class _GenerateUsername extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateUsername";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateUsername().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateUsername().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateUsername().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateUsername, a, b);
  }
};
var GenerateUtcTimestamp = class _GenerateUtcTimestamp extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateUtcTimestamp";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateUtcTimestamp().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateUtcTimestamp().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateUtcTimestamp().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateUtcTimestamp, a, b);
  }
};
var GenerateUuid = class _GenerateUuid extends Message3 {
  /**
   * @generated from field: bool include_hyphens = 1;
   */
  includeHyphens = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateUuid";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "include_hyphens",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateUuid().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateUuid().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateUuid().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateUuid, a, b);
  }
};
var GenerateZipcode = class _GenerateZipcode extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.GenerateZipcode";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GenerateZipcode().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateZipcode().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateZipcode().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_GenerateZipcode, a, b);
  }
};
var TransformE164Phone = class _TransformE164Phone extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformE164Phone";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformE164Phone().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformE164Phone().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformE164Phone().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformE164Phone, a, b);
  }
};
var TransformFirstName = class _TransformFirstName extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformFirstName";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformFirstName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformFirstName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformFirstName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformFirstName, a, b);
  }
};
var TransformFloat = class _TransformFloat extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  /**
   * @generated from field: bool preserve_sign = 2;
   */
  preserveSign = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformFloat";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    {
      no: 2,
      name: "preserve_sign",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformFloat().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformFloat().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformFloat().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformFloat, a, b);
  }
};
var TransformFullName = class _TransformFullName extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformFullName";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformFullName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformFullName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformFullName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformFullName, a, b);
  }
};
var TransformIntPhone = class _TransformIntPhone extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformIntPhone";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformIntPhone().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformIntPhone().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformIntPhone().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformIntPhone, a, b);
  }
};
var TransformInt = class _TransformInt extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  /**
   * @generated from field: bool preserve_sign = 2;
   */
  preserveSign = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformInt";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    {
      no: 2,
      name: "preserve_sign",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformInt().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformInt().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformInt().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformInt, a, b);
  }
};
var TransformLastName = class _TransformLastName extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformLastName";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformLastName().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformLastName().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformLastName().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformLastName, a, b);
  }
};
var TransformPhone = class _TransformPhone extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  /**
   * @generated from field: bool include_hyphens = 2;
   */
  includeHyphens = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformPhone";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    {
      no: 2,
      name: "include_hyphens",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformPhone().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformPhone().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformPhone().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformPhone, a, b);
  }
};
var TransformString = class _TransformString extends Message3 {
  /**
   * @generated from field: bool preserve_length = 1;
   */
  preserveLength = false;
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.TransformString";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "preserve_length",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _TransformString().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _TransformString().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _TransformString().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_TransformString, a, b);
  }
};
var Passthrough = class _Passthrough extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.Passthrough";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _Passthrough().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _Passthrough().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _Passthrough().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_Passthrough, a, b);
  }
};
var Null = class _Null extends Message3 {
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.Null";
  static fields = proto33.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _Null().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _Null().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _Null().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_Null, a, b);
  }
};
var UserDefinedTransformerConfig = class _UserDefinedTransformerConfig extends Message3 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto33.util.initPartial(data, this);
  }
  static runtime = proto33;
  static typeName = "mgmt.v1alpha1.UserDefinedTransformerConfig";
  static fields = proto33.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _UserDefinedTransformerConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UserDefinedTransformerConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UserDefinedTransformerConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto33.util.equals(_UserDefinedTransformerConfig, a, b);
  }
};

// src/client/mgmt/v1alpha1/job_pb.ts
var JobStatus = /* @__PURE__ */ ((JobStatus2) => {
  JobStatus2[JobStatus2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  JobStatus2[JobStatus2["ENABLED"] = 1] = "ENABLED";
  JobStatus2[JobStatus2["PAUSED"] = 3] = "PAUSED";
  JobStatus2[JobStatus2["DISABLED"] = 4] = "DISABLED";
  return JobStatus2;
})(JobStatus || {});
proto34.util.setEnumType(JobStatus, "mgmt.v1alpha1.JobStatus", [
  { no: 0, name: "JOB_STATUS_UNSPECIFIED" },
  { no: 1, name: "JOB_STATUS_ENABLED" },
  { no: 3, name: "JOB_STATUS_PAUSED" },
  { no: 4, name: "JOB_STATUS_DISABLED" }
]);
var ActivityStatus = /* @__PURE__ */ ((ActivityStatus2) => {
  ActivityStatus2[ActivityStatus2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  ActivityStatus2[ActivityStatus2["SCHEDULED"] = 1] = "SCHEDULED";
  ActivityStatus2[ActivityStatus2["STARTED"] = 2] = "STARTED";
  ActivityStatus2[ActivityStatus2["CANCELED"] = 3] = "CANCELED";
  ActivityStatus2[ActivityStatus2["FAILED"] = 4] = "FAILED";
  return ActivityStatus2;
})(ActivityStatus || {});
proto34.util.setEnumType(ActivityStatus, "mgmt.v1alpha1.ActivityStatus", [
  { no: 0, name: "ACTIVITY_STATUS_UNSPECIFIED" },
  { no: 1, name: "ACTIVITY_STATUS_SCHEDULED" },
  { no: 2, name: "ACTIVITY_STATUS_STARTED" },
  { no: 3, name: "ACTIVITY_STATUS_CANCELED" },
  { no: 4, name: "ACTIVITY_STATUS_FAILED" }
]);
var JobRunStatus = /* @__PURE__ */ ((JobRunStatus2) => {
  JobRunStatus2[JobRunStatus2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  JobRunStatus2[JobRunStatus2["PENDING"] = 1] = "PENDING";
  JobRunStatus2[JobRunStatus2["RUNNING"] = 2] = "RUNNING";
  JobRunStatus2[JobRunStatus2["COMPLETE"] = 3] = "COMPLETE";
  JobRunStatus2[JobRunStatus2["ERROR"] = 4] = "ERROR";
  JobRunStatus2[JobRunStatus2["CANCELED"] = 5] = "CANCELED";
  JobRunStatus2[JobRunStatus2["TERMINATED"] = 6] = "TERMINATED";
  JobRunStatus2[JobRunStatus2["FAILED"] = 7] = "FAILED";
  return JobRunStatus2;
})(JobRunStatus || {});
proto34.util.setEnumType(JobRunStatus, "mgmt.v1alpha1.JobRunStatus", [
  { no: 0, name: "JOB_RUN_STATUS_UNSPECIFIED" },
  { no: 1, name: "JOB_RUN_STATUS_PENDING" },
  { no: 2, name: "JOB_RUN_STATUS_RUNNING" },
  { no: 3, name: "JOB_RUN_STATUS_COMPLETE" },
  { no: 4, name: "JOB_RUN_STATUS_ERROR" },
  { no: 5, name: "JOB_RUN_STATUS_CANCELED" },
  { no: 6, name: "JOB_RUN_STATUS_TERMINATED" },
  { no: 7, name: "JOB_RUN_STATUS_FAILED" }
]);
var GetJobsRequest = class _GetJobsRequest extends Message4 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobsRequest, a, b);
  }
};
var GetJobsResponse = class _GetJobsResponse extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.Job jobs = 1;
   */
  jobs = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "jobs", kind: "message", T: Job, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobsResponse, a, b);
  }
};
var JobSource = class _JobSource extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.JobSourceOptions options = 1;
   */
  options;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobSource";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "options", kind: "message", T: JobSourceOptions }
  ]);
  static fromBinary(bytes, options) {
    return new _JobSource().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobSource().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobSource().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobSource, a, b);
  }
};
var JobSourceOptions = class _JobSourceOptions extends Message4 {
  /**
   * @generated from oneof mgmt.v1alpha1.JobSourceOptions.config
   */
  config = { case: void 0 };
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobSourceOptions";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "postgres", kind: "message", T: PostgresSourceConnectionOptions, oneof: "config" },
    { no: 2, name: "aws_s3", kind: "message", T: AwsS3SourceConnectionOptions, oneof: "config" },
    { no: 3, name: "mysql", kind: "message", T: MysqlSourceConnectionOptions, oneof: "config" },
    { no: 4, name: "generate", kind: "message", T: GenerateSourceOptions, oneof: "config" }
  ]);
  static fromBinary(bytes, options) {
    return new _JobSourceOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobSourceOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobSourceOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobSourceOptions, a, b);
  }
};
var CreateJobDestination = class _CreateJobDestination extends Message4 {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 2;
   */
  options;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobDestination";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "options", kind: "message", T: JobDestinationOptions }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobDestination().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobDestination().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobDestination().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobDestination, a, b);
  }
};
var JobDestination = class _JobDestination extends Message4 {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 2;
   */
  options;
  /**
   * @generated from field: string id = 3;
   */
  id = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobDestination";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "options", kind: "message", T: JobDestinationOptions },
    {
      no: 3,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _JobDestination().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobDestination().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobDestination().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobDestination, a, b);
  }
};
var GenerateSourceOptions = class _GenerateSourceOptions extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.GenerateSourceSchemaOption schemas = 1;
   */
  schemas = [];
  /**
   * @generated from field: optional string fk_source_connection_id = 3;
   */
  fkSourceConnectionId;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GenerateSourceOptions";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "schemas", kind: "message", T: GenerateSourceSchemaOption, repeated: true },
    { no: 3, name: "fk_source_connection_id", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateSourceOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateSourceOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateSourceOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GenerateSourceOptions, a, b);
  }
};
var GenerateSourceSchemaOption = class _GenerateSourceSchemaOption extends Message4 {
  /**
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * @generated from field: repeated mgmt.v1alpha1.GenerateSourceTableOption tables = 2;
   */
  tables = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GenerateSourceSchemaOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "tables", kind: "message", T: GenerateSourceTableOption, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateSourceSchemaOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateSourceSchemaOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateSourceSchemaOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GenerateSourceSchemaOption, a, b);
  }
};
var GenerateSourceTableOption = class _GenerateSourceTableOption extends Message4 {
  /**
   * @generated from field: string table = 1;
   */
  table = "";
  /**
   * @generated from field: int64 row_count = 2;
   */
  rowCount = protoInt642.zero;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GenerateSourceTableOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "row_count",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GenerateSourceTableOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GenerateSourceTableOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GenerateSourceTableOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GenerateSourceTableOption, a, b);
  }
};
var PostgresSourceConnectionOptions = class _PostgresSourceConnectionOptions extends Message4 {
  /**
   * @generated from field: bool halt_on_new_column_addition = 1;
   */
  haltOnNewColumnAddition = false;
  /**
   * @generated from field: repeated mgmt.v1alpha1.PostgresSourceSchemaOption schemas = 2;
   */
  schemas = [];
  /**
   * @generated from field: string connection_id = 3;
   */
  connectionId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresSourceConnectionOptions";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "halt_on_new_column_addition",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 2, name: "schemas", kind: "message", T: PostgresSourceSchemaOption, repeated: true },
    {
      no: 3,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresSourceConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresSourceConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresSourceConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresSourceConnectionOptions, a, b);
  }
};
var PostgresSourceSchemaOption = class _PostgresSourceSchemaOption extends Message4 {
  /**
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * @generated from field: repeated mgmt.v1alpha1.PostgresSourceTableOption tables = 2;
   */
  tables = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresSourceSchemaOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "tables", kind: "message", T: PostgresSourceTableOption, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresSourceSchemaOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresSourceSchemaOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresSourceSchemaOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresSourceSchemaOption, a, b);
  }
};
var PostgresSourceTableOption = class _PostgresSourceTableOption extends Message4 {
  /**
   * @generated from field: string table = 1;
   */
  table = "";
  /**
   * @generated from field: optional string where_clause = 2;
   */
  whereClause;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresSourceTableOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "where_clause", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresSourceTableOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresSourceTableOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresSourceTableOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresSourceTableOption, a, b);
  }
};
var MysqlSourceConnectionOptions = class _MysqlSourceConnectionOptions extends Message4 {
  /**
   * @generated from field: bool halt_on_new_column_addition = 1;
   */
  haltOnNewColumnAddition = false;
  /**
   * @generated from field: repeated mgmt.v1alpha1.MysqlSourceSchemaOption schemas = 2;
   */
  schemas = [];
  /**
   * @generated from field: string connection_id = 3;
   */
  connectionId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlSourceConnectionOptions";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "halt_on_new_column_addition",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 2, name: "schemas", kind: "message", T: MysqlSourceSchemaOption, repeated: true },
    {
      no: 3,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlSourceConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlSourceConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlSourceConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlSourceConnectionOptions, a, b);
  }
};
var MysqlSourceSchemaOption = class _MysqlSourceSchemaOption extends Message4 {
  /**
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * @generated from field: repeated mgmt.v1alpha1.MysqlSourceTableOption tables = 2;
   */
  tables = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlSourceSchemaOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "tables", kind: "message", T: MysqlSourceTableOption, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlSourceSchemaOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlSourceSchemaOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlSourceSchemaOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlSourceSchemaOption, a, b);
  }
};
var MysqlSourceTableOption = class _MysqlSourceTableOption extends Message4 {
  /**
   * @generated from field: string table = 1;
   */
  table = "";
  /**
   * @generated from field: optional string where_clause = 2;
   */
  whereClause;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlSourceTableOption";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "where_clause", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlSourceTableOption().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlSourceTableOption().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlSourceTableOption().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlSourceTableOption, a, b);
  }
};
var AwsS3SourceConnectionOptions = class _AwsS3SourceConnectionOptions extends Message4 {
  /**
   * @generated from field: string connection_id = 1;
   */
  connectionId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.AwsS3SourceConnectionOptions";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _AwsS3SourceConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AwsS3SourceConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AwsS3SourceConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_AwsS3SourceConnectionOptions, a, b);
  }
};
var JobDestinationOptions = class _JobDestinationOptions extends Message4 {
  /**
   * @generated from oneof mgmt.v1alpha1.JobDestinationOptions.config
   */
  config = { case: void 0 };
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobDestinationOptions";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "postgres_options", kind: "message", T: PostgresDestinationConnectionOptions, oneof: "config" },
    { no: 2, name: "aws_s3_options", kind: "message", T: AwsS3DestinationConnectionOptions, oneof: "config" },
    { no: 3, name: "mysql_options", kind: "message", T: MysqlDestinationConnectionOptions, oneof: "config" }
  ]);
  static fromBinary(bytes, options) {
    return new _JobDestinationOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobDestinationOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobDestinationOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobDestinationOptions, a, b);
  }
};
var PostgresDestinationConnectionOptions = class _PostgresDestinationConnectionOptions extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.PostgresTruncateTableConfig truncate_table = 1;
   */
  truncateTable;
  /**
   * @generated from field: bool init_table_schema = 2;
   */
  initTableSchema = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresDestinationConnectionOptions";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "truncate_table", kind: "message", T: PostgresTruncateTableConfig },
    {
      no: 2,
      name: "init_table_schema",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresDestinationConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresDestinationConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresDestinationConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresDestinationConnectionOptions, a, b);
  }
};
var PostgresTruncateTableConfig = class _PostgresTruncateTableConfig extends Message4 {
  /**
   * @generated from field: bool truncate_before_insert = 1;
   */
  truncateBeforeInsert = false;
  /**
   * @generated from field: bool cascade = 2;
   */
  cascade = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresTruncateTableConfig";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "truncate_before_insert",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    {
      no: 2,
      name: "cascade",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresTruncateTableConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresTruncateTableConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresTruncateTableConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresTruncateTableConfig, a, b);
  }
};
var MysqlDestinationConnectionOptions = class _MysqlDestinationConnectionOptions extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.MysqlTruncateTableConfig truncate_table = 1;
   */
  truncateTable;
  /**
   * @generated from field: bool init_table_schema = 2;
   */
  initTableSchema = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlDestinationConnectionOptions";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "truncate_table", kind: "message", T: MysqlTruncateTableConfig },
    {
      no: 2,
      name: "init_table_schema",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlDestinationConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlDestinationConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlDestinationConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlDestinationConnectionOptions, a, b);
  }
};
var MysqlTruncateTableConfig = class _MysqlTruncateTableConfig extends Message4 {
  /**
   * @generated from field: bool truncate_before_insert = 1;
   */
  truncateBeforeInsert = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlTruncateTableConfig";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "truncate_before_insert",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlTruncateTableConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlTruncateTableConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlTruncateTableConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlTruncateTableConfig, a, b);
  }
};
var AwsS3DestinationConnectionOptions = class _AwsS3DestinationConnectionOptions extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.AwsS3DestinationConnectionOptions";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _AwsS3DestinationConnectionOptions().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AwsS3DestinationConnectionOptions().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AwsS3DestinationConnectionOptions().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_AwsS3DestinationConnectionOptions, a, b);
  }
};
var CreateJobRequest = class _CreateJobRequest extends Message4 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string job_name = 2;
   */
  jobName = "";
  /**
   * @generated from field: optional string cron_schedule = 3;
   */
  cronSchedule;
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 4;
   */
  mappings = [];
  /**
   * @generated from field: mgmt.v1alpha1.JobSource source = 5;
   */
  source;
  /**
   * @generated from field: repeated mgmt.v1alpha1.CreateJobDestination destinations = 6;
   */
  destinations = [];
  /**
   * @generated from field: bool initiate_job_run = 7;
   */
  initiateJobRun = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "job_name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "cron_schedule", kind: "scalar", T: 9, opt: true },
    { no: 4, name: "mappings", kind: "message", T: JobMapping, repeated: true },
    { no: 5, name: "source", kind: "message", T: JobSource },
    { no: 6, name: "destinations", kind: "message", T: CreateJobDestination, repeated: true },
    {
      no: 7,
      name: "initiate_job_run",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobRequest, a, b);
  }
};
var CreateJobResponse = class _CreateJobResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobResponse, a, b);
  }
};
var JobMappingTransformer = class _JobMappingTransformer extends Message4 {
  /**
   * @generated from field: string source = 1;
   */
  source = "";
  /**
   * @generated from field: mgmt.v1alpha1.TransformerConfig config = 3;
   */
  config;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobMappingTransformer";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "source",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "config", kind: "message", T: TransformerConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _JobMappingTransformer().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobMappingTransformer().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobMappingTransformer().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobMappingTransformer, a, b);
  }
};
var JobMapping = class _JobMapping extends Message4 {
  /**
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * @generated from field: string table = 2;
   */
  table = "";
  /**
   * @generated from field: string column = 3;
   */
  column = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobMappingTransformer transformer = 5;
   */
  transformer;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobMapping";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "column",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 5, name: "transformer", kind: "message", T: JobMappingTransformer }
  ]);
  static fromBinary(bytes, options) {
    return new _JobMapping().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobMapping().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobMapping().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobMapping, a, b);
  }
};
var GetJobRequest = class _GetJobRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRequest, a, b);
  }
};
var GetJobResponse = class _GetJobResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobResponse, a, b);
  }
};
var UpdateJobScheduleRequest = class _UpdateJobScheduleRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: optional string cron_schedule = 2;
   */
  cronSchedule;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobScheduleRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "cron_schedule", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobScheduleRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobScheduleRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobScheduleRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobScheduleRequest, a, b);
  }
};
var UpdateJobScheduleResponse = class _UpdateJobScheduleResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobScheduleResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobScheduleResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobScheduleResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobScheduleResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobScheduleResponse, a, b);
  }
};
var PauseJobRequest = class _PauseJobRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: bool pause = 2;
   */
  pause = false;
  /**
   * @generated from field: optional string note = 3;
   */
  note;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PauseJobRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "pause",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 3, name: "note", kind: "scalar", T: 9, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PauseJobRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PauseJobRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PauseJobRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PauseJobRequest, a, b);
  }
};
var PauseJobResponse = class _PauseJobResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PauseJobResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _PauseJobResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PauseJobResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PauseJobResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PauseJobResponse, a, b);
  }
};
var UpdateJobSourceConnectionRequest = class _UpdateJobSourceConnectionRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobSource source = 2;
   */
  source;
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 3;
   */
  mappings = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobSourceConnectionRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "source", kind: "message", T: JobSource },
    { no: 3, name: "mappings", kind: "message", T: JobMapping, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobSourceConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobSourceConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobSourceConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobSourceConnectionRequest, a, b);
  }
};
var UpdateJobSourceConnectionResponse = class _UpdateJobSourceConnectionResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobSourceConnectionResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobSourceConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobSourceConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobSourceConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobSourceConnectionResponse, a, b);
  }
};
var PostgresSourceSchemaSubset = class _PostgresSourceSchemaSubset extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.PostgresSourceSchemaOption postgres_schemas = 1;
   */
  postgresSchemas = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PostgresSourceSchemaSubset";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "postgres_schemas", kind: "message", T: PostgresSourceSchemaOption, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PostgresSourceSchemaSubset().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PostgresSourceSchemaSubset().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PostgresSourceSchemaSubset().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PostgresSourceSchemaSubset, a, b);
  }
};
var MysqlSourceSchemaSubset = class _MysqlSourceSchemaSubset extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.MysqlSourceSchemaOption mysql_schemas = 1;
   */
  mysqlSchemas = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.MysqlSourceSchemaSubset";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "mysql_schemas", kind: "message", T: MysqlSourceSchemaOption, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _MysqlSourceSchemaSubset().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _MysqlSourceSchemaSubset().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _MysqlSourceSchemaSubset().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_MysqlSourceSchemaSubset, a, b);
  }
};
var JobSourceSqlSubetSchemas = class _JobSourceSqlSubetSchemas extends Message4 {
  /**
   * @generated from oneof mgmt.v1alpha1.JobSourceSqlSubetSchemas.schemas
   */
  schemas = { case: void 0 };
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobSourceSqlSubetSchemas";
  static fields = proto34.util.newFieldList(() => [
    { no: 2, name: "postgres_subset", kind: "message", T: PostgresSourceSchemaSubset, oneof: "schemas" },
    { no: 3, name: "mysql_subset", kind: "message", T: MysqlSourceSchemaSubset, oneof: "schemas" }
  ]);
  static fromBinary(bytes, options) {
    return new _JobSourceSqlSubetSchemas().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobSourceSqlSubetSchemas().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobSourceSqlSubetSchemas().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobSourceSqlSubetSchemas, a, b);
  }
};
var SetJobSourceSqlConnectionSubsetsRequest = class _SetJobSourceSqlConnectionSubsetsRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobSourceSqlSubetSchemas schemas = 2;
   */
  schemas;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "schemas", kind: "message", T: JobSourceSqlSubetSchemas }
  ]);
  static fromBinary(bytes, options) {
    return new _SetJobSourceSqlConnectionSubsetsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetJobSourceSqlConnectionSubsetsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetJobSourceSqlConnectionSubsetsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_SetJobSourceSqlConnectionSubsetsRequest, a, b);
  }
};
var SetJobSourceSqlConnectionSubsetsResponse = class _SetJobSourceSqlConnectionSubsetsResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _SetJobSourceSqlConnectionSubsetsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetJobSourceSqlConnectionSubsetsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetJobSourceSqlConnectionSubsetsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_SetJobSourceSqlConnectionSubsetsResponse, a, b);
  }
};
var UpdateJobDestinationConnectionRequest = class _UpdateJobDestinationConnectionRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  /**
   * @generated from field: string connection_id = 2;
   */
  connectionId = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 3;
   */
  options;
  /**
   * @generated from field: string destination_id = 4;
   */
  destinationId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobDestinationConnectionRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "connection_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "options", kind: "message", T: JobDestinationOptions },
    {
      no: 4,
      name: "destination_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobDestinationConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobDestinationConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobDestinationConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobDestinationConnectionRequest, a, b);
  }
};
var UpdateJobDestinationConnectionResponse = class _UpdateJobDestinationConnectionResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.UpdateJobDestinationConnectionResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _UpdateJobDestinationConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UpdateJobDestinationConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UpdateJobDestinationConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_UpdateJobDestinationConnectionResponse, a, b);
  }
};
var DeleteJobDestinationConnectionRequest = class _DeleteJobDestinationConnectionRequest extends Message4 {
  /**
   * @generated from field: string destination_id = 1;
   */
  destinationId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobDestinationConnectionRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "destination_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteJobDestinationConnectionRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobDestinationConnectionRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobDestinationConnectionRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobDestinationConnectionRequest, a, b);
  }
};
var DeleteJobDestinationConnectionResponse = class _DeleteJobDestinationConnectionResponse extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobDestinationConnectionResponse";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteJobDestinationConnectionResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobDestinationConnectionResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobDestinationConnectionResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobDestinationConnectionResponse, a, b);
  }
};
var CreateJobDestinationConnectionsRequest = class _CreateJobDestinationConnectionsRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  /**
   * @generated from field: repeated mgmt.v1alpha1.CreateJobDestination destinations = 2;
   */
  destinations = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobDestinationConnectionsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "destinations", kind: "message", T: CreateJobDestination, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobDestinationConnectionsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobDestinationConnectionsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobDestinationConnectionsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobDestinationConnectionsRequest, a, b);
  }
};
var CreateJobDestinationConnectionsResponse = class _CreateJobDestinationConnectionsResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.Job job = 1;
   */
  job;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobDestinationConnectionsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job", kind: "message", T: Job }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobDestinationConnectionsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobDestinationConnectionsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobDestinationConnectionsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobDestinationConnectionsResponse, a, b);
  }
};
var DeleteJobRequest = class _DeleteJobRequest extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteJobRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobRequest, a, b);
  }
};
var DeleteJobResponse = class _DeleteJobResponse extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobResponse";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteJobResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobResponse, a, b);
  }
};
var IsJobNameAvailableRequest = class _IsJobNameAvailableRequest extends Message4 {
  /**
   * @generated from field: string name = 1;
   */
  name = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.IsJobNameAvailableRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsJobNameAvailableRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsJobNameAvailableRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsJobNameAvailableRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_IsJobNameAvailableRequest, a, b);
  }
};
var IsJobNameAvailableResponse = class _IsJobNameAvailableResponse extends Message4 {
  /**
   * @generated from field: bool is_available = 1;
   */
  isAvailable = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.IsJobNameAvailableResponse";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "is_available",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsJobNameAvailableResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsJobNameAvailableResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsJobNameAvailableResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_IsJobNameAvailableResponse, a, b);
  }
};
var GetJobRunsRequest = class _GetJobRunsRequest extends Message4 {
  /**
   * @generated from oneof mgmt.v1alpha1.GetJobRunsRequest.id
   */
  id = { case: void 0 };
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunsRequest";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job_id", kind: "scalar", T: 9, oneof: "id" },
    { no: 2, name: "account_id", kind: "scalar", T: 9, oneof: "id" }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunsRequest, a, b);
  }
};
var GetJobRunsResponse = class _GetJobRunsResponse extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobRun job_runs = 1;
   */
  jobRuns = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job_runs", kind: "message", T: JobRun, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunsResponse, a, b);
  }
};
var GetJobRunRequest = class _GetJobRunRequest extends Message4 {
  /**
   * @generated from field: string job_run_id = 1;
   */
  jobRunId = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_run_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunRequest, a, b);
  }
};
var GetJobRunResponse = class _GetJobRunResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.JobRun job_run = 1;
   */
  jobRun;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "job_run", kind: "message", T: JobRun }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunResponse, a, b);
  }
};
var CreateJobRunRequest = class _CreateJobRunRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobRunRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateJobRunRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobRunRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobRunRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobRunRequest, a, b);
  }
};
var CreateJobRunResponse = class _CreateJobRunResponse extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CreateJobRunResponse";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _CreateJobRunResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateJobRunResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateJobRunResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CreateJobRunResponse, a, b);
  }
};
var CancelJobRunRequest = class _CancelJobRunRequest extends Message4 {
  /**
   * @generated from field: string job_run_id = 1;
   */
  jobRunId = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CancelJobRunRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_run_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CancelJobRunRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CancelJobRunRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CancelJobRunRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CancelJobRunRequest, a, b);
  }
};
var CancelJobRunResponse = class _CancelJobRunResponse extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.CancelJobRunResponse";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _CancelJobRunResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CancelJobRunResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CancelJobRunResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_CancelJobRunResponse, a, b);
  }
};
var Job = class _Job extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string created_by_user_id = 2;
   */
  createdByUserId = "";
  /**
   * @generated from field: google.protobuf.Timestamp created_at = 3;
   */
  createdAt;
  /**
   * @generated from field: string updated_by_user_id = 4;
   */
  updatedByUserId = "";
  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 5;
   */
  updatedAt;
  /**
   * @generated from field: string name = 6;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobSource source = 7;
   */
  source;
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobDestination destinations = 8;
   */
  destinations = [];
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 9;
   */
  mappings = [];
  /**
   * @generated from field: optional string cron_schedule = 10;
   */
  cronSchedule;
  /**
   * @generated from field: string account_id = 11;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.Job";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "created_by_user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "created_at", kind: "message", T: Timestamp4 },
    {
      no: 4,
      name: "updated_by_user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 5, name: "updated_at", kind: "message", T: Timestamp4 },
    {
      no: 6,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 7, name: "source", kind: "message", T: JobSource },
    { no: 8, name: "destinations", kind: "message", T: JobDestination, repeated: true },
    { no: 9, name: "mappings", kind: "message", T: JobMapping, repeated: true },
    { no: 10, name: "cron_schedule", kind: "scalar", T: 9, opt: true },
    {
      no: 11,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _Job().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _Job().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _Job().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_Job, a, b);
  }
};
var JobRecentRun = class _JobRecentRun extends Message4 {
  /**
   * @generated from field: google.protobuf.Timestamp start_time = 1;
   */
  startTime;
  /**
   * @generated from field: string job_run_id = 2;
   */
  jobRunId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRecentRun";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "start_time", kind: "message", T: Timestamp4 },
    {
      no: 2,
      name: "job_run_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRecentRun().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRecentRun().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRecentRun().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRecentRun, a, b);
  }
};
var GetJobRecentRunsRequest = class _GetJobRecentRunsRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRecentRunsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRecentRunsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRecentRunsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRecentRunsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRecentRunsRequest, a, b);
  }
};
var GetJobRecentRunsResponse = class _GetJobRecentRunsResponse extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobRecentRun recent_runs = 1;
   */
  recentRuns = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRecentRunsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "recent_runs", kind: "message", T: JobRecentRun, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRecentRunsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRecentRunsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRecentRunsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRecentRunsResponse, a, b);
  }
};
var JobNextRuns = class _JobNextRuns extends Message4 {
  /**
   * @generated from field: repeated google.protobuf.Timestamp next_run_times = 1;
   */
  nextRunTimes = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobNextRuns";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "next_run_times", kind: "message", T: Timestamp4, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _JobNextRuns().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobNextRuns().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobNextRuns().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobNextRuns, a, b);
  }
};
var GetJobNextRunsRequest = class _GetJobNextRunsRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobNextRunsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobNextRunsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobNextRunsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobNextRunsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobNextRunsRequest, a, b);
  }
};
var GetJobNextRunsResponse = class _GetJobNextRunsResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.JobNextRuns next_runs = 1;
   */
  nextRuns;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobNextRunsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "next_runs", kind: "message", T: JobNextRuns }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobNextRunsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobNextRunsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobNextRunsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobNextRunsResponse, a, b);
  }
};
var GetJobStatusRequest = class _GetJobStatusRequest extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobStatusRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobStatusRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobStatusRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobStatusRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobStatusRequest, a, b);
  }
};
var GetJobStatusResponse = class _GetJobStatusResponse extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.JobStatus status = 1;
   */
  status = 0 /* UNSPECIFIED */;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobStatusResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "status", kind: "enum", T: proto34.getEnumType(JobStatus) }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobStatusResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobStatusResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobStatusResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobStatusResponse, a, b);
  }
};
var JobStatusRecord = class _JobStatusRecord extends Message4 {
  /**
   * @generated from field: string job_id = 1;
   */
  jobId = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobStatus status = 2;
   */
  status = 0 /* UNSPECIFIED */;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobStatusRecord";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "status", kind: "enum", T: proto34.getEnumType(JobStatus) }
  ]);
  static fromBinary(bytes, options) {
    return new _JobStatusRecord().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobStatusRecord().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobStatusRecord().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobStatusRecord, a, b);
  }
};
var GetJobStatusesRequest = class _GetJobStatusesRequest extends Message4 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobStatusesRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobStatusesRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobStatusesRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobStatusesRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobStatusesRequest, a, b);
  }
};
var GetJobStatusesResponse = class _GetJobStatusesResponse extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobStatusRecord statuses = 1;
   */
  statuses = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobStatusesResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "statuses", kind: "message", T: JobStatusRecord, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobStatusesResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobStatusesResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobStatusesResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobStatusesResponse, a, b);
  }
};
var ActivityFailure = class _ActivityFailure extends Message4 {
  /**
   * @generated from field: string message = 1;
   */
  message = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.ActivityFailure";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "message",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _ActivityFailure().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _ActivityFailure().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _ActivityFailure().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_ActivityFailure, a, b);
  }
};
var PendingActivity = class _PendingActivity extends Message4 {
  /**
   * @generated from field: mgmt.v1alpha1.ActivityStatus status = 1;
   */
  status = 0 /* UNSPECIFIED */;
  /**
   * @generated from field: string activity_name = 2;
   */
  activityName = "";
  /**
   * @generated from field: optional mgmt.v1alpha1.ActivityFailure last_failure = 3;
   */
  lastFailure;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.PendingActivity";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "status", kind: "enum", T: proto34.getEnumType(ActivityStatus) },
    {
      no: 2,
      name: "activity_name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "last_failure", kind: "message", T: ActivityFailure, opt: true }
  ]);
  static fromBinary(bytes, options) {
    return new _PendingActivity().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _PendingActivity().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _PendingActivity().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_PendingActivity, a, b);
  }
};
var JobRun = class _JobRun extends Message4 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string job_id = 2;
   */
  jobId = "";
  /**
   * @generated from field: string name = 3;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.JobRunStatus status = 4;
   */
  status = 0 /* UNSPECIFIED */;
  /**
   * @generated from field: google.protobuf.Timestamp started_at = 6;
   */
  startedAt;
  /**
   * @generated from field: optional google.protobuf.Timestamp completed_at = 7;
   */
  completedAt;
  /**
   * @generated from field: repeated mgmt.v1alpha1.PendingActivity pending_activities = 8;
   */
  pendingActivities = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRun";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "job_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 4, name: "status", kind: "enum", T: proto34.getEnumType(JobRunStatus) },
    { no: 6, name: "started_at", kind: "message", T: Timestamp4 },
    { no: 7, name: "completed_at", kind: "message", T: Timestamp4, opt: true },
    { no: 8, name: "pending_activities", kind: "message", T: PendingActivity, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRun().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRun().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRun().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRun, a, b);
  }
};
var JobRunEventTaskError = class _JobRunEventTaskError extends Message4 {
  /**
   * @generated from field: string message = 1;
   */
  message = "";
  /**
   * @generated from field: string retry_state = 2;
   */
  retryState = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRunEventTaskError";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "message",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "retry_state",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRunEventTaskError().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRunEventTaskError().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRunEventTaskError().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRunEventTaskError, a, b);
  }
};
var JobRunEventTask = class _JobRunEventTask extends Message4 {
  /**
   * @generated from field: int64 id = 1;
   */
  id = protoInt642.zero;
  /**
   * @generated from field: string type = 2;
   */
  type = "";
  /**
   * @generated from field: google.protobuf.Timestamp event_time = 3;
   */
  eventTime;
  /**
   * @generated from field: mgmt.v1alpha1.JobRunEventTaskError error = 4;
   */
  error;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRunEventTask";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    },
    {
      no: 2,
      name: "type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "event_time", kind: "message", T: Timestamp4 },
    { no: 4, name: "error", kind: "message", T: JobRunEventTaskError }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRunEventTask().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRunEventTask().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRunEventTask().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRunEventTask, a, b);
  }
};
var JobRunSyncMetadata = class _JobRunSyncMetadata extends Message4 {
  /**
   * @generated from field: string schema = 1;
   */
  schema = "";
  /**
   * @generated from field: string table = 2;
   */
  table = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRunSyncMetadata";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "schema",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "table",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRunSyncMetadata().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRunSyncMetadata().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRunSyncMetadata().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRunSyncMetadata, a, b);
  }
};
var JobRunEventMetadata = class _JobRunEventMetadata extends Message4 {
  /**
   * @generated from oneof mgmt.v1alpha1.JobRunEventMetadata.metadata
   */
  metadata = { case: void 0 };
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRunEventMetadata";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "sync_metadata", kind: "message", T: JobRunSyncMetadata, oneof: "metadata" }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRunEventMetadata().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRunEventMetadata().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRunEventMetadata().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRunEventMetadata, a, b);
  }
};
var JobRunEvent = class _JobRunEvent extends Message4 {
  /**
   * @generated from field: int64 id = 1;
   */
  id = protoInt642.zero;
  /**
   * @generated from field: string type = 2;
   */
  type = "";
  /**
   * @generated from field: google.protobuf.Timestamp start_time = 3;
   */
  startTime;
  /**
   * @generated from field: google.protobuf.Timestamp close_time = 4;
   */
  closeTime;
  /**
   * @generated from field: mgmt.v1alpha1.JobRunEventMetadata metadata = 5;
   */
  metadata;
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobRunEventTask tasks = 6;
   */
  tasks = [];
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.JobRunEvent";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 3
      /* ScalarType.INT64 */
    },
    {
      no: 2,
      name: "type",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "start_time", kind: "message", T: Timestamp4 },
    { no: 4, name: "close_time", kind: "message", T: Timestamp4 },
    { no: 5, name: "metadata", kind: "message", T: JobRunEventMetadata },
    { no: 6, name: "tasks", kind: "message", T: JobRunEventTask, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _JobRunEvent().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _JobRunEvent().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _JobRunEvent().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_JobRunEvent, a, b);
  }
};
var GetJobRunEventsRequest = class _GetJobRunEventsRequest extends Message4 {
  /**
   * @generated from field: string job_run_id = 1;
   */
  jobRunId = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunEventsRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_run_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunEventsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunEventsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunEventsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunEventsRequest, a, b);
  }
};
var GetJobRunEventsResponse = class _GetJobRunEventsResponse extends Message4 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.JobRunEvent events = 1;
   */
  events = [];
  /**
   * @generated from field: bool is_run_complete = 2;
   */
  isRunComplete = false;
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.GetJobRunEventsResponse";
  static fields = proto34.util.newFieldList(() => [
    { no: 1, name: "events", kind: "message", T: JobRunEvent, repeated: true },
    {
      no: 2,
      name: "is_run_complete",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetJobRunEventsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetJobRunEventsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetJobRunEventsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_GetJobRunEventsResponse, a, b);
  }
};
var DeleteJobRunRequest = class _DeleteJobRunRequest extends Message4 {
  /**
   * @generated from field: string job_run_id = 1;
   */
  jobRunId = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobRunRequest";
  static fields = proto34.util.newFieldList(() => [
    {
      no: 1,
      name: "job_run_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _DeleteJobRunRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobRunRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobRunRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobRunRequest, a, b);
  }
};
var DeleteJobRunResponse = class _DeleteJobRunResponse extends Message4 {
  constructor(data) {
    super();
    proto34.util.initPartial(data, this);
  }
  static runtime = proto34;
  static typeName = "mgmt.v1alpha1.DeleteJobRunResponse";
  static fields = proto34.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _DeleteJobRunResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _DeleteJobRunResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _DeleteJobRunResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto34.util.equals(_DeleteJobRunResponse, a, b);
  }
};

// src/client/mgmt/v1alpha1/job_connect.ts
import { MethodKind as MethodKind3 } from "@bufbuild/protobuf";
var JobService = {
  typeName: "mgmt.v1alpha1.JobService",
  methods: {
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobs
     */
    getJobs: {
      name: "GetJobs",
      I: GetJobsRequest,
      O: GetJobsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJob
     */
    getJob: {
      name: "GetJob",
      I: GetJobRequest,
      O: GetJobResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.CreateJob
     */
    createJob: {
      name: "CreateJob",
      I: CreateJobRequest,
      O: CreateJobResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.DeleteJob
     */
    deleteJob: {
      name: "DeleteJob",
      I: DeleteJobRequest,
      O: DeleteJobResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.IsJobNameAvailable
     */
    isJobNameAvailable: {
      name: "IsJobNameAvailable",
      I: IsJobNameAvailableRequest,
      O: IsJobNameAvailableResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobSchedule
     */
    updateJobSchedule: {
      name: "UpdateJobSchedule",
      I: UpdateJobScheduleRequest,
      O: UpdateJobScheduleResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobSourceConnection
     */
    updateJobSourceConnection: {
      name: "UpdateJobSourceConnection",
      I: UpdateJobSourceConnectionRequest,
      O: UpdateJobSourceConnectionResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.SetJobSourceSqlConnectionSubsets
     */
    setJobSourceSqlConnectionSubsets: {
      name: "SetJobSourceSqlConnectionSubsets",
      I: SetJobSourceSqlConnectionSubsetsRequest,
      O: SetJobSourceSqlConnectionSubsetsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobDestinationConnection
     */
    updateJobDestinationConnection: {
      name: "UpdateJobDestinationConnection",
      I: UpdateJobDestinationConnectionRequest,
      O: UpdateJobDestinationConnectionResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.DeleteJobDestinationConnection
     */
    deleteJobDestinationConnection: {
      name: "DeleteJobDestinationConnection",
      I: DeleteJobDestinationConnectionRequest,
      O: DeleteJobDestinationConnectionResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.CreateJobDestinationConnections
     */
    createJobDestinationConnections: {
      name: "CreateJobDestinationConnections",
      I: CreateJobDestinationConnectionsRequest,
      O: CreateJobDestinationConnectionsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.PauseJob
     */
    pauseJob: {
      name: "PauseJob",
      I: PauseJobRequest,
      O: PauseJobResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobRecentRuns
     */
    getJobRecentRuns: {
      name: "GetJobRecentRuns",
      I: GetJobRecentRunsRequest,
      O: GetJobRecentRunsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobNextRuns
     */
    getJobNextRuns: {
      name: "GetJobNextRuns",
      I: GetJobNextRunsRequest,
      O: GetJobNextRunsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobStatus
     */
    getJobStatus: {
      name: "GetJobStatus",
      I: GetJobStatusRequest,
      O: GetJobStatusResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobStatuses
     */
    getJobStatuses: {
      name: "GetJobStatuses",
      I: GetJobStatusesRequest,
      O: GetJobStatusesResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobRuns
     */
    getJobRuns: {
      name: "GetJobRuns",
      I: GetJobRunsRequest,
      O: GetJobRunsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobRunEvents
     */
    getJobRunEvents: {
      name: "GetJobRunEvents",
      I: GetJobRunEventsRequest,
      O: GetJobRunEventsResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.GetJobRun
     */
    getJobRun: {
      name: "GetJobRun",
      I: GetJobRunRequest,
      O: GetJobRunResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.DeleteJobRun
     */
    deleteJobRun: {
      name: "DeleteJobRun",
      I: DeleteJobRunRequest,
      O: DeleteJobRunResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.CreateJobRun
     */
    createJobRun: {
      name: "CreateJobRun",
      I: CreateJobRunRequest,
      O: CreateJobRunResponse,
      kind: MethodKind3.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.JobService.CancelJobRun
     */
    cancelJobRun: {
      name: "CancelJobRun",
      I: CancelJobRunRequest,
      O: CancelJobRunResponse,
      kind: MethodKind3.Unary
    }
  }
};

// src/client/mgmt/v1alpha1/transformer_connect.ts
import { MethodKind as MethodKind4 } from "@bufbuild/protobuf";
var TransformersService = {
  typeName: "mgmt.v1alpha1.TransformersService",
  methods: {
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.GetSystemTransformers
     */
    getSystemTransformers: {
      name: "GetSystemTransformers",
      I: GetSystemTransformersRequest,
      O: GetSystemTransformersResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.GetUserDefinedTransformers
     */
    getUserDefinedTransformers: {
      name: "GetUserDefinedTransformers",
      I: GetUserDefinedTransformersRequest,
      O: GetUserDefinedTransformersResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.GetUserDefinedTransformerById
     */
    getUserDefinedTransformerById: {
      name: "GetUserDefinedTransformerById",
      I: GetUserDefinedTransformerByIdRequest,
      O: GetUserDefinedTransformerByIdResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.CreateUserDefinedTransformer
     */
    createUserDefinedTransformer: {
      name: "CreateUserDefinedTransformer",
      I: CreateUserDefinedTransformerRequest,
      O: CreateUserDefinedTransformerResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.DeleteUserDefinedTransformer
     */
    deleteUserDefinedTransformer: {
      name: "DeleteUserDefinedTransformer",
      I: DeleteUserDefinedTransformerRequest,
      O: DeleteUserDefinedTransformerResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.UpdateUserDefinedTransformer
     */
    updateUserDefinedTransformer: {
      name: "UpdateUserDefinedTransformer",
      I: UpdateUserDefinedTransformerRequest,
      O: UpdateUserDefinedTransformerResponse,
      kind: MethodKind4.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.TransformersService.IsTransformerNameAvailable
     */
    isTransformerNameAvailable: {
      name: "IsTransformerNameAvailable",
      I: IsTransformerNameAvailableRequest,
      O: IsTransformerNameAvailableResponse,
      kind: MethodKind4.Unary
    }
  }
};

// src/client/mgmt/v1alpha1/user_account_pb.ts
import { Message as Message5, proto3 as proto35, Timestamp as Timestamp5 } from "@bufbuild/protobuf";
var UserAccountType = /* @__PURE__ */ ((UserAccountType2) => {
  UserAccountType2[UserAccountType2["UNSPECIFIED"] = 0] = "UNSPECIFIED";
  UserAccountType2[UserAccountType2["PERSONAL"] = 1] = "PERSONAL";
  UserAccountType2[UserAccountType2["TEAM"] = 2] = "TEAM";
  return UserAccountType2;
})(UserAccountType || {});
proto35.util.setEnumType(UserAccountType, "mgmt.v1alpha1.UserAccountType", [
  { no: 0, name: "USER_ACCOUNT_TYPE_UNSPECIFIED" },
  { no: 1, name: "USER_ACCOUNT_TYPE_PERSONAL" },
  { no: 2, name: "USER_ACCOUNT_TYPE_TEAM" }
]);
var GetUserRequest = class _GetUserRequest extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetUserRequest";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GetUserRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetUserRequest, a, b);
  }
};
var GetUserResponse = class _GetUserResponse extends Message5 {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetUserResponse";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetUserResponse, a, b);
  }
};
var SetUserRequest = class _SetUserRequest extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetUserRequest";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _SetUserRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetUserRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetUserRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetUserRequest, a, b);
  }
};
var SetUserResponse = class _SetUserResponse extends Message5 {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetUserResponse";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _SetUserResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetUserResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetUserResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetUserResponse, a, b);
  }
};
var GetUserAccountsRequest = class _GetUserAccountsRequest extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetUserAccountsRequest";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _GetUserAccountsRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserAccountsRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserAccountsRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetUserAccountsRequest, a, b);
  }
};
var GetUserAccountsResponse = class _GetUserAccountsResponse extends Message5 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.UserAccount accounts = 1;
   */
  accounts = [];
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetUserAccountsResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "accounts", kind: "message", T: UserAccount, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetUserAccountsResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetUserAccountsResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetUserAccountsResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetUserAccountsResponse, a, b);
  }
};
var UserAccount = class _UserAccount extends Message5 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: mgmt.v1alpha1.UserAccountType type = 3;
   */
  type = 0 /* UNSPECIFIED */;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.UserAccount";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 3, name: "type", kind: "enum", T: proto35.getEnumType(UserAccountType) }
  ]);
  static fromBinary(bytes, options) {
    return new _UserAccount().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _UserAccount().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _UserAccount().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_UserAccount, a, b);
  }
};
var ConvertPersonalToTeamAccountRequest = class _ConvertPersonalToTeamAccountRequest extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.ConvertPersonalToTeamAccountRequest";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _ConvertPersonalToTeamAccountRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _ConvertPersonalToTeamAccountRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _ConvertPersonalToTeamAccountRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_ConvertPersonalToTeamAccountRequest, a, b);
  }
};
var ConvertPersonalToTeamAccountResponse = class _ConvertPersonalToTeamAccountResponse extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.ConvertPersonalToTeamAccountResponse";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _ConvertPersonalToTeamAccountResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _ConvertPersonalToTeamAccountResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _ConvertPersonalToTeamAccountResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_ConvertPersonalToTeamAccountResponse, a, b);
  }
};
var SetPersonalAccountRequest = class _SetPersonalAccountRequest extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetPersonalAccountRequest";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _SetPersonalAccountRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetPersonalAccountRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetPersonalAccountRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetPersonalAccountRequest, a, b);
  }
};
var SetPersonalAccountResponse = class _SetPersonalAccountResponse extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetPersonalAccountResponse";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _SetPersonalAccountResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetPersonalAccountResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetPersonalAccountResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetPersonalAccountResponse, a, b);
  }
};
var IsUserInAccountRequest = class _IsUserInAccountRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.IsUserInAccountRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsUserInAccountRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsUserInAccountRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsUserInAccountRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_IsUserInAccountRequest, a, b);
  }
};
var IsUserInAccountResponse = class _IsUserInAccountResponse extends Message5 {
  /**
   * @generated from field: bool ok = 1;
   */
  ok = false;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.IsUserInAccountResponse";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "ok",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _IsUserInAccountResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _IsUserInAccountResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _IsUserInAccountResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_IsUserInAccountResponse, a, b);
  }
};
var GetAccountTemporalConfigRequest = class _GetAccountTemporalConfigRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetAccountTemporalConfigRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountTemporalConfigRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountTemporalConfigRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountTemporalConfigRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetAccountTemporalConfigRequest, a, b);
  }
};
var GetAccountTemporalConfigResponse = class _GetAccountTemporalConfigResponse extends Message5 {
  /**
   * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 1;
   */
  config;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetAccountTemporalConfigResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: AccountTemporalConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _GetAccountTemporalConfigResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetAccountTemporalConfigResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetAccountTemporalConfigResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetAccountTemporalConfigResponse, a, b);
  }
};
var SetAccountTemporalConfigRequest = class _SetAccountTemporalConfigRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 2;
   */
  config;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetAccountTemporalConfigRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    { no: 2, name: "config", kind: "message", T: AccountTemporalConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _SetAccountTemporalConfigRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetAccountTemporalConfigRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetAccountTemporalConfigRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetAccountTemporalConfigRequest, a, b);
  }
};
var SetAccountTemporalConfigResponse = class _SetAccountTemporalConfigResponse extends Message5 {
  /**
   * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 1;
   */
  config;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.SetAccountTemporalConfigResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "config", kind: "message", T: AccountTemporalConfig }
  ]);
  static fromBinary(bytes, options) {
    return new _SetAccountTemporalConfigResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _SetAccountTemporalConfigResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _SetAccountTemporalConfigResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_SetAccountTemporalConfigResponse, a, b);
  }
};
var AccountTemporalConfig = class _AccountTemporalConfig extends Message5 {
  /**
   * @generated from field: string url = 1;
   */
  url = "";
  /**
   * @generated from field: string namespace = 2;
   */
  namespace = "";
  /**
   * @generated from field: string sync_job_queue_name = 3;
   */
  syncJobQueueName = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.AccountTemporalConfig";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "url",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "namespace",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "sync_job_queue_name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _AccountTemporalConfig().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AccountTemporalConfig().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AccountTemporalConfig().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_AccountTemporalConfig, a, b);
  }
};
var CreateTeamAccountRequest = class _CreateTeamAccountRequest extends Message5 {
  /**
   * @generated from field: string name = 1;
   */
  name = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.CreateTeamAccountRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateTeamAccountRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateTeamAccountRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateTeamAccountRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_CreateTeamAccountRequest, a, b);
  }
};
var CreateTeamAccountResponse = class _CreateTeamAccountResponse extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.CreateTeamAccountResponse";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _CreateTeamAccountResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _CreateTeamAccountResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _CreateTeamAccountResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_CreateTeamAccountResponse, a, b);
  }
};
var AccountUser = class _AccountUser extends Message5 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string name = 2;
   */
  name = "";
  /**
   * @generated from field: string image = 3;
   */
  image = "";
  /**
   * @generated from field: string email = 4;
   */
  email = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.AccountUser";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "name",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "image",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "email",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _AccountUser().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AccountUser().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AccountUser().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_AccountUser, a, b);
  }
};
var GetTeamAccountMembersRequest = class _GetTeamAccountMembersRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetTeamAccountMembersRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetTeamAccountMembersRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetTeamAccountMembersRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetTeamAccountMembersRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetTeamAccountMembersRequest, a, b);
  }
};
var GetTeamAccountMembersResponse = class _GetTeamAccountMembersResponse extends Message5 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.AccountUser users = 1;
   */
  users = [];
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetTeamAccountMembersResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "users", kind: "message", T: AccountUser, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetTeamAccountMembersResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetTeamAccountMembersResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetTeamAccountMembersResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetTeamAccountMembersResponse, a, b);
  }
};
var RemoveTeamAccountMemberRequest = class _RemoveTeamAccountMemberRequest extends Message5 {
  /**
   * @generated from field: string user_id = 1;
   */
  userId = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.RemoveTeamAccountMemberRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _RemoveTeamAccountMemberRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RemoveTeamAccountMemberRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RemoveTeamAccountMemberRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_RemoveTeamAccountMemberRequest, a, b);
  }
};
var RemoveTeamAccountMemberResponse = class _RemoveTeamAccountMemberResponse extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.RemoveTeamAccountMemberResponse";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _RemoveTeamAccountMemberResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RemoveTeamAccountMemberResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RemoveTeamAccountMemberResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_RemoveTeamAccountMemberResponse, a, b);
  }
};
var InviteUserToTeamAccountRequest = class _InviteUserToTeamAccountRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  /**
   * @generated from field: string email = 2;
   */
  email = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.InviteUserToTeamAccountRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "email",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _InviteUserToTeamAccountRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _InviteUserToTeamAccountRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _InviteUserToTeamAccountRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_InviteUserToTeamAccountRequest, a, b);
  }
};
var AccountInvite = class _AccountInvite extends Message5 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  /**
   * @generated from field: string account_id = 2;
   */
  accountId = "";
  /**
   * @generated from field: string sender_user_id = 3;
   */
  senderUserId = "";
  /**
   * @generated from field: string email = 4;
   */
  email = "";
  /**
   * @generated from field: string token = 5;
   */
  token = "";
  /**
   * @generated from field: bool accepted = 6;
   */
  accepted = false;
  /**
   * @generated from field: google.protobuf.Timestamp created_at = 7;
   */
  createdAt;
  /**
   * @generated from field: google.protobuf.Timestamp updated_at = 8;
   */
  updatedAt;
  /**
   * @generated from field: google.protobuf.Timestamp expires_at = 9;
   */
  expiresAt;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.AccountInvite";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 2,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 3,
      name: "sender_user_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 4,
      name: "email",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 5,
      name: "token",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    },
    {
      no: 6,
      name: "accepted",
      kind: "scalar",
      T: 8
      /* ScalarType.BOOL */
    },
    { no: 7, name: "created_at", kind: "message", T: Timestamp5 },
    { no: 8, name: "updated_at", kind: "message", T: Timestamp5 },
    { no: 9, name: "expires_at", kind: "message", T: Timestamp5 }
  ]);
  static fromBinary(bytes, options) {
    return new _AccountInvite().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AccountInvite().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AccountInvite().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_AccountInvite, a, b);
  }
};
var InviteUserToTeamAccountResponse = class _InviteUserToTeamAccountResponse extends Message5 {
  /**
   * @generated from field: mgmt.v1alpha1.AccountInvite invite = 1;
   */
  invite;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.InviteUserToTeamAccountResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "invite", kind: "message", T: AccountInvite }
  ]);
  static fromBinary(bytes, options) {
    return new _InviteUserToTeamAccountResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _InviteUserToTeamAccountResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _InviteUserToTeamAccountResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_InviteUserToTeamAccountResponse, a, b);
  }
};
var GetTeamAccountInvitesRequest = class _GetTeamAccountInvitesRequest extends Message5 {
  /**
   * @generated from field: string account_id = 1;
   */
  accountId = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetTeamAccountInvitesRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "account_id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _GetTeamAccountInvitesRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetTeamAccountInvitesRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetTeamAccountInvitesRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetTeamAccountInvitesRequest, a, b);
  }
};
var GetTeamAccountInvitesResponse = class _GetTeamAccountInvitesResponse extends Message5 {
  /**
   * @generated from field: repeated mgmt.v1alpha1.AccountInvite invites = 1;
   */
  invites = [];
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.GetTeamAccountInvitesResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "invites", kind: "message", T: AccountInvite, repeated: true }
  ]);
  static fromBinary(bytes, options) {
    return new _GetTeamAccountInvitesResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _GetTeamAccountInvitesResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _GetTeamAccountInvitesResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_GetTeamAccountInvitesResponse, a, b);
  }
};
var RemoveTeamAccountInviteRequest = class _RemoveTeamAccountInviteRequest extends Message5 {
  /**
   * @generated from field: string id = 1;
   */
  id = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.RemoveTeamAccountInviteRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "id",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _RemoveTeamAccountInviteRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RemoveTeamAccountInviteRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RemoveTeamAccountInviteRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_RemoveTeamAccountInviteRequest, a, b);
  }
};
var RemoveTeamAccountInviteResponse = class _RemoveTeamAccountInviteResponse extends Message5 {
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.RemoveTeamAccountInviteResponse";
  static fields = proto35.util.newFieldList(() => []);
  static fromBinary(bytes, options) {
    return new _RemoveTeamAccountInviteResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _RemoveTeamAccountInviteResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _RemoveTeamAccountInviteResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_RemoveTeamAccountInviteResponse, a, b);
  }
};
var AcceptTeamAccountInviteRequest = class _AcceptTeamAccountInviteRequest extends Message5 {
  /**
   * @generated from field: string token = 1;
   */
  token = "";
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.AcceptTeamAccountInviteRequest";
  static fields = proto35.util.newFieldList(() => [
    {
      no: 1,
      name: "token",
      kind: "scalar",
      T: 9
      /* ScalarType.STRING */
    }
  ]);
  static fromBinary(bytes, options) {
    return new _AcceptTeamAccountInviteRequest().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AcceptTeamAccountInviteRequest().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AcceptTeamAccountInviteRequest().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_AcceptTeamAccountInviteRequest, a, b);
  }
};
var AcceptTeamAccountInviteResponse = class _AcceptTeamAccountInviteResponse extends Message5 {
  /**
   * @generated from field: mgmt.v1alpha1.UserAccount account = 1;
   */
  account;
  constructor(data) {
    super();
    proto35.util.initPartial(data, this);
  }
  static runtime = proto35;
  static typeName = "mgmt.v1alpha1.AcceptTeamAccountInviteResponse";
  static fields = proto35.util.newFieldList(() => [
    { no: 1, name: "account", kind: "message", T: UserAccount }
  ]);
  static fromBinary(bytes, options) {
    return new _AcceptTeamAccountInviteResponse().fromBinary(bytes, options);
  }
  static fromJson(jsonValue, options) {
    return new _AcceptTeamAccountInviteResponse().fromJson(jsonValue, options);
  }
  static fromJsonString(jsonString, options) {
    return new _AcceptTeamAccountInviteResponse().fromJsonString(jsonString, options);
  }
  static equals(a, b) {
    return proto35.util.equals(_AcceptTeamAccountInviteResponse, a, b);
  }
};

// src/client/mgmt/v1alpha1/user_account_connect.ts
import { MethodKind as MethodKind5 } from "@bufbuild/protobuf";
var UserAccountService = {
  typeName: "mgmt.v1alpha1.UserAccountService",
  methods: {
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUser
     */
    getUser: {
      name: "GetUser",
      I: GetUserRequest,
      O: GetUserResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.SetUser
     */
    setUser: {
      name: "SetUser",
      I: SetUserRequest,
      O: SetUserResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUserAccounts
     */
    getUserAccounts: {
      name: "GetUserAccounts",
      I: GetUserAccountsRequest,
      O: GetUserAccountsResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.SetPersonalAccount
     */
    setPersonalAccount: {
      name: "SetPersonalAccount",
      I: SetPersonalAccountRequest,
      O: SetPersonalAccountResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.ConvertPersonalToTeamAccount
     */
    convertPersonalToTeamAccount: {
      name: "ConvertPersonalToTeamAccount",
      I: ConvertPersonalToTeamAccountRequest,
      O: ConvertPersonalToTeamAccountResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.CreateTeamAccount
     */
    createTeamAccount: {
      name: "CreateTeamAccount",
      I: CreateTeamAccountRequest,
      O: CreateTeamAccountResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.IsUserInAccount
     */
    isUserInAccount: {
      name: "IsUserInAccount",
      I: IsUserInAccountRequest,
      O: IsUserInAccountResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountTemporalConfig
     */
    getAccountTemporalConfig: {
      name: "GetAccountTemporalConfig",
      I: GetAccountTemporalConfigRequest,
      O: GetAccountTemporalConfigResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.SetAccountTemporalConfig
     */
    setAccountTemporalConfig: {
      name: "SetAccountTemporalConfig",
      I: SetAccountTemporalConfigRequest,
      O: SetAccountTemporalConfigResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountMembers
     */
    getTeamAccountMembers: {
      name: "GetTeamAccountMembers",
      I: GetTeamAccountMembersRequest,
      O: GetTeamAccountMembersResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountMember
     */
    removeTeamAccountMember: {
      name: "RemoveTeamAccountMember",
      I: RemoveTeamAccountMemberRequest,
      O: RemoveTeamAccountMemberResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.InviteUserToTeamAccount
     */
    inviteUserToTeamAccount: {
      name: "InviteUserToTeamAccount",
      I: InviteUserToTeamAccountRequest,
      O: InviteUserToTeamAccountResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountInvites
     */
    getTeamAccountInvites: {
      name: "GetTeamAccountInvites",
      I: GetTeamAccountInvitesRequest,
      O: GetTeamAccountInvitesResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountInvite
     */
    removeTeamAccountInvite: {
      name: "RemoveTeamAccountInvite",
      I: RemoveTeamAccountInviteRequest,
      O: RemoveTeamAccountInviteResponse,
      kind: MethodKind5.Unary
    },
    /**
     * @generated from rpc mgmt.v1alpha1.UserAccountService.AcceptTeamAccountInvite
     */
    acceptTeamAccountInvite: {
      name: "AcceptTeamAccountInvite",
      I: AcceptTeamAccountInviteRequest,
      O: AcceptTeamAccountInviteResponse,
      kind: MethodKind5.Unary
    }
  }
};

// src/client/client.ts
import {
  createPromiseClient
} from "@connectrpc/connect";
function getNeosyncClient(config, _version) {
  return getNeosyncV1alpha1Client(config);
}
function getNeosyncV1alpha1Client(config) {
  const interceptors = config.getAccessToken ? [getAuthInterceptor(config.getAccessToken)] : [];
  const transport = config.getTransport(interceptors);
  return {
    connections: createPromiseClient(ConnectionService, transport),
    users: createPromiseClient(UserAccountService, transport),
    jobs: createPromiseClient(JobService, transport),
    transformers: createPromiseClient(TransformersService, transport),
    apikeys: createPromiseClient(ApiKeyService, transport)
  };
}
function getAuthInterceptor(getAccessToken) {
  return (next) => async (req) => {
    const accessToken = await getAccessToken();
    req.header.set("Authorization", `Bearer ${accessToken}`);
    return next(req);
  };
}
export {
  AcceptTeamAccountInviteRequest,
  AcceptTeamAccountInviteResponse,
  AccountApiKey,
  AccountInvite,
  AccountTemporalConfig,
  AccountUser,
  ActivityFailure,
  ActivityStatus,
  ApiKeyService,
  AwsS3ConnectionConfig,
  AwsS3Credentials,
  AwsS3DestinationConnectionOptions,
  AwsS3SourceConnectionOptions,
  CancelJobRunRequest,
  CancelJobRunResponse,
  CheckConnectionConfigRequest,
  CheckConnectionConfigResponse,
  CheckSqlQueryRequest,
  CheckSqlQueryResponse,
  Code,
  ConnectError,
  Connection,
  ConnectionConfig,
  ConnectionService,
  ConvertPersonalToTeamAccountRequest,
  ConvertPersonalToTeamAccountResponse,
  CreateAccountApiKeyRequest,
  CreateAccountApiKeyResponse,
  CreateConnectionRequest,
  CreateConnectionResponse,
  CreateJobDestination,
  CreateJobDestinationConnectionsRequest,
  CreateJobDestinationConnectionsResponse,
  CreateJobRequest,
  CreateJobResponse,
  CreateJobRunRequest,
  CreateJobRunResponse,
  CreateTeamAccountRequest,
  CreateTeamAccountResponse,
  CreateUserDefinedTransformerRequest,
  CreateUserDefinedTransformerResponse,
  DatabaseColumn,
  DeleteAccountApiKeyRequest,
  DeleteAccountApiKeyResponse,
  DeleteConnectionRequest,
  DeleteConnectionResponse,
  DeleteJobDestinationConnectionRequest,
  DeleteJobDestinationConnectionResponse,
  DeleteJobRequest,
  DeleteJobResponse,
  DeleteJobRunRequest,
  DeleteJobRunResponse,
  DeleteUserDefinedTransformerRequest,
  DeleteUserDefinedTransformerResponse,
  ForeignConstraintTables,
  GenerateBool,
  GenerateCardNumber,
  GenerateCity,
  GenerateDefault,
  GenerateE164Number,
  GenerateEmail,
  GenerateFirstName,
  GenerateFloat,
  GenerateFullAddress,
  GenerateFullName,
  GenerateGender,
  GenerateInt,
  GenerateInt64Phone,
  GenerateLastName,
  GenerateRealisticEmail,
  GenerateSSN,
  GenerateSha256Hash,
  GenerateSourceOptions,
  GenerateSourceSchemaOption,
  GenerateSourceTableOption,
  GenerateState,
  GenerateStreetAddress,
  GenerateString,
  GenerateStringPhone,
  GenerateUnixTimestamp,
  GenerateUsername,
  GenerateUtcTimestamp,
  GenerateUuid,
  GenerateZipcode,
  GetAccountApiKeyRequest,
  GetAccountApiKeyResponse,
  GetAccountApiKeysRequest,
  GetAccountApiKeysResponse,
  GetAccountTemporalConfigRequest,
  GetAccountTemporalConfigResponse,
  GetConnectionDataStreamRequest,
  GetConnectionDataStreamResponse,
  GetConnectionForeignConstraintsRequest,
  GetConnectionForeignConstraintsResponse,
  GetConnectionRequest,
  GetConnectionResponse,
  GetConnectionSchemaRequest,
  GetConnectionSchemaResponse,
  GetConnectionsRequest,
  GetConnectionsResponse,
  GetJobNextRunsRequest,
  GetJobNextRunsResponse,
  GetJobRecentRunsRequest,
  GetJobRecentRunsResponse,
  GetJobRequest,
  GetJobResponse,
  GetJobRunEventsRequest,
  GetJobRunEventsResponse,
  GetJobRunRequest,
  GetJobRunResponse,
  GetJobRunsRequest,
  GetJobRunsResponse,
  GetJobStatusRequest,
  GetJobStatusResponse,
  GetJobStatusesRequest,
  GetJobStatusesResponse,
  GetJobsRequest,
  GetJobsResponse,
  GetSystemTransformersRequest,
  GetSystemTransformersResponse,
  GetTeamAccountInvitesRequest,
  GetTeamAccountInvitesResponse,
  GetTeamAccountMembersRequest,
  GetTeamAccountMembersResponse,
  GetUserAccountsRequest,
  GetUserAccountsResponse,
  GetUserDefinedTransformerByIdRequest,
  GetUserDefinedTransformerByIdResponse,
  GetUserDefinedTransformersRequest,
  GetUserDefinedTransformersResponse,
  GetUserRequest,
  GetUserResponse,
  InviteUserToTeamAccountRequest,
  InviteUserToTeamAccountResponse,
  IsConnectionNameAvailableRequest,
  IsConnectionNameAvailableResponse,
  IsJobNameAvailableRequest,
  IsJobNameAvailableResponse,
  IsTransformerNameAvailableRequest,
  IsTransformerNameAvailableResponse,
  IsUserInAccountRequest,
  IsUserInAccountResponse,
  Job,
  JobDestination,
  JobDestinationOptions,
  JobMapping,
  JobMappingTransformer,
  JobNextRuns,
  JobRecentRun,
  JobRun,
  JobRunEvent,
  JobRunEventMetadata,
  JobRunEventTask,
  JobRunEventTaskError,
  JobRunStatus,
  JobRunSyncMetadata,
  JobService,
  JobSource,
  JobSourceOptions,
  JobSourceSqlSubetSchemas,
  JobStatus,
  JobStatusRecord,
  MysqlConnection,
  MysqlConnectionConfig,
  MysqlDestinationConnectionOptions,
  MysqlSourceConnectionOptions,
  MysqlSourceSchemaOption,
  MysqlSourceSchemaSubset,
  MysqlSourceTableOption,
  MysqlTruncateTableConfig,
  Null,
  Passthrough,
  PauseJobRequest,
  PauseJobResponse,
  PendingActivity,
  PostgresConnection,
  PostgresConnectionConfig,
  PostgresDestinationConnectionOptions,
  PostgresSourceConnectionOptions,
  PostgresSourceSchemaOption,
  PostgresSourceSchemaSubset,
  PostgresSourceTableOption,
  PostgresTruncateTableConfig,
  RegenerateAccountApiKeyRequest,
  RegenerateAccountApiKeyResponse,
  RemoveTeamAccountInviteRequest,
  RemoveTeamAccountInviteResponse,
  RemoveTeamAccountMemberRequest,
  RemoveTeamAccountMemberResponse,
  SetAccountTemporalConfigRequest,
  SetAccountTemporalConfigResponse,
  SetJobSourceSqlConnectionSubsetsRequest,
  SetJobSourceSqlConnectionSubsetsResponse,
  SetPersonalAccountRequest,
  SetPersonalAccountResponse,
  SetUserRequest,
  SetUserResponse,
  SystemTransformer,
  TransformE164Phone,
  TransformEmail,
  TransformFirstName,
  TransformFloat,
  TransformFullName,
  TransformInt,
  TransformIntPhone,
  TransformLastName,
  TransformPhone,
  TransformString,
  TransformerConfig,
  TransformersService,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
  UpdateJobDestinationConnectionRequest,
  UpdateJobDestinationConnectionResponse,
  UpdateJobScheduleRequest,
  UpdateJobScheduleResponse,
  UpdateJobSourceConnectionRequest,
  UpdateJobSourceConnectionResponse,
  UpdateUserDefinedTransformerRequest,
  UpdateUserDefinedTransformerResponse,
  UserAccount,
  UserAccountService,
  UserAccountType,
  UserDefinedTransformer,
  UserDefinedTransformerConfig,
  getNeosyncClient,
  getNeosyncV1alpha1Client
};
