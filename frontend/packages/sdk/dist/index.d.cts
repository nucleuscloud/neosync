import { PromiseClient, Interceptor, Transport } from '@connectrpc/connect';
export { Code, ConnectError, PromiseClient } from '@connectrpc/connect';
import { Message, Timestamp, PartialMessage, proto3, FieldList, BinaryReadOptions, JsonValue, JsonReadOptions, PlainMessage, MethodKind } from '@bufbuild/protobuf';

/**
 * @generated from message mgmt.v1alpha1.CreateAccountApiKeyRequest
 */
declare class CreateAccountApiKeyRequest extends Message<CreateAccountApiKeyRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * Validate between now and one year: now < x < 365 days
     *
     * @generated from field: google.protobuf.Timestamp expires_at = 3;
     */
    expiresAt?: Timestamp;
    constructor(data?: PartialMessage<CreateAccountApiKeyRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateAccountApiKeyRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateAccountApiKeyRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateAccountApiKeyRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateAccountApiKeyRequest;
    static equals(a: CreateAccountApiKeyRequest | PlainMessage<CreateAccountApiKeyRequest> | undefined, b: CreateAccountApiKeyRequest | PlainMessage<CreateAccountApiKeyRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateAccountApiKeyResponse
 */
declare class CreateAccountApiKeyResponse extends Message<CreateAccountApiKeyResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
     */
    apiKey?: AccountApiKey;
    constructor(data?: PartialMessage<CreateAccountApiKeyResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateAccountApiKeyResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateAccountApiKeyResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateAccountApiKeyResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateAccountApiKeyResponse;
    static equals(a: CreateAccountApiKeyResponse | PlainMessage<CreateAccountApiKeyResponse> | undefined, b: CreateAccountApiKeyResponse | PlainMessage<CreateAccountApiKeyResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AccountApiKey
 */
declare class AccountApiKey extends Message<AccountApiKey> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * The friendly name of the API Key
     *
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string account_id = 3;
     */
    accountId: string;
    /**
     * @generated from field: string created_by_id = 4;
     */
    createdById: string;
    /**
     * @generated from field: google.protobuf.Timestamp created_at = 5;
     */
    createdAt?: Timestamp;
    /**
     * @generated from field: string updated_by_id = 6;
     */
    updatedById: string;
    /**
     * @generated from field: google.protobuf.Timestamp updated_at = 7;
     */
    updatedAt?: Timestamp;
    /**
     * key_value is only returned on initial creation or when it is regenerated
     *
     * @generated from field: optional string key_value = 8;
     */
    keyValue?: string;
    /**
     * @generated from field: string user_id = 9;
     */
    userId: string;
    /**
     * The timestamp of what the API key expires and will not longer be usable.
     *
     * @generated from field: google.protobuf.Timestamp expires_at = 10;
     */
    expiresAt?: Timestamp;
    constructor(data?: PartialMessage<AccountApiKey>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AccountApiKey";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AccountApiKey;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AccountApiKey;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AccountApiKey;
    static equals(a: AccountApiKey | PlainMessage<AccountApiKey> | undefined, b: AccountApiKey | PlainMessage<AccountApiKey> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountApiKeysRequest
 */
declare class GetAccountApiKeysRequest extends Message<GetAccountApiKeysRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetAccountApiKeysRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountApiKeysRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountApiKeysRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountApiKeysRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountApiKeysRequest;
    static equals(a: GetAccountApiKeysRequest | PlainMessage<GetAccountApiKeysRequest> | undefined, b: GetAccountApiKeysRequest | PlainMessage<GetAccountApiKeysRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountApiKeysResponse
 */
declare class GetAccountApiKeysResponse extends Message<GetAccountApiKeysResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.AccountApiKey api_keys = 1;
     */
    apiKeys: AccountApiKey[];
    constructor(data?: PartialMessage<GetAccountApiKeysResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountApiKeysResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountApiKeysResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountApiKeysResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountApiKeysResponse;
    static equals(a: GetAccountApiKeysResponse | PlainMessage<GetAccountApiKeysResponse> | undefined, b: GetAccountApiKeysResponse | PlainMessage<GetAccountApiKeysResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountApiKeyRequest
 */
declare class GetAccountApiKeyRequest extends Message<GetAccountApiKeyRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<GetAccountApiKeyRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountApiKeyRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountApiKeyRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountApiKeyRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountApiKeyRequest;
    static equals(a: GetAccountApiKeyRequest | PlainMessage<GetAccountApiKeyRequest> | undefined, b: GetAccountApiKeyRequest | PlainMessage<GetAccountApiKeyRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountApiKeyResponse
 */
declare class GetAccountApiKeyResponse extends Message<GetAccountApiKeyResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
     */
    apiKey?: AccountApiKey;
    constructor(data?: PartialMessage<GetAccountApiKeyResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountApiKeyResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountApiKeyResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountApiKeyResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountApiKeyResponse;
    static equals(a: GetAccountApiKeyResponse | PlainMessage<GetAccountApiKeyResponse> | undefined, b: GetAccountApiKeyResponse | PlainMessage<GetAccountApiKeyResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RegenerateAccountApiKeyRequest
 */
declare class RegenerateAccountApiKeyRequest extends Message<RegenerateAccountApiKeyRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * Validate between now and one year: now < x < 365 days
     *
     * @generated from field: google.protobuf.Timestamp expires_at = 2;
     */
    expiresAt?: Timestamp;
    constructor(data?: PartialMessage<RegenerateAccountApiKeyRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RegenerateAccountApiKeyRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RegenerateAccountApiKeyRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RegenerateAccountApiKeyRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RegenerateAccountApiKeyRequest;
    static equals(a: RegenerateAccountApiKeyRequest | PlainMessage<RegenerateAccountApiKeyRequest> | undefined, b: RegenerateAccountApiKeyRequest | PlainMessage<RegenerateAccountApiKeyRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RegenerateAccountApiKeyResponse
 */
declare class RegenerateAccountApiKeyResponse extends Message<RegenerateAccountApiKeyResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountApiKey api_key = 1;
     */
    apiKey?: AccountApiKey;
    constructor(data?: PartialMessage<RegenerateAccountApiKeyResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RegenerateAccountApiKeyResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RegenerateAccountApiKeyResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RegenerateAccountApiKeyResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RegenerateAccountApiKeyResponse;
    static equals(a: RegenerateAccountApiKeyResponse | PlainMessage<RegenerateAccountApiKeyResponse> | undefined, b: RegenerateAccountApiKeyResponse | PlainMessage<RegenerateAccountApiKeyResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteAccountApiKeyRequest
 */
declare class DeleteAccountApiKeyRequest extends Message<DeleteAccountApiKeyRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<DeleteAccountApiKeyRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteAccountApiKeyRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteAccountApiKeyRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteAccountApiKeyRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteAccountApiKeyRequest;
    static equals(a: DeleteAccountApiKeyRequest | PlainMessage<DeleteAccountApiKeyRequest> | undefined, b: DeleteAccountApiKeyRequest | PlainMessage<DeleteAccountApiKeyRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteAccountApiKeyResponse
 */
declare class DeleteAccountApiKeyResponse extends Message<DeleteAccountApiKeyResponse> {
    constructor(data?: PartialMessage<DeleteAccountApiKeyResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteAccountApiKeyResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteAccountApiKeyResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteAccountApiKeyResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteAccountApiKeyResponse;
    static equals(a: DeleteAccountApiKeyResponse | PlainMessage<DeleteAccountApiKeyResponse> | undefined, b: DeleteAccountApiKeyResponse | PlainMessage<DeleteAccountApiKeyResponse> | undefined): boolean;
}

/**
 * Service that manages the lifecycle of API Keys that are associated with a specific Account.
 *
 * @generated from service mgmt.v1alpha1.ApiKeyService
 */
declare const ApiKeyService: {
    readonly typeName: "mgmt.v1alpha1.ApiKeyService";
    readonly methods: {
        /**
         * Retrieves a list of Account API Keys
         *
         * @generated from rpc mgmt.v1alpha1.ApiKeyService.GetAccountApiKeys
         */
        readonly getAccountApiKeys: {
            readonly name: "GetAccountApiKeys";
            readonly I: typeof GetAccountApiKeysRequest;
            readonly O: typeof GetAccountApiKeysResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Retrieves a single API Key
         *
         * @generated from rpc mgmt.v1alpha1.ApiKeyService.GetAccountApiKey
         */
        readonly getAccountApiKey: {
            readonly name: "GetAccountApiKey";
            readonly I: typeof GetAccountApiKeyRequest;
            readonly O: typeof GetAccountApiKeyResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Creates a single API Key
         * This method will return the decrypted contents of the API key
         *
         * @generated from rpc mgmt.v1alpha1.ApiKeyService.CreateAccountApiKey
         */
        readonly createAccountApiKey: {
            readonly name: "CreateAccountApiKey";
            readonly I: typeof CreateAccountApiKeyRequest;
            readonly O: typeof CreateAccountApiKeyResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Regenerates a single API Key with a new expiration time
         * This method will return the decrypted contents of the API key
         *
         * @generated from rpc mgmt.v1alpha1.ApiKeyService.RegenerateAccountApiKey
         */
        readonly regenerateAccountApiKey: {
            readonly name: "RegenerateAccountApiKey";
            readonly I: typeof RegenerateAccountApiKeyRequest;
            readonly O: typeof RegenerateAccountApiKeyResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Deletes an API Key from the system.
         *
         * @generated from rpc mgmt.v1alpha1.ApiKeyService.DeleteAccountApiKey
         */
        readonly deleteAccountApiKey: {
            readonly name: "DeleteAccountApiKey";
            readonly I: typeof DeleteAccountApiKeyRequest;
            readonly O: typeof DeleteAccountApiKeyResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};

/**
 * @generated from message mgmt.v1alpha1.GetConnectionsRequest
 */
declare class GetConnectionsRequest extends Message<GetConnectionsRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetConnectionsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionsRequest;
    static equals(a: GetConnectionsRequest | PlainMessage<GetConnectionsRequest> | undefined, b: GetConnectionsRequest | PlainMessage<GetConnectionsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionsResponse
 */
declare class GetConnectionsResponse extends Message<GetConnectionsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.Connection connections = 1;
     */
    connections: Connection[];
    constructor(data?: PartialMessage<GetConnectionsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionsResponse;
    static equals(a: GetConnectionsResponse | PlainMessage<GetConnectionsResponse> | undefined, b: GetConnectionsResponse | PlainMessage<GetConnectionsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionRequest
 */
declare class GetConnectionRequest extends Message<GetConnectionRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<GetConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionRequest;
    static equals(a: GetConnectionRequest | PlainMessage<GetConnectionRequest> | undefined, b: GetConnectionRequest | PlainMessage<GetConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionResponse
 */
declare class GetConnectionResponse extends Message<GetConnectionResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Connection connection = 1;
     */
    connection?: Connection;
    constructor(data?: PartialMessage<GetConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionResponse;
    static equals(a: GetConnectionResponse | PlainMessage<GetConnectionResponse> | undefined, b: GetConnectionResponse | PlainMessage<GetConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateConnectionRequest
 */
declare class CreateConnectionRequest extends Message<CreateConnectionRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * The friendly name of the connection
     *
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
     */
    connectionConfig?: ConnectionConfig;
    constructor(data?: PartialMessage<CreateConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateConnectionRequest;
    static equals(a: CreateConnectionRequest | PlainMessage<CreateConnectionRequest> | undefined, b: CreateConnectionRequest | PlainMessage<CreateConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateConnectionResponse
 */
declare class CreateConnectionResponse extends Message<CreateConnectionResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Connection connection = 1;
     */
    connection?: Connection;
    constructor(data?: PartialMessage<CreateConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateConnectionResponse;
    static equals(a: CreateConnectionResponse | PlainMessage<CreateConnectionResponse> | undefined, b: CreateConnectionResponse | PlainMessage<CreateConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateConnectionRequest
 */
declare class UpdateConnectionRequest extends Message<UpdateConnectionRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
     */
    connectionConfig?: ConnectionConfig;
    constructor(data?: PartialMessage<UpdateConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateConnectionRequest;
    static equals(a: UpdateConnectionRequest | PlainMessage<UpdateConnectionRequest> | undefined, b: UpdateConnectionRequest | PlainMessage<UpdateConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateConnectionResponse
 */
declare class UpdateConnectionResponse extends Message<UpdateConnectionResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Connection connection = 1;
     */
    connection?: Connection;
    constructor(data?: PartialMessage<UpdateConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateConnectionResponse;
    static equals(a: UpdateConnectionResponse | PlainMessage<UpdateConnectionResponse> | undefined, b: UpdateConnectionResponse | PlainMessage<UpdateConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteConnectionRequest
 */
declare class DeleteConnectionRequest extends Message<DeleteConnectionRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<DeleteConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteConnectionRequest;
    static equals(a: DeleteConnectionRequest | PlainMessage<DeleteConnectionRequest> | undefined, b: DeleteConnectionRequest | PlainMessage<DeleteConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteConnectionResponse
 */
declare class DeleteConnectionResponse extends Message<DeleteConnectionResponse> {
    constructor(data?: PartialMessage<DeleteConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteConnectionResponse;
    static equals(a: DeleteConnectionResponse | PlainMessage<DeleteConnectionResponse> | undefined, b: DeleteConnectionResponse | PlainMessage<DeleteConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CheckConnectionConfigRequest
 */
declare class CheckConnectionConfigRequest extends Message<CheckConnectionConfigRequest> {
    /**
     * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 1;
     */
    connectionConfig?: ConnectionConfig;
    constructor(data?: PartialMessage<CheckConnectionConfigRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CheckConnectionConfigRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CheckConnectionConfigRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CheckConnectionConfigRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CheckConnectionConfigRequest;
    static equals(a: CheckConnectionConfigRequest | PlainMessage<CheckConnectionConfigRequest> | undefined, b: CheckConnectionConfigRequest | PlainMessage<CheckConnectionConfigRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CheckConnectionConfigResponse
 */
declare class CheckConnectionConfigResponse extends Message<CheckConnectionConfigResponse> {
    /**
     * Whether or not the API was able to ping the connection
     *
     * @generated from field: bool is_connected = 1;
     */
    isConnected: boolean;
    /**
     * This is the error that was received if the API was unable to connect
     *
     * @generated from field: optional string connection_error = 2;
     */
    connectionError?: string;
    constructor(data?: PartialMessage<CheckConnectionConfigResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CheckConnectionConfigResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CheckConnectionConfigResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CheckConnectionConfigResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CheckConnectionConfigResponse;
    static equals(a: CheckConnectionConfigResponse | PlainMessage<CheckConnectionConfigResponse> | undefined, b: CheckConnectionConfigResponse | PlainMessage<CheckConnectionConfigResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.Connection
 */
declare class Connection extends Message<Connection> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.ConnectionConfig connection_config = 3;
     */
    connectionConfig?: ConnectionConfig;
    /**
     * @generated from field: string created_by_user_id = 4;
     */
    createdByUserId: string;
    /**
     * @generated from field: google.protobuf.Timestamp created_at = 5;
     */
    createdAt?: Timestamp;
    /**
     * @generated from field: string updated_by_user_id = 6;
     */
    updatedByUserId: string;
    /**
     * @generated from field: google.protobuf.Timestamp updated_at = 7;
     */
    updatedAt?: Timestamp;
    /**
     * @generated from field: string account_id = 8;
     */
    accountId: string;
    constructor(data?: PartialMessage<Connection>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.Connection";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Connection;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Connection;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Connection;
    static equals(a: Connection | PlainMessage<Connection> | undefined, b: Connection | PlainMessage<Connection> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.ConnectionConfig
 */
declare class ConnectionConfig extends Message<ConnectionConfig> {
    /**
     * @generated from oneof mgmt.v1alpha1.ConnectionConfig.config
     */
    config: {
        /**
         * @generated from field: mgmt.v1alpha1.PostgresConnectionConfig pg_config = 1;
         */
        value: PostgresConnectionConfig;
        case: "pgConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.AwsS3ConnectionConfig aws_s3_config = 2;
         */
        value: AwsS3ConnectionConfig;
        case: "awsS3Config";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.MysqlConnectionConfig mysql_config = 3;
         */
        value: MysqlConnectionConfig;
        case: "mysqlConfig";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<ConnectionConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.ConnectionConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConnectionConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConnectionConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConnectionConfig;
    static equals(a: ConnectionConfig | PlainMessage<ConnectionConfig> | undefined, b: ConnectionConfig | PlainMessage<ConnectionConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresConnectionConfig
 */
declare class PostgresConnectionConfig extends Message<PostgresConnectionConfig> {
    /**
     * May provide either a raw string url, or a structured version
     *
     * @generated from oneof mgmt.v1alpha1.PostgresConnectionConfig.connection_config
     */
    connectionConfig: {
        /**
         * @generated from field: string url = 1;
         */
        value: string;
        case: "url";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.PostgresConnection connection = 2;
         */
        value: PostgresConnection;
        case: "connection";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<PostgresConnectionConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresConnectionConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresConnectionConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresConnectionConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresConnectionConfig;
    static equals(a: PostgresConnectionConfig | PlainMessage<PostgresConnectionConfig> | undefined, b: PostgresConnectionConfig | PlainMessage<PostgresConnectionConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresConnection
 */
declare class PostgresConnection extends Message<PostgresConnection> {
    /**
     * @generated from field: string host = 1;
     */
    host: string;
    /**
     * @generated from field: int32 port = 2;
     */
    port: number;
    /**
     * @generated from field: string name = 3;
     */
    name: string;
    /**
     * @generated from field: string user = 4;
     */
    user: string;
    /**
     * @generated from field: string pass = 5;
     */
    pass: string;
    /**
     * @generated from field: optional string ssl_mode = 6;
     */
    sslMode?: string;
    constructor(data?: PartialMessage<PostgresConnection>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresConnection";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresConnection;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresConnection;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresConnection;
    static equals(a: PostgresConnection | PlainMessage<PostgresConnection> | undefined, b: PostgresConnection | PlainMessage<PostgresConnection> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlConnection
 */
declare class MysqlConnection extends Message<MysqlConnection> {
    /**
     * @generated from field: string user = 1;
     */
    user: string;
    /**
     * @generated from field: string pass = 2;
     */
    pass: string;
    /**
     * @generated from field: string protocol = 3;
     */
    protocol: string;
    /**
     * @generated from field: string host = 4;
     */
    host: string;
    /**
     * @generated from field: int32 port = 5;
     */
    port: number;
    /**
     * @generated from field: string name = 6;
     */
    name: string;
    constructor(data?: PartialMessage<MysqlConnection>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlConnection";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlConnection;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlConnection;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlConnection;
    static equals(a: MysqlConnection | PlainMessage<MysqlConnection> | undefined, b: MysqlConnection | PlainMessage<MysqlConnection> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlConnectionConfig
 */
declare class MysqlConnectionConfig extends Message<MysqlConnectionConfig> {
    /**
     * May provide either a raw string url, or a structured version
     *
     * @generated from oneof mgmt.v1alpha1.MysqlConnectionConfig.connection_config
     */
    connectionConfig: {
        /**
         * @generated from field: string url = 1;
         */
        value: string;
        case: "url";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.MysqlConnection connection = 2;
         */
        value: MysqlConnection;
        case: "connection";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<MysqlConnectionConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlConnectionConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlConnectionConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlConnectionConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlConnectionConfig;
    static equals(a: MysqlConnectionConfig | PlainMessage<MysqlConnectionConfig> | undefined, b: MysqlConnectionConfig | PlainMessage<MysqlConnectionConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AwsS3ConnectionConfig
 */
declare class AwsS3ConnectionConfig extends Message<AwsS3ConnectionConfig> {
    /**
     * @generated from field: string bucket_arn = 1;
     */
    bucketArn: string;
    /**
     * @generated from field: optional string path_prefix = 2;
     */
    pathPrefix?: string;
    /**
     * @generated from field: optional mgmt.v1alpha1.AwsS3Credentials credentials = 3;
     */
    credentials?: AwsS3Credentials;
    /**
     * @generated from field: optional string region = 4;
     */
    region?: string;
    /**
     * @generated from field: optional string endpoint = 5;
     */
    endpoint?: string;
    constructor(data?: PartialMessage<AwsS3ConnectionConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AwsS3ConnectionConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AwsS3ConnectionConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AwsS3ConnectionConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AwsS3ConnectionConfig;
    static equals(a: AwsS3ConnectionConfig | PlainMessage<AwsS3ConnectionConfig> | undefined, b: AwsS3ConnectionConfig | PlainMessage<AwsS3ConnectionConfig> | undefined): boolean;
}
/**
 * S3 Credentials that are used by the worker process.
 * Note: this may be optionally provided if the worker that is being hosted has environment credentials to the S3 bucket instead.
 *
 * @generated from message mgmt.v1alpha1.AwsS3Credentials
 */
declare class AwsS3Credentials extends Message<AwsS3Credentials> {
    /**
     * @generated from field: optional string profile = 1;
     */
    profile?: string;
    /**
     * @generated from field: optional string access_key_id = 2;
     */
    accessKeyId?: string;
    /**
     * @generated from field: optional string secret_access_key = 3;
     */
    secretAccessKey?: string;
    /**
     * @generated from field: optional string session_token = 4;
     */
    sessionToken?: string;
    /**
     * @generated from field: optional bool from_ec2_role = 5;
     */
    fromEc2Role?: boolean;
    /**
     * @generated from field: optional string role_arn = 6;
     */
    roleArn?: string;
    /**
     * @generated from field: optional string role_external_id = 7;
     */
    roleExternalId?: string;
    constructor(data?: PartialMessage<AwsS3Credentials>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AwsS3Credentials";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AwsS3Credentials;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AwsS3Credentials;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AwsS3Credentials;
    static equals(a: AwsS3Credentials | PlainMessage<AwsS3Credentials> | undefined, b: AwsS3Credentials | PlainMessage<AwsS3Credentials> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsConnectionNameAvailableRequest
 */
declare class IsConnectionNameAvailableRequest extends Message<IsConnectionNameAvailableRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string connection_name = 2;
     */
    connectionName: string;
    constructor(data?: PartialMessage<IsConnectionNameAvailableRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsConnectionNameAvailableRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsConnectionNameAvailableRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsConnectionNameAvailableRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsConnectionNameAvailableRequest;
    static equals(a: IsConnectionNameAvailableRequest | PlainMessage<IsConnectionNameAvailableRequest> | undefined, b: IsConnectionNameAvailableRequest | PlainMessage<IsConnectionNameAvailableRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsConnectionNameAvailableResponse
 */
declare class IsConnectionNameAvailableResponse extends Message<IsConnectionNameAvailableResponse> {
    /**
     * @generated from field: bool is_available = 1;
     */
    isAvailable: boolean;
    constructor(data?: PartialMessage<IsConnectionNameAvailableResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsConnectionNameAvailableResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsConnectionNameAvailableResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsConnectionNameAvailableResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsConnectionNameAvailableResponse;
    static equals(a: IsConnectionNameAvailableResponse | PlainMessage<IsConnectionNameAvailableResponse> | undefined, b: IsConnectionNameAvailableResponse | PlainMessage<IsConnectionNameAvailableResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DatabaseColumn
 */
declare class DatabaseColumn extends Message<DatabaseColumn> {
    /**
     * The database schema. Ex: public
     *
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * The name of the table in the schema
     *
     * @generated from field: string table = 2;
     */
    table: string;
    /**
     * The name of the column
     *
     * @generated from field: string column = 3;
     */
    column: string;
    /**
     * The datatype of the column
     *
     * @generated from field: string data_type = 4;
     */
    dataType: string;
    constructor(data?: PartialMessage<DatabaseColumn>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DatabaseColumn";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DatabaseColumn;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DatabaseColumn;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DatabaseColumn;
    static equals(a: DatabaseColumn | PlainMessage<DatabaseColumn> | undefined, b: DatabaseColumn | PlainMessage<DatabaseColumn> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionSchemaRequest
 */
declare class GetConnectionSchemaRequest extends Message<GetConnectionSchemaRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<GetConnectionSchemaRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionSchemaRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionSchemaRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionSchemaRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionSchemaRequest;
    static equals(a: GetConnectionSchemaRequest | PlainMessage<GetConnectionSchemaRequest> | undefined, b: GetConnectionSchemaRequest | PlainMessage<GetConnectionSchemaRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionSchemaResponse
 */
declare class GetConnectionSchemaResponse extends Message<GetConnectionSchemaResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.DatabaseColumn schemas = 1;
     */
    schemas: DatabaseColumn[];
    constructor(data?: PartialMessage<GetConnectionSchemaResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionSchemaResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionSchemaResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionSchemaResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionSchemaResponse;
    static equals(a: GetConnectionSchemaResponse | PlainMessage<GetConnectionSchemaResponse> | undefined, b: GetConnectionSchemaResponse | PlainMessage<GetConnectionSchemaResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CheckSqlQueryRequest
 */
declare class CheckSqlQueryRequest extends Message<CheckSqlQueryRequest> {
    /**
     * The connection id that the query will be checked against
     *
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * The full query that will be run through a PREPARE statement
     *
     * @generated from field: string query = 2;
     */
    query: string;
    constructor(data?: PartialMessage<CheckSqlQueryRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CheckSqlQueryRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CheckSqlQueryRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CheckSqlQueryRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CheckSqlQueryRequest;
    static equals(a: CheckSqlQueryRequest | PlainMessage<CheckSqlQueryRequest> | undefined, b: CheckSqlQueryRequest | PlainMessage<CheckSqlQueryRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CheckSqlQueryResponse
 */
declare class CheckSqlQueryResponse extends Message<CheckSqlQueryResponse> {
    /**
     * The query is run through PREPARE. Returns valid if it correctly compiled
     *
     * @generated from field: bool is_valid = 1;
     */
    isValid: boolean;
    /**
     * The error message returned by the sql client if the prepare did not return successfully
     *
     * @generated from field: optional string erorr_message = 2;
     */
    erorrMessage?: string;
    constructor(data?: PartialMessage<CheckSqlQueryResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CheckSqlQueryResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CheckSqlQueryResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CheckSqlQueryResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CheckSqlQueryResponse;
    static equals(a: CheckSqlQueryResponse | PlainMessage<CheckSqlQueryResponse> | undefined, b: CheckSqlQueryResponse | PlainMessage<CheckSqlQueryResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionDataStreamRequest
 */
declare class GetConnectionDataStreamRequest extends Message<GetConnectionDataStreamRequest> {
    /**
     * @generated from field: string connection_id = 1;
     */
    connectionId: string;
    /**
     * @generated from field: string schema = 2;
     */
    schema: string;
    /**
     * @generated from field: string table = 3;
     */
    table: string;
    constructor(data?: PartialMessage<GetConnectionDataStreamRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionDataStreamRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionDataStreamRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionDataStreamRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionDataStreamRequest;
    static equals(a: GetConnectionDataStreamRequest | PlainMessage<GetConnectionDataStreamRequest> | undefined, b: GetConnectionDataStreamRequest | PlainMessage<GetConnectionDataStreamRequest> | undefined): boolean;
}
/**
 * Each stream response is a single row in the requested schema and table
 *
 * @generated from message mgmt.v1alpha1.GetConnectionDataStreamResponse
 */
declare class GetConnectionDataStreamResponse extends Message<GetConnectionDataStreamResponse> {
    /**
     * A map of column name to the bytes value of the data that was found for that column and row
     *
     * @generated from field: map<string, bytes> row = 1;
     */
    row: {
        [key: string]: Uint8Array;
    };
    constructor(data?: PartialMessage<GetConnectionDataStreamResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionDataStreamResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionDataStreamResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionDataStreamResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionDataStreamResponse;
    static equals(a: GetConnectionDataStreamResponse | PlainMessage<GetConnectionDataStreamResponse> | undefined, b: GetConnectionDataStreamResponse | PlainMessage<GetConnectionDataStreamResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetConnectionForeignConstraintsRequest
 */
declare class GetConnectionForeignConstraintsRequest extends Message<GetConnectionForeignConstraintsRequest> {
    /**
     * @generated from field: string connection_id = 1;
     */
    connectionId: string;
    constructor(data?: PartialMessage<GetConnectionForeignConstraintsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionForeignConstraintsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionForeignConstraintsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionForeignConstraintsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionForeignConstraintsRequest;
    static equals(a: GetConnectionForeignConstraintsRequest | PlainMessage<GetConnectionForeignConstraintsRequest> | undefined, b: GetConnectionForeignConstraintsRequest | PlainMessage<GetConnectionForeignConstraintsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.ForeignConstraintTables
 */
declare class ForeignConstraintTables extends Message<ForeignConstraintTables> {
    /**
     * @generated from field: repeated string tables = 1;
     */
    tables: string[];
    constructor(data?: PartialMessage<ForeignConstraintTables>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.ForeignConstraintTables";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ForeignConstraintTables;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ForeignConstraintTables;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ForeignConstraintTables;
    static equals(a: ForeignConstraintTables | PlainMessage<ForeignConstraintTables> | undefined, b: ForeignConstraintTables | PlainMessage<ForeignConstraintTables> | undefined): boolean;
}
/**
 * Dependency constraints for a specific table
 *
 * @generated from message mgmt.v1alpha1.GetConnectionForeignConstraintsResponse
 */
declare class GetConnectionForeignConstraintsResponse extends Message<GetConnectionForeignConstraintsResponse> {
    /**
     * the key here is <schema>.<table> and the list of tables that it depends on, also `<schema>.<table>` format.
     *
     * @generated from field: map<string, mgmt.v1alpha1.ForeignConstraintTables> table_constraints = 1;
     */
    tableConstraints: {
        [key: string]: ForeignConstraintTables;
    };
    constructor(data?: PartialMessage<GetConnectionForeignConstraintsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetConnectionForeignConstraintsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetConnectionForeignConstraintsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetConnectionForeignConstraintsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetConnectionForeignConstraintsResponse;
    static equals(a: GetConnectionForeignConstraintsResponse | PlainMessage<GetConnectionForeignConstraintsResponse> | undefined, b: GetConnectionForeignConstraintsResponse | PlainMessage<GetConnectionForeignConstraintsResponse> | undefined): boolean;
}

/**
 * Service for managing datasource connections.
 * This is a primary data model in Neosync and is used in reference when hooking up Jobs to synchronize and generate data.
 *
 * @generated from service mgmt.v1alpha1.ConnectionService
 */
declare const ConnectionService: {
    readonly typeName: "mgmt.v1alpha1.ConnectionService";
    readonly methods: {
        /**
         * Returns a list of connections associated with the account
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnections
         */
        readonly getConnections: {
            readonly name: "GetConnections";
            readonly I: typeof GetConnectionsRequest;
            readonly O: typeof GetConnectionsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Returns a single connection
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnection
         */
        readonly getConnection: {
            readonly name: "GetConnection";
            readonly I: typeof GetConnectionRequest;
            readonly O: typeof GetConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Creates a new connection
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.CreateConnection
         */
        readonly createConnection: {
            readonly name: "CreateConnection";
            readonly I: typeof CreateConnectionRequest;
            readonly O: typeof CreateConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Updates an existing connection
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.UpdateConnection
         */
        readonly updateConnection: {
            readonly name: "UpdateConnection";
            readonly I: typeof UpdateConnectionRequest;
            readonly O: typeof UpdateConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Removes a connection from the system.
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.DeleteConnection
         */
        readonly deleteConnection: {
            readonly name: "DeleteConnection";
            readonly I: typeof DeleteConnectionRequest;
            readonly O: typeof DeleteConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Connections have friendly names, this method checks if the requested name is available in the system based on the account
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.IsConnectionNameAvailable
         */
        readonly isConnectionNameAvailable: {
            readonly name: "IsConnectionNameAvailable";
            readonly I: typeof IsConnectionNameAvailableRequest;
            readonly O: typeof IsConnectionNameAvailableResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Checks if the connection config is connectable by the backend.
         * Used mostly to verify that a connection is valid prior to creating a Connection object.
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.CheckConnectionConfig
         */
        readonly checkConnectionConfig: {
            readonly name: "CheckConnectionConfig";
            readonly I: typeof CheckConnectionConfigRequest;
            readonly O: typeof CheckConnectionConfigResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Returns the schema for a specific connection. Used mostly for SQL-based connections
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionSchema
         */
        readonly getConnectionSchema: {
            readonly name: "GetConnectionSchema";
            readonly I: typeof GetConnectionSchemaRequest;
            readonly O: typeof GetConnectionSchemaResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Checks a constructed SQL query against a sql-based connection to see if it's valid based on that connection's data schema
         * This is useful when constructing subsets to see if the WHERE clause is correct
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.CheckSqlQuery
         */
        readonly checkSqlQuery: {
            readonly name: "CheckSqlQuery";
            readonly I: typeof CheckSqlQueryRequest;
            readonly O: typeof CheckSqlQueryResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * Streaming endpoint that will stream the data available from the Connection to the client.
         * Used primarily by the CLI sync command.
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionDataStream
         */
        readonly getConnectionDataStream: {
            readonly name: "GetConnectionDataStream";
            readonly I: typeof GetConnectionDataStreamRequest;
            readonly O: typeof GetConnectionDataStreamResponse;
            readonly kind: MethodKind.ServerStreaming;
        };
        /**
         * For a specific connection, returns the foreign key constraints. Mostly useful for SQL-based Connections.
         * Used primarily by the CLI sync command to determine stream order.
         *
         * @generated from rpc mgmt.v1alpha1.ConnectionService.GetConnectionForeignConstraints
         */
        readonly getConnectionForeignConstraints: {
            readonly name: "GetConnectionForeignConstraints";
            readonly I: typeof GetConnectionForeignConstraintsRequest;
            readonly O: typeof GetConnectionForeignConstraintsResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};

/**
 * @generated from message mgmt.v1alpha1.GetSystemTransformersRequest
 */
declare class GetSystemTransformersRequest extends Message<GetSystemTransformersRequest> {
    constructor(data?: PartialMessage<GetSystemTransformersRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetSystemTransformersRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSystemTransformersRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSystemTransformersRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSystemTransformersRequest;
    static equals(a: GetSystemTransformersRequest | PlainMessage<GetSystemTransformersRequest> | undefined, b: GetSystemTransformersRequest | PlainMessage<GetSystemTransformersRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetSystemTransformersResponse
 */
declare class GetSystemTransformersResponse extends Message<GetSystemTransformersResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.SystemTransformer transformers = 1;
     */
    transformers: SystemTransformer[];
    constructor(data?: PartialMessage<GetSystemTransformersResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetSystemTransformersResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSystemTransformersResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSystemTransformersResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSystemTransformersResponse;
    static equals(a: GetSystemTransformersResponse | PlainMessage<GetSystemTransformersResponse> | undefined, b: GetSystemTransformersResponse | PlainMessage<GetSystemTransformersResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserDefinedTransformersRequest
 */
declare class GetUserDefinedTransformersRequest extends Message<GetUserDefinedTransformersRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetUserDefinedTransformersRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserDefinedTransformersRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserDefinedTransformersRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserDefinedTransformersRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserDefinedTransformersRequest;
    static equals(a: GetUserDefinedTransformersRequest | PlainMessage<GetUserDefinedTransformersRequest> | undefined, b: GetUserDefinedTransformersRequest | PlainMessage<GetUserDefinedTransformersRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserDefinedTransformersResponse
 */
declare class GetUserDefinedTransformersResponse extends Message<GetUserDefinedTransformersResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.UserDefinedTransformer transformers = 1;
     */
    transformers: UserDefinedTransformer[];
    constructor(data?: PartialMessage<GetUserDefinedTransformersResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserDefinedTransformersResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserDefinedTransformersResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserDefinedTransformersResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserDefinedTransformersResponse;
    static equals(a: GetUserDefinedTransformersResponse | PlainMessage<GetUserDefinedTransformersResponse> | undefined, b: GetUserDefinedTransformersResponse | PlainMessage<GetUserDefinedTransformersResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserDefinedTransformerByIdRequest
 */
declare class GetUserDefinedTransformerByIdRequest extends Message<GetUserDefinedTransformerByIdRequest> {
    /**
     * @generated from field: string transformer_id = 1;
     */
    transformerId: string;
    constructor(data?: PartialMessage<GetUserDefinedTransformerByIdRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserDefinedTransformerByIdRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserDefinedTransformerByIdRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserDefinedTransformerByIdRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserDefinedTransformerByIdRequest;
    static equals(a: GetUserDefinedTransformerByIdRequest | PlainMessage<GetUserDefinedTransformerByIdRequest> | undefined, b: GetUserDefinedTransformerByIdRequest | PlainMessage<GetUserDefinedTransformerByIdRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserDefinedTransformerByIdResponse
 */
declare class GetUserDefinedTransformerByIdResponse extends Message<GetUserDefinedTransformerByIdResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
     */
    transformer?: UserDefinedTransformer;
    constructor(data?: PartialMessage<GetUserDefinedTransformerByIdResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserDefinedTransformerByIdResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserDefinedTransformerByIdResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserDefinedTransformerByIdResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserDefinedTransformerByIdResponse;
    static equals(a: GetUserDefinedTransformerByIdResponse | PlainMessage<GetUserDefinedTransformerByIdResponse> | undefined, b: GetUserDefinedTransformerByIdResponse | PlainMessage<GetUserDefinedTransformerByIdResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateUserDefinedTransformerRequest
 */
declare class CreateUserDefinedTransformerRequest extends Message<CreateUserDefinedTransformerRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string description = 3;
     */
    description: string;
    /**
     * @generated from field: string type = 4;
     */
    type: string;
    /**
     * @generated from field: string source = 5;
     */
    source: string;
    /**
     * @generated from field: mgmt.v1alpha1.TransformerConfig transformer_config = 6;
     */
    transformerConfig?: TransformerConfig;
    constructor(data?: PartialMessage<CreateUserDefinedTransformerRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateUserDefinedTransformerRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateUserDefinedTransformerRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateUserDefinedTransformerRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateUserDefinedTransformerRequest;
    static equals(a: CreateUserDefinedTransformerRequest | PlainMessage<CreateUserDefinedTransformerRequest> | undefined, b: CreateUserDefinedTransformerRequest | PlainMessage<CreateUserDefinedTransformerRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateUserDefinedTransformerResponse
 */
declare class CreateUserDefinedTransformerResponse extends Message<CreateUserDefinedTransformerResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
     */
    transformer?: UserDefinedTransformer;
    constructor(data?: PartialMessage<CreateUserDefinedTransformerResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateUserDefinedTransformerResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateUserDefinedTransformerResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateUserDefinedTransformerResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateUserDefinedTransformerResponse;
    static equals(a: CreateUserDefinedTransformerResponse | PlainMessage<CreateUserDefinedTransformerResponse> | undefined, b: CreateUserDefinedTransformerResponse | PlainMessage<CreateUserDefinedTransformerResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteUserDefinedTransformerRequest
 */
declare class DeleteUserDefinedTransformerRequest extends Message<DeleteUserDefinedTransformerRequest> {
    /**
     * @generated from field: string transformer_id = 1;
     */
    transformerId: string;
    constructor(data?: PartialMessage<DeleteUserDefinedTransformerRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteUserDefinedTransformerRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteUserDefinedTransformerRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteUserDefinedTransformerRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteUserDefinedTransformerRequest;
    static equals(a: DeleteUserDefinedTransformerRequest | PlainMessage<DeleteUserDefinedTransformerRequest> | undefined, b: DeleteUserDefinedTransformerRequest | PlainMessage<DeleteUserDefinedTransformerRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteUserDefinedTransformerResponse
 */
declare class DeleteUserDefinedTransformerResponse extends Message<DeleteUserDefinedTransformerResponse> {
    constructor(data?: PartialMessage<DeleteUserDefinedTransformerResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteUserDefinedTransformerResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteUserDefinedTransformerResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteUserDefinedTransformerResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteUserDefinedTransformerResponse;
    static equals(a: DeleteUserDefinedTransformerResponse | PlainMessage<DeleteUserDefinedTransformerResponse> | undefined, b: DeleteUserDefinedTransformerResponse | PlainMessage<DeleteUserDefinedTransformerResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateUserDefinedTransformerRequest
 */
declare class UpdateUserDefinedTransformerRequest extends Message<UpdateUserDefinedTransformerRequest> {
    /**
     * @generated from field: string transformer_id = 1;
     */
    transformerId: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string description = 3;
     */
    description: string;
    /**
     * @generated from field: mgmt.v1alpha1.TransformerConfig transformer_config = 4;
     */
    transformerConfig?: TransformerConfig;
    constructor(data?: PartialMessage<UpdateUserDefinedTransformerRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateUserDefinedTransformerRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateUserDefinedTransformerRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateUserDefinedTransformerRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateUserDefinedTransformerRequest;
    static equals(a: UpdateUserDefinedTransformerRequest | PlainMessage<UpdateUserDefinedTransformerRequest> | undefined, b: UpdateUserDefinedTransformerRequest | PlainMessage<UpdateUserDefinedTransformerRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateUserDefinedTransformerResponse
 */
declare class UpdateUserDefinedTransformerResponse extends Message<UpdateUserDefinedTransformerResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.UserDefinedTransformer transformer = 1;
     */
    transformer?: UserDefinedTransformer;
    constructor(data?: PartialMessage<UpdateUserDefinedTransformerResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateUserDefinedTransformerResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateUserDefinedTransformerResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateUserDefinedTransformerResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateUserDefinedTransformerResponse;
    static equals(a: UpdateUserDefinedTransformerResponse | PlainMessage<UpdateUserDefinedTransformerResponse> | undefined, b: UpdateUserDefinedTransformerResponse | PlainMessage<UpdateUserDefinedTransformerResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsTransformerNameAvailableRequest
 */
declare class IsTransformerNameAvailableRequest extends Message<IsTransformerNameAvailableRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string transformer_name = 2;
     */
    transformerName: string;
    constructor(data?: PartialMessage<IsTransformerNameAvailableRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsTransformerNameAvailableRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsTransformerNameAvailableRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsTransformerNameAvailableRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsTransformerNameAvailableRequest;
    static equals(a: IsTransformerNameAvailableRequest | PlainMessage<IsTransformerNameAvailableRequest> | undefined, b: IsTransformerNameAvailableRequest | PlainMessage<IsTransformerNameAvailableRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsTransformerNameAvailableResponse
 */
declare class IsTransformerNameAvailableResponse extends Message<IsTransformerNameAvailableResponse> {
    /**
     * @generated from field: bool is_available = 1;
     */
    isAvailable: boolean;
    constructor(data?: PartialMessage<IsTransformerNameAvailableResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsTransformerNameAvailableResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsTransformerNameAvailableResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsTransformerNameAvailableResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsTransformerNameAvailableResponse;
    static equals(a: IsTransformerNameAvailableResponse | PlainMessage<IsTransformerNameAvailableResponse> | undefined, b: IsTransformerNameAvailableResponse | PlainMessage<IsTransformerNameAvailableResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UserDefinedTransformer
 */
declare class UserDefinedTransformer extends Message<UserDefinedTransformer> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string description = 3;
     */
    description: string;
    /**
     * @generated from field: string data_type = 5;
     */
    dataType: string;
    /**
     * @generated from field: string source = 6;
     */
    source: string;
    /**
     * @generated from field: mgmt.v1alpha1.TransformerConfig config = 7;
     */
    config?: TransformerConfig;
    /**
     * @generated from field: google.protobuf.Timestamp created_at = 8;
     */
    createdAt?: Timestamp;
    /**
     * @generated from field: google.protobuf.Timestamp updated_at = 9;
     */
    updatedAt?: Timestamp;
    /**
     * @generated from field: string account_id = 10;
     */
    accountId: string;
    constructor(data?: PartialMessage<UserDefinedTransformer>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UserDefinedTransformer";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UserDefinedTransformer;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UserDefinedTransformer;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UserDefinedTransformer;
    static equals(a: UserDefinedTransformer | PlainMessage<UserDefinedTransformer> | undefined, b: UserDefinedTransformer | PlainMessage<UserDefinedTransformer> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SystemTransformer
 */
declare class SystemTransformer extends Message<SystemTransformer> {
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string description = 3;
     */
    description: string;
    /**
     * @generated from field: string data_type = 5;
     */
    dataType: string;
    /**
     * @generated from field: string source = 6;
     */
    source: string;
    /**
     * @generated from field: mgmt.v1alpha1.TransformerConfig config = 7;
     */
    config?: TransformerConfig;
    constructor(data?: PartialMessage<SystemTransformer>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SystemTransformer";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SystemTransformer;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SystemTransformer;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SystemTransformer;
    static equals(a: SystemTransformer | PlainMessage<SystemTransformer> | undefined, b: SystemTransformer | PlainMessage<SystemTransformer> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformerConfig
 */
declare class TransformerConfig extends Message<TransformerConfig> {
    /**
     * @generated from oneof mgmt.v1alpha1.TransformerConfig.config
     */
    config: {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateEmail generate_email_config = 1;
         */
        value: GenerateEmail;
        case: "generateEmailConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateRealisticEmail generate_realistic_email_config = 2;
         */
        value: GenerateRealisticEmail;
        case: "generateRealisticEmailConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformEmail transform_email_config = 3;
         */
        value: TransformEmail;
        case: "transformEmailConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateBool generate_bool_config = 4;
         */
        value: GenerateBool;
        case: "generateBoolConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateCardNumber generate_card_number_config = 5;
         */
        value: GenerateCardNumber;
        case: "generateCardNumberConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateCity generate_city_config = 6;
         */
        value: GenerateCity;
        case: "generateCityConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateE164Number generate_e164_number_config = 7;
         */
        value: GenerateE164Number;
        case: "generateE164NumberConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateFirstName generate_first_name_config = 8;
         */
        value: GenerateFirstName;
        case: "generateFirstNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateFloat generate_float_config = 9;
         */
        value: GenerateFloat;
        case: "generateFloatConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateFullAddress generate_full_address_config = 10;
         */
        value: GenerateFullAddress;
        case: "generateFullAddressConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateFullName generate_full_name_config = 11;
         */
        value: GenerateFullName;
        case: "generateFullNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateGender generate_gender_config = 12;
         */
        value: GenerateGender;
        case: "generateGenderConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateInt64Phone generate_int64_phone_config = 13;
         */
        value: GenerateInt64Phone;
        case: "generateInt64PhoneConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateInt generate_int_config = 14;
         */
        value: GenerateInt;
        case: "generateIntConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateLastName generate_last_name_config = 15;
         */
        value: GenerateLastName;
        case: "generateLastNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateSha256Hash generate_sha256hash_config = 16;
         */
        value: GenerateSha256Hash;
        case: "generateSha256hashConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateSSN generate_ssn_config = 17;
         */
        value: GenerateSSN;
        case: "generateSsnConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateState generate_state_config = 18;
         */
        value: GenerateState;
        case: "generateStateConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateStreetAddress generate_street_address_config = 19;
         */
        value: GenerateStreetAddress;
        case: "generateStreetAddressConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateStringPhone generate_string_phone_config = 20;
         */
        value: GenerateStringPhone;
        case: "generateStringPhoneConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateString generate_string_config = 21;
         */
        value: GenerateString;
        case: "generateStringConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateUnixTimestamp generate_unixtimestamp_config = 22;
         */
        value: GenerateUnixTimestamp;
        case: "generateUnixtimestampConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateUsername generate_username_config = 23;
         */
        value: GenerateUsername;
        case: "generateUsernameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateUtcTimestamp generate_utctimestamp_config = 24;
         */
        value: GenerateUtcTimestamp;
        case: "generateUtctimestampConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateUuid generate_uuid_config = 25;
         */
        value: GenerateUuid;
        case: "generateUuidConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateZipcode generate_zipcode_config = 26;
         */
        value: GenerateZipcode;
        case: "generateZipcodeConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformE164Phone transform_e164_phone_config = 27;
         */
        value: TransformE164Phone;
        case: "transformE164PhoneConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformFirstName transform_first_name_config = 28;
         */
        value: TransformFirstName;
        case: "transformFirstNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformFloat transform_float_config = 29;
         */
        value: TransformFloat;
        case: "transformFloatConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformFullName transform_full_name_config = 30;
         */
        value: TransformFullName;
        case: "transformFullNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformIntPhone transform_int_phone_config = 31;
         */
        value: TransformIntPhone;
        case: "transformIntPhoneConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformInt transform_int_config = 32;
         */
        value: TransformInt;
        case: "transformIntConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformLastName transform_last_name_config = 33;
         */
        value: TransformLastName;
        case: "transformLastNameConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformPhone transform_phone_config = 34;
         */
        value: TransformPhone;
        case: "transformPhoneConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.TransformString transform_string_config = 35;
         */
        value: TransformString;
        case: "transformStringConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.Passthrough passthrough_config = 36;
         */
        value: Passthrough;
        case: "passthroughConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.Null nullconfig = 37;
         */
        value: Null;
        case: "nullconfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.UserDefinedTransformerConfig user_defined_transformer_config = 38;
         */
        value: UserDefinedTransformerConfig;
        case: "userDefinedTransformerConfig";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateDefault generate_default_config = 39;
         */
        value: GenerateDefault;
        case: "generateDefaultConfig";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<TransformerConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformerConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformerConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformerConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformerConfig;
    static equals(a: TransformerConfig | PlainMessage<TransformerConfig> | undefined, b: TransformerConfig | PlainMessage<TransformerConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateEmail
 */
declare class GenerateEmail extends Message<GenerateEmail> {
    constructor(data?: PartialMessage<GenerateEmail>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateEmail";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateEmail;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateEmail;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateEmail;
    static equals(a: GenerateEmail | PlainMessage<GenerateEmail> | undefined, b: GenerateEmail | PlainMessage<GenerateEmail> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateRealisticEmail
 */
declare class GenerateRealisticEmail extends Message<GenerateRealisticEmail> {
    constructor(data?: PartialMessage<GenerateRealisticEmail>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateRealisticEmail";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateRealisticEmail;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateRealisticEmail;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateRealisticEmail;
    static equals(a: GenerateRealisticEmail | PlainMessage<GenerateRealisticEmail> | undefined, b: GenerateRealisticEmail | PlainMessage<GenerateRealisticEmail> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformEmail
 */
declare class TransformEmail extends Message<TransformEmail> {
    /**
     * @generated from field: bool preserve_domain = 1;
     */
    preserveDomain: boolean;
    /**
     * @generated from field: bool preserve_length = 2;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformEmail>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformEmail";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformEmail;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformEmail;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformEmail;
    static equals(a: TransformEmail | PlainMessage<TransformEmail> | undefined, b: TransformEmail | PlainMessage<TransformEmail> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateBool
 */
declare class GenerateBool extends Message<GenerateBool> {
    constructor(data?: PartialMessage<GenerateBool>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateBool";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateBool;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateBool;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateBool;
    static equals(a: GenerateBool | PlainMessage<GenerateBool> | undefined, b: GenerateBool | PlainMessage<GenerateBool> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateCardNumber
 */
declare class GenerateCardNumber extends Message<GenerateCardNumber> {
    /**
     * @generated from field: bool valid_luhn = 1;
     */
    validLuhn: boolean;
    constructor(data?: PartialMessage<GenerateCardNumber>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateCardNumber";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateCardNumber;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateCardNumber;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateCardNumber;
    static equals(a: GenerateCardNumber | PlainMessage<GenerateCardNumber> | undefined, b: GenerateCardNumber | PlainMessage<GenerateCardNumber> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateCity
 */
declare class GenerateCity extends Message<GenerateCity> {
    constructor(data?: PartialMessage<GenerateCity>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateCity";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateCity;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateCity;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateCity;
    static equals(a: GenerateCity | PlainMessage<GenerateCity> | undefined, b: GenerateCity | PlainMessage<GenerateCity> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateDefault
 */
declare class GenerateDefault extends Message<GenerateDefault> {
    constructor(data?: PartialMessage<GenerateDefault>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateDefault";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateDefault;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateDefault;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateDefault;
    static equals(a: GenerateDefault | PlainMessage<GenerateDefault> | undefined, b: GenerateDefault | PlainMessage<GenerateDefault> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateE164Number
 */
declare class GenerateE164Number extends Message<GenerateE164Number> {
    /**
     * @generated from field: int64 length = 1;
     */
    length: bigint;
    constructor(data?: PartialMessage<GenerateE164Number>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateE164Number";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateE164Number;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateE164Number;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateE164Number;
    static equals(a: GenerateE164Number | PlainMessage<GenerateE164Number> | undefined, b: GenerateE164Number | PlainMessage<GenerateE164Number> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateFirstName
 */
declare class GenerateFirstName extends Message<GenerateFirstName> {
    constructor(data?: PartialMessage<GenerateFirstName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateFirstName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateFirstName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateFirstName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateFirstName;
    static equals(a: GenerateFirstName | PlainMessage<GenerateFirstName> | undefined, b: GenerateFirstName | PlainMessage<GenerateFirstName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateFloat
 */
declare class GenerateFloat extends Message<GenerateFloat> {
    /**
     * @generated from field: string sign = 1;
     */
    sign: string;
    /**
     * @generated from field: int64 digits_before_decimal = 2;
     */
    digitsBeforeDecimal: bigint;
    /**
     * @generated from field: int64 digits_after_decimal = 3;
     */
    digitsAfterDecimal: bigint;
    constructor(data?: PartialMessage<GenerateFloat>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateFloat";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateFloat;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateFloat;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateFloat;
    static equals(a: GenerateFloat | PlainMessage<GenerateFloat> | undefined, b: GenerateFloat | PlainMessage<GenerateFloat> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateFullAddress
 */
declare class GenerateFullAddress extends Message<GenerateFullAddress> {
    constructor(data?: PartialMessage<GenerateFullAddress>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateFullAddress";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateFullAddress;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateFullAddress;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateFullAddress;
    static equals(a: GenerateFullAddress | PlainMessage<GenerateFullAddress> | undefined, b: GenerateFullAddress | PlainMessage<GenerateFullAddress> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateFullName
 */
declare class GenerateFullName extends Message<GenerateFullName> {
    constructor(data?: PartialMessage<GenerateFullName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateFullName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateFullName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateFullName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateFullName;
    static equals(a: GenerateFullName | PlainMessage<GenerateFullName> | undefined, b: GenerateFullName | PlainMessage<GenerateFullName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateGender
 */
declare class GenerateGender extends Message<GenerateGender> {
    /**
     * @generated from field: bool abbreviate = 1;
     */
    abbreviate: boolean;
    constructor(data?: PartialMessage<GenerateGender>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateGender";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateGender;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateGender;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateGender;
    static equals(a: GenerateGender | PlainMessage<GenerateGender> | undefined, b: GenerateGender | PlainMessage<GenerateGender> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateInt64Phone
 */
declare class GenerateInt64Phone extends Message<GenerateInt64Phone> {
    constructor(data?: PartialMessage<GenerateInt64Phone>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateInt64Phone";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateInt64Phone;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateInt64Phone;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateInt64Phone;
    static equals(a: GenerateInt64Phone | PlainMessage<GenerateInt64Phone> | undefined, b: GenerateInt64Phone | PlainMessage<GenerateInt64Phone> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateInt
 */
declare class GenerateInt extends Message<GenerateInt> {
    /**
     * @generated from field: int64 length = 1;
     */
    length: bigint;
    /**
     * @generated from field: string sign = 2;
     */
    sign: string;
    constructor(data?: PartialMessage<GenerateInt>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateInt";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateInt;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateInt;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateInt;
    static equals(a: GenerateInt | PlainMessage<GenerateInt> | undefined, b: GenerateInt | PlainMessage<GenerateInt> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateLastName
 */
declare class GenerateLastName extends Message<GenerateLastName> {
    constructor(data?: PartialMessage<GenerateLastName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateLastName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateLastName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateLastName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateLastName;
    static equals(a: GenerateLastName | PlainMessage<GenerateLastName> | undefined, b: GenerateLastName | PlainMessage<GenerateLastName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateSha256Hash
 */
declare class GenerateSha256Hash extends Message<GenerateSha256Hash> {
    constructor(data?: PartialMessage<GenerateSha256Hash>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateSha256Hash";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateSha256Hash;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateSha256Hash;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateSha256Hash;
    static equals(a: GenerateSha256Hash | PlainMessage<GenerateSha256Hash> | undefined, b: GenerateSha256Hash | PlainMessage<GenerateSha256Hash> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateSSN
 */
declare class GenerateSSN extends Message<GenerateSSN> {
    constructor(data?: PartialMessage<GenerateSSN>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateSSN";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateSSN;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateSSN;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateSSN;
    static equals(a: GenerateSSN | PlainMessage<GenerateSSN> | undefined, b: GenerateSSN | PlainMessage<GenerateSSN> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateState
 */
declare class GenerateState extends Message<GenerateState> {
    constructor(data?: PartialMessage<GenerateState>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateState";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateState;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateState;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateState;
    static equals(a: GenerateState | PlainMessage<GenerateState> | undefined, b: GenerateState | PlainMessage<GenerateState> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateStreetAddress
 */
declare class GenerateStreetAddress extends Message<GenerateStreetAddress> {
    constructor(data?: PartialMessage<GenerateStreetAddress>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateStreetAddress";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateStreetAddress;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateStreetAddress;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateStreetAddress;
    static equals(a: GenerateStreetAddress | PlainMessage<GenerateStreetAddress> | undefined, b: GenerateStreetAddress | PlainMessage<GenerateStreetAddress> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateStringPhone
 */
declare class GenerateStringPhone extends Message<GenerateStringPhone> {
    /**
     * @generated from field: bool include_hyphens = 2;
     */
    includeHyphens: boolean;
    constructor(data?: PartialMessage<GenerateStringPhone>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateStringPhone";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateStringPhone;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateStringPhone;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateStringPhone;
    static equals(a: GenerateStringPhone | PlainMessage<GenerateStringPhone> | undefined, b: GenerateStringPhone | PlainMessage<GenerateStringPhone> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateString
 */
declare class GenerateString extends Message<GenerateString> {
    /**
     * @generated from field: int64 length = 1;
     */
    length: bigint;
    constructor(data?: PartialMessage<GenerateString>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateString";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateString;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateString;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateString;
    static equals(a: GenerateString | PlainMessage<GenerateString> | undefined, b: GenerateString | PlainMessage<GenerateString> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateUnixTimestamp
 */
declare class GenerateUnixTimestamp extends Message<GenerateUnixTimestamp> {
    constructor(data?: PartialMessage<GenerateUnixTimestamp>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateUnixTimestamp";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateUnixTimestamp;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateUnixTimestamp;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateUnixTimestamp;
    static equals(a: GenerateUnixTimestamp | PlainMessage<GenerateUnixTimestamp> | undefined, b: GenerateUnixTimestamp | PlainMessage<GenerateUnixTimestamp> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateUsername
 */
declare class GenerateUsername extends Message<GenerateUsername> {
    constructor(data?: PartialMessage<GenerateUsername>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateUsername";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateUsername;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateUsername;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateUsername;
    static equals(a: GenerateUsername | PlainMessage<GenerateUsername> | undefined, b: GenerateUsername | PlainMessage<GenerateUsername> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateUtcTimestamp
 */
declare class GenerateUtcTimestamp extends Message<GenerateUtcTimestamp> {
    constructor(data?: PartialMessage<GenerateUtcTimestamp>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateUtcTimestamp";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateUtcTimestamp;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateUtcTimestamp;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateUtcTimestamp;
    static equals(a: GenerateUtcTimestamp | PlainMessage<GenerateUtcTimestamp> | undefined, b: GenerateUtcTimestamp | PlainMessage<GenerateUtcTimestamp> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateUuid
 */
declare class GenerateUuid extends Message<GenerateUuid> {
    /**
     * @generated from field: bool include_hyphens = 1;
     */
    includeHyphens: boolean;
    constructor(data?: PartialMessage<GenerateUuid>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateUuid";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateUuid;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateUuid;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateUuid;
    static equals(a: GenerateUuid | PlainMessage<GenerateUuid> | undefined, b: GenerateUuid | PlainMessage<GenerateUuid> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateZipcode
 */
declare class GenerateZipcode extends Message<GenerateZipcode> {
    constructor(data?: PartialMessage<GenerateZipcode>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateZipcode";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateZipcode;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateZipcode;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateZipcode;
    static equals(a: GenerateZipcode | PlainMessage<GenerateZipcode> | undefined, b: GenerateZipcode | PlainMessage<GenerateZipcode> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformE164Phone
 */
declare class TransformE164Phone extends Message<TransformE164Phone> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformE164Phone>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformE164Phone";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformE164Phone;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformE164Phone;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformE164Phone;
    static equals(a: TransformE164Phone | PlainMessage<TransformE164Phone> | undefined, b: TransformE164Phone | PlainMessage<TransformE164Phone> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformFirstName
 */
declare class TransformFirstName extends Message<TransformFirstName> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformFirstName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformFirstName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformFirstName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformFirstName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformFirstName;
    static equals(a: TransformFirstName | PlainMessage<TransformFirstName> | undefined, b: TransformFirstName | PlainMessage<TransformFirstName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformFloat
 */
declare class TransformFloat extends Message<TransformFloat> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    /**
     * @generated from field: bool preserve_sign = 2;
     */
    preserveSign: boolean;
    constructor(data?: PartialMessage<TransformFloat>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformFloat";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformFloat;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformFloat;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformFloat;
    static equals(a: TransformFloat | PlainMessage<TransformFloat> | undefined, b: TransformFloat | PlainMessage<TransformFloat> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformFullName
 */
declare class TransformFullName extends Message<TransformFullName> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformFullName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformFullName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformFullName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformFullName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformFullName;
    static equals(a: TransformFullName | PlainMessage<TransformFullName> | undefined, b: TransformFullName | PlainMessage<TransformFullName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformIntPhone
 */
declare class TransformIntPhone extends Message<TransformIntPhone> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformIntPhone>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformIntPhone";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformIntPhone;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformIntPhone;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformIntPhone;
    static equals(a: TransformIntPhone | PlainMessage<TransformIntPhone> | undefined, b: TransformIntPhone | PlainMessage<TransformIntPhone> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformInt
 */
declare class TransformInt extends Message<TransformInt> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    /**
     * @generated from field: bool preserve_sign = 2;
     */
    preserveSign: boolean;
    constructor(data?: PartialMessage<TransformInt>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformInt";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformInt;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformInt;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformInt;
    static equals(a: TransformInt | PlainMessage<TransformInt> | undefined, b: TransformInt | PlainMessage<TransformInt> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformLastName
 */
declare class TransformLastName extends Message<TransformLastName> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformLastName>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformLastName";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformLastName;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformLastName;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformLastName;
    static equals(a: TransformLastName | PlainMessage<TransformLastName> | undefined, b: TransformLastName | PlainMessage<TransformLastName> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformPhone
 */
declare class TransformPhone extends Message<TransformPhone> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    /**
     * @generated from field: bool include_hyphens = 2;
     */
    includeHyphens: boolean;
    constructor(data?: PartialMessage<TransformPhone>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformPhone";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformPhone;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformPhone;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformPhone;
    static equals(a: TransformPhone | PlainMessage<TransformPhone> | undefined, b: TransformPhone | PlainMessage<TransformPhone> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.TransformString
 */
declare class TransformString extends Message<TransformString> {
    /**
     * @generated from field: bool preserve_length = 1;
     */
    preserveLength: boolean;
    constructor(data?: PartialMessage<TransformString>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.TransformString";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TransformString;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TransformString;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TransformString;
    static equals(a: TransformString | PlainMessage<TransformString> | undefined, b: TransformString | PlainMessage<TransformString> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.Passthrough
 */
declare class Passthrough extends Message<Passthrough> {
    constructor(data?: PartialMessage<Passthrough>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.Passthrough";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Passthrough;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Passthrough;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Passthrough;
    static equals(a: Passthrough | PlainMessage<Passthrough> | undefined, b: Passthrough | PlainMessage<Passthrough> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.Null
 */
declare class Null extends Message<Null> {
    constructor(data?: PartialMessage<Null>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.Null";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Null;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Null;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Null;
    static equals(a: Null | PlainMessage<Null> | undefined, b: Null | PlainMessage<Null> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UserDefinedTransformerConfig
 */
declare class UserDefinedTransformerConfig extends Message<UserDefinedTransformerConfig> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<UserDefinedTransformerConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UserDefinedTransformerConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UserDefinedTransformerConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UserDefinedTransformerConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UserDefinedTransformerConfig;
    static equals(a: UserDefinedTransformerConfig | PlainMessage<UserDefinedTransformerConfig> | undefined, b: UserDefinedTransformerConfig | PlainMessage<UserDefinedTransformerConfig> | undefined): boolean;
}

/**
 * @generated from enum mgmt.v1alpha1.JobStatus
 */
declare enum JobStatus {
    /**
     * @generated from enum value: JOB_STATUS_UNSPECIFIED = 0;
     */
    UNSPECIFIED = 0,
    /**
     * @generated from enum value: JOB_STATUS_ENABLED = 1;
     */
    ENABLED = 1,
    /**
     * @generated from enum value: JOB_STATUS_PAUSED = 3;
     */
    PAUSED = 3,
    /**
     * @generated from enum value: JOB_STATUS_DISABLED = 4;
     */
    DISABLED = 4
}
/**
 * @generated from enum mgmt.v1alpha1.ActivityStatus
 */
declare enum ActivityStatus {
    /**
     * @generated from enum value: ACTIVITY_STATUS_UNSPECIFIED = 0;
     */
    UNSPECIFIED = 0,
    /**
     * @generated from enum value: ACTIVITY_STATUS_SCHEDULED = 1;
     */
    SCHEDULED = 1,
    /**
     * @generated from enum value: ACTIVITY_STATUS_STARTED = 2;
     */
    STARTED = 2,
    /**
     * @generated from enum value: ACTIVITY_STATUS_CANCELED = 3;
     */
    CANCELED = 3,
    /**
     * @generated from enum value: ACTIVITY_STATUS_FAILED = 4;
     */
    FAILED = 4
}
/**
 * @generated from enum mgmt.v1alpha1.JobRunStatus
 */
declare enum JobRunStatus {
    /**
     * @generated from enum value: JOB_RUN_STATUS_UNSPECIFIED = 0;
     */
    UNSPECIFIED = 0,
    /**
     * @generated from enum value: JOB_RUN_STATUS_PENDING = 1;
     */
    PENDING = 1,
    /**
     * @generated from enum value: JOB_RUN_STATUS_RUNNING = 2;
     */
    RUNNING = 2,
    /**
     * @generated from enum value: JOB_RUN_STATUS_COMPLETE = 3;
     */
    COMPLETE = 3,
    /**
     * @generated from enum value: JOB_RUN_STATUS_ERROR = 4;
     */
    ERROR = 4,
    /**
     * @generated from enum value: JOB_RUN_STATUS_CANCELED = 5;
     */
    CANCELED = 5,
    /**
     * @generated from enum value: JOB_RUN_STATUS_TERMINATED = 6;
     */
    TERMINATED = 6,
    /**
     * @generated from enum value: JOB_RUN_STATUS_FAILED = 7;
     */
    FAILED = 7
}
/**
 * @generated from message mgmt.v1alpha1.GetJobsRequest
 */
declare class GetJobsRequest extends Message<GetJobsRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetJobsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobsRequest;
    static equals(a: GetJobsRequest | PlainMessage<GetJobsRequest> | undefined, b: GetJobsRequest | PlainMessage<GetJobsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobsResponse
 */
declare class GetJobsResponse extends Message<GetJobsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.Job jobs = 1;
     */
    jobs: Job[];
    constructor(data?: PartialMessage<GetJobsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobsResponse;
    static equals(a: GetJobsResponse | PlainMessage<GetJobsResponse> | undefined, b: GetJobsResponse | PlainMessage<GetJobsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobSource
 */
declare class JobSource extends Message<JobSource> {
    /**
     * @generated from field: mgmt.v1alpha1.JobSourceOptions options = 1;
     */
    options?: JobSourceOptions;
    constructor(data?: PartialMessage<JobSource>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobSource";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobSource;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobSource;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobSource;
    static equals(a: JobSource | PlainMessage<JobSource> | undefined, b: JobSource | PlainMessage<JobSource> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobSourceOptions
 */
declare class JobSourceOptions extends Message<JobSourceOptions> {
    /**
     * @generated from oneof mgmt.v1alpha1.JobSourceOptions.config
     */
    config: {
        /**
         * @generated from field: mgmt.v1alpha1.PostgresSourceConnectionOptions postgres = 1;
         */
        value: PostgresSourceConnectionOptions;
        case: "postgres";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.AwsS3SourceConnectionOptions aws_s3 = 2;
         */
        value: AwsS3SourceConnectionOptions;
        case: "awsS3";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.MysqlSourceConnectionOptions mysql = 3;
         */
        value: MysqlSourceConnectionOptions;
        case: "mysql";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.GenerateSourceOptions generate = 4;
         */
        value: GenerateSourceOptions;
        case: "generate";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<JobSourceOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobSourceOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobSourceOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobSourceOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobSourceOptions;
    static equals(a: JobSourceOptions | PlainMessage<JobSourceOptions> | undefined, b: JobSourceOptions | PlainMessage<JobSourceOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobDestination
 */
declare class CreateJobDestination extends Message<CreateJobDestination> {
    /**
     * @generated from field: string connection_id = 1;
     */
    connectionId: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 2;
     */
    options?: JobDestinationOptions;
    constructor(data?: PartialMessage<CreateJobDestination>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobDestination";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobDestination;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobDestination;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobDestination;
    static equals(a: CreateJobDestination | PlainMessage<CreateJobDestination> | undefined, b: CreateJobDestination | PlainMessage<CreateJobDestination> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobDestination
 */
declare class JobDestination extends Message<JobDestination> {
    /**
     * @generated from field: string connection_id = 1;
     */
    connectionId: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 2;
     */
    options?: JobDestinationOptions;
    /**
     * @generated from field: string id = 3;
     */
    id: string;
    constructor(data?: PartialMessage<JobDestination>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobDestination";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobDestination;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobDestination;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobDestination;
    static equals(a: JobDestination | PlainMessage<JobDestination> | undefined, b: JobDestination | PlainMessage<JobDestination> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateSourceOptions
 */
declare class GenerateSourceOptions extends Message<GenerateSourceOptions> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.GenerateSourceSchemaOption schemas = 1;
     */
    schemas: GenerateSourceSchemaOption[];
    /**
     * @generated from field: optional string fk_source_connection_id = 3;
     */
    fkSourceConnectionId?: string;
    constructor(data?: PartialMessage<GenerateSourceOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateSourceOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateSourceOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateSourceOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateSourceOptions;
    static equals(a: GenerateSourceOptions | PlainMessage<GenerateSourceOptions> | undefined, b: GenerateSourceOptions | PlainMessage<GenerateSourceOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateSourceSchemaOption
 */
declare class GenerateSourceSchemaOption extends Message<GenerateSourceSchemaOption> {
    /**
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * @generated from field: repeated mgmt.v1alpha1.GenerateSourceTableOption tables = 2;
     */
    tables: GenerateSourceTableOption[];
    constructor(data?: PartialMessage<GenerateSourceSchemaOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateSourceSchemaOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateSourceSchemaOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateSourceSchemaOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateSourceSchemaOption;
    static equals(a: GenerateSourceSchemaOption | PlainMessage<GenerateSourceSchemaOption> | undefined, b: GenerateSourceSchemaOption | PlainMessage<GenerateSourceSchemaOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GenerateSourceTableOption
 */
declare class GenerateSourceTableOption extends Message<GenerateSourceTableOption> {
    /**
     * @generated from field: string table = 1;
     */
    table: string;
    /**
     * @generated from field: int64 row_count = 2;
     */
    rowCount: bigint;
    constructor(data?: PartialMessage<GenerateSourceTableOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GenerateSourceTableOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GenerateSourceTableOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GenerateSourceTableOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GenerateSourceTableOption;
    static equals(a: GenerateSourceTableOption | PlainMessage<GenerateSourceTableOption> | undefined, b: GenerateSourceTableOption | PlainMessage<GenerateSourceTableOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresSourceConnectionOptions
 */
declare class PostgresSourceConnectionOptions extends Message<PostgresSourceConnectionOptions> {
    /**
     * @generated from field: bool halt_on_new_column_addition = 1;
     */
    haltOnNewColumnAddition: boolean;
    /**
     * @generated from field: repeated mgmt.v1alpha1.PostgresSourceSchemaOption schemas = 2;
     */
    schemas: PostgresSourceSchemaOption[];
    /**
     * @generated from field: string connection_id = 3;
     */
    connectionId: string;
    constructor(data?: PartialMessage<PostgresSourceConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresSourceConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresSourceConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresSourceConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresSourceConnectionOptions;
    static equals(a: PostgresSourceConnectionOptions | PlainMessage<PostgresSourceConnectionOptions> | undefined, b: PostgresSourceConnectionOptions | PlainMessage<PostgresSourceConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresSourceSchemaOption
 */
declare class PostgresSourceSchemaOption extends Message<PostgresSourceSchemaOption> {
    /**
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * @generated from field: repeated mgmt.v1alpha1.PostgresSourceTableOption tables = 2;
     */
    tables: PostgresSourceTableOption[];
    constructor(data?: PartialMessage<PostgresSourceSchemaOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresSourceSchemaOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresSourceSchemaOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresSourceSchemaOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresSourceSchemaOption;
    static equals(a: PostgresSourceSchemaOption | PlainMessage<PostgresSourceSchemaOption> | undefined, b: PostgresSourceSchemaOption | PlainMessage<PostgresSourceSchemaOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresSourceTableOption
 */
declare class PostgresSourceTableOption extends Message<PostgresSourceTableOption> {
    /**
     * @generated from field: string table = 1;
     */
    table: string;
    /**
     * @generated from field: optional string where_clause = 2;
     */
    whereClause?: string;
    constructor(data?: PartialMessage<PostgresSourceTableOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresSourceTableOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresSourceTableOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresSourceTableOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresSourceTableOption;
    static equals(a: PostgresSourceTableOption | PlainMessage<PostgresSourceTableOption> | undefined, b: PostgresSourceTableOption | PlainMessage<PostgresSourceTableOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlSourceConnectionOptions
 */
declare class MysqlSourceConnectionOptions extends Message<MysqlSourceConnectionOptions> {
    /**
     * @generated from field: bool halt_on_new_column_addition = 1;
     */
    haltOnNewColumnAddition: boolean;
    /**
     * @generated from field: repeated mgmt.v1alpha1.MysqlSourceSchemaOption schemas = 2;
     */
    schemas: MysqlSourceSchemaOption[];
    /**
     * @generated from field: string connection_id = 3;
     */
    connectionId: string;
    constructor(data?: PartialMessage<MysqlSourceConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlSourceConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlSourceConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlSourceConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlSourceConnectionOptions;
    static equals(a: MysqlSourceConnectionOptions | PlainMessage<MysqlSourceConnectionOptions> | undefined, b: MysqlSourceConnectionOptions | PlainMessage<MysqlSourceConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlSourceSchemaOption
 */
declare class MysqlSourceSchemaOption extends Message<MysqlSourceSchemaOption> {
    /**
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * @generated from field: repeated mgmt.v1alpha1.MysqlSourceTableOption tables = 2;
     */
    tables: MysqlSourceTableOption[];
    constructor(data?: PartialMessage<MysqlSourceSchemaOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlSourceSchemaOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlSourceSchemaOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlSourceSchemaOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlSourceSchemaOption;
    static equals(a: MysqlSourceSchemaOption | PlainMessage<MysqlSourceSchemaOption> | undefined, b: MysqlSourceSchemaOption | PlainMessage<MysqlSourceSchemaOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlSourceTableOption
 */
declare class MysqlSourceTableOption extends Message<MysqlSourceTableOption> {
    /**
     * @generated from field: string table = 1;
     */
    table: string;
    /**
     * @generated from field: optional string where_clause = 2;
     */
    whereClause?: string;
    constructor(data?: PartialMessage<MysqlSourceTableOption>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlSourceTableOption";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlSourceTableOption;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlSourceTableOption;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlSourceTableOption;
    static equals(a: MysqlSourceTableOption | PlainMessage<MysqlSourceTableOption> | undefined, b: MysqlSourceTableOption | PlainMessage<MysqlSourceTableOption> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AwsS3SourceConnectionOptions
 */
declare class AwsS3SourceConnectionOptions extends Message<AwsS3SourceConnectionOptions> {
    /**
     * @generated from field: string connection_id = 1;
     */
    connectionId: string;
    constructor(data?: PartialMessage<AwsS3SourceConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AwsS3SourceConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AwsS3SourceConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AwsS3SourceConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AwsS3SourceConnectionOptions;
    static equals(a: AwsS3SourceConnectionOptions | PlainMessage<AwsS3SourceConnectionOptions> | undefined, b: AwsS3SourceConnectionOptions | PlainMessage<AwsS3SourceConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobDestinationOptions
 */
declare class JobDestinationOptions extends Message<JobDestinationOptions> {
    /**
     * @generated from oneof mgmt.v1alpha1.JobDestinationOptions.config
     */
    config: {
        /**
         * @generated from field: mgmt.v1alpha1.PostgresDestinationConnectionOptions postgres_options = 1;
         */
        value: PostgresDestinationConnectionOptions;
        case: "postgresOptions";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.AwsS3DestinationConnectionOptions aws_s3_options = 2;
         */
        value: AwsS3DestinationConnectionOptions;
        case: "awsS3Options";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.MysqlDestinationConnectionOptions mysql_options = 3;
         */
        value: MysqlDestinationConnectionOptions;
        case: "mysqlOptions";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<JobDestinationOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobDestinationOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobDestinationOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobDestinationOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobDestinationOptions;
    static equals(a: JobDestinationOptions | PlainMessage<JobDestinationOptions> | undefined, b: JobDestinationOptions | PlainMessage<JobDestinationOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresDestinationConnectionOptions
 */
declare class PostgresDestinationConnectionOptions extends Message<PostgresDestinationConnectionOptions> {
    /**
     * @generated from field: mgmt.v1alpha1.PostgresTruncateTableConfig truncate_table = 1;
     */
    truncateTable?: PostgresTruncateTableConfig;
    /**
     * @generated from field: bool init_table_schema = 2;
     */
    initTableSchema: boolean;
    constructor(data?: PartialMessage<PostgresDestinationConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresDestinationConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresDestinationConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresDestinationConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresDestinationConnectionOptions;
    static equals(a: PostgresDestinationConnectionOptions | PlainMessage<PostgresDestinationConnectionOptions> | undefined, b: PostgresDestinationConnectionOptions | PlainMessage<PostgresDestinationConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresTruncateTableConfig
 */
declare class PostgresTruncateTableConfig extends Message<PostgresTruncateTableConfig> {
    /**
     * @generated from field: bool truncate_before_insert = 1;
     */
    truncateBeforeInsert: boolean;
    /**
     * @generated from field: bool cascade = 2;
     */
    cascade: boolean;
    constructor(data?: PartialMessage<PostgresTruncateTableConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresTruncateTableConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresTruncateTableConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresTruncateTableConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresTruncateTableConfig;
    static equals(a: PostgresTruncateTableConfig | PlainMessage<PostgresTruncateTableConfig> | undefined, b: PostgresTruncateTableConfig | PlainMessage<PostgresTruncateTableConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlDestinationConnectionOptions
 */
declare class MysqlDestinationConnectionOptions extends Message<MysqlDestinationConnectionOptions> {
    /**
     * @generated from field: mgmt.v1alpha1.MysqlTruncateTableConfig truncate_table = 1;
     */
    truncateTable?: MysqlTruncateTableConfig;
    /**
     * @generated from field: bool init_table_schema = 2;
     */
    initTableSchema: boolean;
    constructor(data?: PartialMessage<MysqlDestinationConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlDestinationConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlDestinationConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlDestinationConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlDestinationConnectionOptions;
    static equals(a: MysqlDestinationConnectionOptions | PlainMessage<MysqlDestinationConnectionOptions> | undefined, b: MysqlDestinationConnectionOptions | PlainMessage<MysqlDestinationConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlTruncateTableConfig
 */
declare class MysqlTruncateTableConfig extends Message<MysqlTruncateTableConfig> {
    /**
     * @generated from field: bool truncate_before_insert = 1;
     */
    truncateBeforeInsert: boolean;
    constructor(data?: PartialMessage<MysqlTruncateTableConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlTruncateTableConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlTruncateTableConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlTruncateTableConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlTruncateTableConfig;
    static equals(a: MysqlTruncateTableConfig | PlainMessage<MysqlTruncateTableConfig> | undefined, b: MysqlTruncateTableConfig | PlainMessage<MysqlTruncateTableConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AwsS3DestinationConnectionOptions
 */
declare class AwsS3DestinationConnectionOptions extends Message<AwsS3DestinationConnectionOptions> {
    constructor(data?: PartialMessage<AwsS3DestinationConnectionOptions>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AwsS3DestinationConnectionOptions";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AwsS3DestinationConnectionOptions;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AwsS3DestinationConnectionOptions;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AwsS3DestinationConnectionOptions;
    static equals(a: AwsS3DestinationConnectionOptions | PlainMessage<AwsS3DestinationConnectionOptions> | undefined, b: AwsS3DestinationConnectionOptions | PlainMessage<AwsS3DestinationConnectionOptions> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobRequest
 */
declare class CreateJobRequest extends Message<CreateJobRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string job_name = 2;
     */
    jobName: string;
    /**
     * @generated from field: optional string cron_schedule = 3;
     */
    cronSchedule?: string;
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 4;
     */
    mappings: JobMapping[];
    /**
     * @generated from field: mgmt.v1alpha1.JobSource source = 5;
     */
    source?: JobSource;
    /**
     * @generated from field: repeated mgmt.v1alpha1.CreateJobDestination destinations = 6;
     */
    destinations: CreateJobDestination[];
    /**
     * @generated from field: bool initiate_job_run = 7;
     */
    initiateJobRun: boolean;
    constructor(data?: PartialMessage<CreateJobRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobRequest;
    static equals(a: CreateJobRequest | PlainMessage<CreateJobRequest> | undefined, b: CreateJobRequest | PlainMessage<CreateJobRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobResponse
 */
declare class CreateJobResponse extends Message<CreateJobResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<CreateJobResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobResponse;
    static equals(a: CreateJobResponse | PlainMessage<CreateJobResponse> | undefined, b: CreateJobResponse | PlainMessage<CreateJobResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobMappingTransformer
 */
declare class JobMappingTransformer extends Message<JobMappingTransformer> {
    /**
     * @generated from field: string source = 1;
     */
    source: string;
    /**
     * @generated from field: mgmt.v1alpha1.TransformerConfig config = 3;
     */
    config?: TransformerConfig;
    constructor(data?: PartialMessage<JobMappingTransformer>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobMappingTransformer";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobMappingTransformer;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobMappingTransformer;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobMappingTransformer;
    static equals(a: JobMappingTransformer | PlainMessage<JobMappingTransformer> | undefined, b: JobMappingTransformer | PlainMessage<JobMappingTransformer> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobMapping
 */
declare class JobMapping extends Message<JobMapping> {
    /**
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * @generated from field: string table = 2;
     */
    table: string;
    /**
     * @generated from field: string column = 3;
     */
    column: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobMappingTransformer transformer = 5;
     */
    transformer?: JobMappingTransformer;
    constructor(data?: PartialMessage<JobMapping>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobMapping";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobMapping;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobMapping;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobMapping;
    static equals(a: JobMapping | PlainMessage<JobMapping> | undefined, b: JobMapping | PlainMessage<JobMapping> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRequest
 */
declare class GetJobRequest extends Message<GetJobRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<GetJobRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRequest;
    static equals(a: GetJobRequest | PlainMessage<GetJobRequest> | undefined, b: GetJobRequest | PlainMessage<GetJobRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobResponse
 */
declare class GetJobResponse extends Message<GetJobResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<GetJobResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobResponse;
    static equals(a: GetJobResponse | PlainMessage<GetJobResponse> | undefined, b: GetJobResponse | PlainMessage<GetJobResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobScheduleRequest
 */
declare class UpdateJobScheduleRequest extends Message<UpdateJobScheduleRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: optional string cron_schedule = 2;
     */
    cronSchedule?: string;
    constructor(data?: PartialMessage<UpdateJobScheduleRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobScheduleRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobScheduleRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobScheduleRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobScheduleRequest;
    static equals(a: UpdateJobScheduleRequest | PlainMessage<UpdateJobScheduleRequest> | undefined, b: UpdateJobScheduleRequest | PlainMessage<UpdateJobScheduleRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobScheduleResponse
 */
declare class UpdateJobScheduleResponse extends Message<UpdateJobScheduleResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<UpdateJobScheduleResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobScheduleResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobScheduleResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobScheduleResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobScheduleResponse;
    static equals(a: UpdateJobScheduleResponse | PlainMessage<UpdateJobScheduleResponse> | undefined, b: UpdateJobScheduleResponse | PlainMessage<UpdateJobScheduleResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PauseJobRequest
 */
declare class PauseJobRequest extends Message<PauseJobRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: bool pause = 2;
     */
    pause: boolean;
    /**
     * @generated from field: optional string note = 3;
     */
    note?: string;
    constructor(data?: PartialMessage<PauseJobRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PauseJobRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PauseJobRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PauseJobRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PauseJobRequest;
    static equals(a: PauseJobRequest | PlainMessage<PauseJobRequest> | undefined, b: PauseJobRequest | PlainMessage<PauseJobRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PauseJobResponse
 */
declare class PauseJobResponse extends Message<PauseJobResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<PauseJobResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PauseJobResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PauseJobResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PauseJobResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PauseJobResponse;
    static equals(a: PauseJobResponse | PlainMessage<PauseJobResponse> | undefined, b: PauseJobResponse | PlainMessage<PauseJobResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobSourceConnectionRequest
 */
declare class UpdateJobSourceConnectionRequest extends Message<UpdateJobSourceConnectionRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobSource source = 2;
     */
    source?: JobSource;
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 3;
     */
    mappings: JobMapping[];
    constructor(data?: PartialMessage<UpdateJobSourceConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobSourceConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobSourceConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobSourceConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobSourceConnectionRequest;
    static equals(a: UpdateJobSourceConnectionRequest | PlainMessage<UpdateJobSourceConnectionRequest> | undefined, b: UpdateJobSourceConnectionRequest | PlainMessage<UpdateJobSourceConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobSourceConnectionResponse
 */
declare class UpdateJobSourceConnectionResponse extends Message<UpdateJobSourceConnectionResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<UpdateJobSourceConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobSourceConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobSourceConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobSourceConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobSourceConnectionResponse;
    static equals(a: UpdateJobSourceConnectionResponse | PlainMessage<UpdateJobSourceConnectionResponse> | undefined, b: UpdateJobSourceConnectionResponse | PlainMessage<UpdateJobSourceConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PostgresSourceSchemaSubset
 */
declare class PostgresSourceSchemaSubset extends Message<PostgresSourceSchemaSubset> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.PostgresSourceSchemaOption postgres_schemas = 1;
     */
    postgresSchemas: PostgresSourceSchemaOption[];
    constructor(data?: PartialMessage<PostgresSourceSchemaSubset>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PostgresSourceSchemaSubset";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PostgresSourceSchemaSubset;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PostgresSourceSchemaSubset;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PostgresSourceSchemaSubset;
    static equals(a: PostgresSourceSchemaSubset | PlainMessage<PostgresSourceSchemaSubset> | undefined, b: PostgresSourceSchemaSubset | PlainMessage<PostgresSourceSchemaSubset> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.MysqlSourceSchemaSubset
 */
declare class MysqlSourceSchemaSubset extends Message<MysqlSourceSchemaSubset> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.MysqlSourceSchemaOption mysql_schemas = 1;
     */
    mysqlSchemas: MysqlSourceSchemaOption[];
    constructor(data?: PartialMessage<MysqlSourceSchemaSubset>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.MysqlSourceSchemaSubset";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): MysqlSourceSchemaSubset;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): MysqlSourceSchemaSubset;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): MysqlSourceSchemaSubset;
    static equals(a: MysqlSourceSchemaSubset | PlainMessage<MysqlSourceSchemaSubset> | undefined, b: MysqlSourceSchemaSubset | PlainMessage<MysqlSourceSchemaSubset> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobSourceSqlSubetSchemas
 */
declare class JobSourceSqlSubetSchemas extends Message<JobSourceSqlSubetSchemas> {
    /**
     * @generated from oneof mgmt.v1alpha1.JobSourceSqlSubetSchemas.schemas
     */
    schemas: {
        /**
         * @generated from field: mgmt.v1alpha1.PostgresSourceSchemaSubset postgres_subset = 2;
         */
        value: PostgresSourceSchemaSubset;
        case: "postgresSubset";
    } | {
        /**
         * @generated from field: mgmt.v1alpha1.MysqlSourceSchemaSubset mysql_subset = 3;
         */
        value: MysqlSourceSchemaSubset;
        case: "mysqlSubset";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<JobSourceSqlSubetSchemas>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobSourceSqlSubetSchemas";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobSourceSqlSubetSchemas;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobSourceSqlSubetSchemas;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobSourceSqlSubetSchemas;
    static equals(a: JobSourceSqlSubetSchemas | PlainMessage<JobSourceSqlSubetSchemas> | undefined, b: JobSourceSqlSubetSchemas | PlainMessage<JobSourceSqlSubetSchemas> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsRequest
 */
declare class SetJobSourceSqlConnectionSubsetsRequest extends Message<SetJobSourceSqlConnectionSubsetsRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobSourceSqlSubetSchemas schemas = 2;
     */
    schemas?: JobSourceSqlSubetSchemas;
    constructor(data?: PartialMessage<SetJobSourceSqlConnectionSubsetsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetJobSourceSqlConnectionSubsetsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetJobSourceSqlConnectionSubsetsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetJobSourceSqlConnectionSubsetsRequest;
    static equals(a: SetJobSourceSqlConnectionSubsetsRequest | PlainMessage<SetJobSourceSqlConnectionSubsetsRequest> | undefined, b: SetJobSourceSqlConnectionSubsetsRequest | PlainMessage<SetJobSourceSqlConnectionSubsetsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsResponse
 */
declare class SetJobSourceSqlConnectionSubsetsResponse extends Message<SetJobSourceSqlConnectionSubsetsResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<SetJobSourceSqlConnectionSubsetsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetJobSourceSqlConnectionSubsetsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetJobSourceSqlConnectionSubsetsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetJobSourceSqlConnectionSubsetsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetJobSourceSqlConnectionSubsetsResponse;
    static equals(a: SetJobSourceSqlConnectionSubsetsResponse | PlainMessage<SetJobSourceSqlConnectionSubsetsResponse> | undefined, b: SetJobSourceSqlConnectionSubsetsResponse | PlainMessage<SetJobSourceSqlConnectionSubsetsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobDestinationConnectionRequest
 */
declare class UpdateJobDestinationConnectionRequest extends Message<UpdateJobDestinationConnectionRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    /**
     * @generated from field: string connection_id = 2;
     */
    connectionId: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobDestinationOptions options = 3;
     */
    options?: JobDestinationOptions;
    /**
     * @generated from field: string destination_id = 4;
     */
    destinationId: string;
    constructor(data?: PartialMessage<UpdateJobDestinationConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobDestinationConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobDestinationConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobDestinationConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobDestinationConnectionRequest;
    static equals(a: UpdateJobDestinationConnectionRequest | PlainMessage<UpdateJobDestinationConnectionRequest> | undefined, b: UpdateJobDestinationConnectionRequest | PlainMessage<UpdateJobDestinationConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UpdateJobDestinationConnectionResponse
 */
declare class UpdateJobDestinationConnectionResponse extends Message<UpdateJobDestinationConnectionResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<UpdateJobDestinationConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UpdateJobDestinationConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateJobDestinationConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateJobDestinationConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateJobDestinationConnectionResponse;
    static equals(a: UpdateJobDestinationConnectionResponse | PlainMessage<UpdateJobDestinationConnectionResponse> | undefined, b: UpdateJobDestinationConnectionResponse | PlainMessage<UpdateJobDestinationConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobDestinationConnectionRequest
 */
declare class DeleteJobDestinationConnectionRequest extends Message<DeleteJobDestinationConnectionRequest> {
    /**
     * @generated from field: string destination_id = 1;
     */
    destinationId: string;
    constructor(data?: PartialMessage<DeleteJobDestinationConnectionRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobDestinationConnectionRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobDestinationConnectionRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobDestinationConnectionRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobDestinationConnectionRequest;
    static equals(a: DeleteJobDestinationConnectionRequest | PlainMessage<DeleteJobDestinationConnectionRequest> | undefined, b: DeleteJobDestinationConnectionRequest | PlainMessage<DeleteJobDestinationConnectionRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobDestinationConnectionResponse
 */
declare class DeleteJobDestinationConnectionResponse extends Message<DeleteJobDestinationConnectionResponse> {
    constructor(data?: PartialMessage<DeleteJobDestinationConnectionResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobDestinationConnectionResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobDestinationConnectionResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobDestinationConnectionResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobDestinationConnectionResponse;
    static equals(a: DeleteJobDestinationConnectionResponse | PlainMessage<DeleteJobDestinationConnectionResponse> | undefined, b: DeleteJobDestinationConnectionResponse | PlainMessage<DeleteJobDestinationConnectionResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobDestinationConnectionsRequest
 */
declare class CreateJobDestinationConnectionsRequest extends Message<CreateJobDestinationConnectionsRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    /**
     * @generated from field: repeated mgmt.v1alpha1.CreateJobDestination destinations = 2;
     */
    destinations: CreateJobDestination[];
    constructor(data?: PartialMessage<CreateJobDestinationConnectionsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobDestinationConnectionsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobDestinationConnectionsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobDestinationConnectionsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobDestinationConnectionsRequest;
    static equals(a: CreateJobDestinationConnectionsRequest | PlainMessage<CreateJobDestinationConnectionsRequest> | undefined, b: CreateJobDestinationConnectionsRequest | PlainMessage<CreateJobDestinationConnectionsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobDestinationConnectionsResponse
 */
declare class CreateJobDestinationConnectionsResponse extends Message<CreateJobDestinationConnectionsResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.Job job = 1;
     */
    job?: Job;
    constructor(data?: PartialMessage<CreateJobDestinationConnectionsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobDestinationConnectionsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobDestinationConnectionsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobDestinationConnectionsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobDestinationConnectionsResponse;
    static equals(a: CreateJobDestinationConnectionsResponse | PlainMessage<CreateJobDestinationConnectionsResponse> | undefined, b: CreateJobDestinationConnectionsResponse | PlainMessage<CreateJobDestinationConnectionsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobRequest
 */
declare class DeleteJobRequest extends Message<DeleteJobRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<DeleteJobRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobRequest;
    static equals(a: DeleteJobRequest | PlainMessage<DeleteJobRequest> | undefined, b: DeleteJobRequest | PlainMessage<DeleteJobRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobResponse
 */
declare class DeleteJobResponse extends Message<DeleteJobResponse> {
    constructor(data?: PartialMessage<DeleteJobResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobResponse;
    static equals(a: DeleteJobResponse | PlainMessage<DeleteJobResponse> | undefined, b: DeleteJobResponse | PlainMessage<DeleteJobResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsJobNameAvailableRequest
 */
declare class IsJobNameAvailableRequest extends Message<IsJobNameAvailableRequest> {
    /**
     * @generated from field: string name = 1;
     */
    name: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<IsJobNameAvailableRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsJobNameAvailableRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsJobNameAvailableRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsJobNameAvailableRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsJobNameAvailableRequest;
    static equals(a: IsJobNameAvailableRequest | PlainMessage<IsJobNameAvailableRequest> | undefined, b: IsJobNameAvailableRequest | PlainMessage<IsJobNameAvailableRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsJobNameAvailableResponse
 */
declare class IsJobNameAvailableResponse extends Message<IsJobNameAvailableResponse> {
    /**
     * @generated from field: bool is_available = 1;
     */
    isAvailable: boolean;
    constructor(data?: PartialMessage<IsJobNameAvailableResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsJobNameAvailableResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsJobNameAvailableResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsJobNameAvailableResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsJobNameAvailableResponse;
    static equals(a: IsJobNameAvailableResponse | PlainMessage<IsJobNameAvailableResponse> | undefined, b: IsJobNameAvailableResponse | PlainMessage<IsJobNameAvailableResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunsRequest
 */
declare class GetJobRunsRequest extends Message<GetJobRunsRequest> {
    /**
     * @generated from oneof mgmt.v1alpha1.GetJobRunsRequest.id
     */
    id: {
        /**
         * @generated from field: string job_id = 1;
         */
        value: string;
        case: "jobId";
    } | {
        /**
         * @generated from field: string account_id = 2;
         */
        value: string;
        case: "accountId";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<GetJobRunsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunsRequest;
    static equals(a: GetJobRunsRequest | PlainMessage<GetJobRunsRequest> | undefined, b: GetJobRunsRequest | PlainMessage<GetJobRunsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunsResponse
 */
declare class GetJobRunsResponse extends Message<GetJobRunsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobRun job_runs = 1;
     */
    jobRuns: JobRun[];
    constructor(data?: PartialMessage<GetJobRunsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunsResponse;
    static equals(a: GetJobRunsResponse | PlainMessage<GetJobRunsResponse> | undefined, b: GetJobRunsResponse | PlainMessage<GetJobRunsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunRequest
 */
declare class GetJobRunRequest extends Message<GetJobRunRequest> {
    /**
     * @generated from field: string job_run_id = 1;
     */
    jobRunId: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetJobRunRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunRequest;
    static equals(a: GetJobRunRequest | PlainMessage<GetJobRunRequest> | undefined, b: GetJobRunRequest | PlainMessage<GetJobRunRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunResponse
 */
declare class GetJobRunResponse extends Message<GetJobRunResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.JobRun job_run = 1;
     */
    jobRun?: JobRun;
    constructor(data?: PartialMessage<GetJobRunResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunResponse;
    static equals(a: GetJobRunResponse | PlainMessage<GetJobRunResponse> | undefined, b: GetJobRunResponse | PlainMessage<GetJobRunResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobRunRequest
 */
declare class CreateJobRunRequest extends Message<CreateJobRunRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    constructor(data?: PartialMessage<CreateJobRunRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobRunRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobRunRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobRunRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobRunRequest;
    static equals(a: CreateJobRunRequest | PlainMessage<CreateJobRunRequest> | undefined, b: CreateJobRunRequest | PlainMessage<CreateJobRunRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateJobRunResponse
 */
declare class CreateJobRunResponse extends Message<CreateJobRunResponse> {
    constructor(data?: PartialMessage<CreateJobRunResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateJobRunResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateJobRunResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateJobRunResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateJobRunResponse;
    static equals(a: CreateJobRunResponse | PlainMessage<CreateJobRunResponse> | undefined, b: CreateJobRunResponse | PlainMessage<CreateJobRunResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CancelJobRunRequest
 */
declare class CancelJobRunRequest extends Message<CancelJobRunRequest> {
    /**
     * @generated from field: string job_run_id = 1;
     */
    jobRunId: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<CancelJobRunRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CancelJobRunRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CancelJobRunRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CancelJobRunRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CancelJobRunRequest;
    static equals(a: CancelJobRunRequest | PlainMessage<CancelJobRunRequest> | undefined, b: CancelJobRunRequest | PlainMessage<CancelJobRunRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CancelJobRunResponse
 */
declare class CancelJobRunResponse extends Message<CancelJobRunResponse> {
    constructor(data?: PartialMessage<CancelJobRunResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CancelJobRunResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CancelJobRunResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CancelJobRunResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CancelJobRunResponse;
    static equals(a: CancelJobRunResponse | PlainMessage<CancelJobRunResponse> | undefined, b: CancelJobRunResponse | PlainMessage<CancelJobRunResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.Job
 */
declare class Job extends Message<Job> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string created_by_user_id = 2;
     */
    createdByUserId: string;
    /**
     * @generated from field: google.protobuf.Timestamp created_at = 3;
     */
    createdAt?: Timestamp;
    /**
     * @generated from field: string updated_by_user_id = 4;
     */
    updatedByUserId: string;
    /**
     * @generated from field: google.protobuf.Timestamp updated_at = 5;
     */
    updatedAt?: Timestamp;
    /**
     * @generated from field: string name = 6;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobSource source = 7;
     */
    source?: JobSource;
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobDestination destinations = 8;
     */
    destinations: JobDestination[];
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobMapping mappings = 9;
     */
    mappings: JobMapping[];
    /**
     * @generated from field: optional string cron_schedule = 10;
     */
    cronSchedule?: string;
    /**
     * @generated from field: string account_id = 11;
     */
    accountId: string;
    constructor(data?: PartialMessage<Job>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.Job";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Job;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Job;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Job;
    static equals(a: Job | PlainMessage<Job> | undefined, b: Job | PlainMessage<Job> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRecentRun
 */
declare class JobRecentRun extends Message<JobRecentRun> {
    /**
     * @generated from field: google.protobuf.Timestamp start_time = 1;
     */
    startTime?: Timestamp;
    /**
     * @generated from field: string job_run_id = 2;
     */
    jobRunId: string;
    constructor(data?: PartialMessage<JobRecentRun>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRecentRun";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRecentRun;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRecentRun;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRecentRun;
    static equals(a: JobRecentRun | PlainMessage<JobRecentRun> | undefined, b: JobRecentRun | PlainMessage<JobRecentRun> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRecentRunsRequest
 */
declare class GetJobRecentRunsRequest extends Message<GetJobRecentRunsRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    constructor(data?: PartialMessage<GetJobRecentRunsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRecentRunsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRecentRunsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRecentRunsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRecentRunsRequest;
    static equals(a: GetJobRecentRunsRequest | PlainMessage<GetJobRecentRunsRequest> | undefined, b: GetJobRecentRunsRequest | PlainMessage<GetJobRecentRunsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRecentRunsResponse
 */
declare class GetJobRecentRunsResponse extends Message<GetJobRecentRunsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobRecentRun recent_runs = 1;
     */
    recentRuns: JobRecentRun[];
    constructor(data?: PartialMessage<GetJobRecentRunsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRecentRunsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRecentRunsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRecentRunsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRecentRunsResponse;
    static equals(a: GetJobRecentRunsResponse | PlainMessage<GetJobRecentRunsResponse> | undefined, b: GetJobRecentRunsResponse | PlainMessage<GetJobRecentRunsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobNextRuns
 */
declare class JobNextRuns extends Message<JobNextRuns> {
    /**
     * @generated from field: repeated google.protobuf.Timestamp next_run_times = 1;
     */
    nextRunTimes: Timestamp[];
    constructor(data?: PartialMessage<JobNextRuns>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobNextRuns";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobNextRuns;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobNextRuns;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobNextRuns;
    static equals(a: JobNextRuns | PlainMessage<JobNextRuns> | undefined, b: JobNextRuns | PlainMessage<JobNextRuns> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobNextRunsRequest
 */
declare class GetJobNextRunsRequest extends Message<GetJobNextRunsRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    constructor(data?: PartialMessage<GetJobNextRunsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobNextRunsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobNextRunsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobNextRunsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobNextRunsRequest;
    static equals(a: GetJobNextRunsRequest | PlainMessage<GetJobNextRunsRequest> | undefined, b: GetJobNextRunsRequest | PlainMessage<GetJobNextRunsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobNextRunsResponse
 */
declare class GetJobNextRunsResponse extends Message<GetJobNextRunsResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.JobNextRuns next_runs = 1;
     */
    nextRuns?: JobNextRuns;
    constructor(data?: PartialMessage<GetJobNextRunsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobNextRunsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobNextRunsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobNextRunsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobNextRunsResponse;
    static equals(a: GetJobNextRunsResponse | PlainMessage<GetJobNextRunsResponse> | undefined, b: GetJobNextRunsResponse | PlainMessage<GetJobNextRunsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobStatusRequest
 */
declare class GetJobStatusRequest extends Message<GetJobStatusRequest> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    constructor(data?: PartialMessage<GetJobStatusRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobStatusRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobStatusRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobStatusRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobStatusRequest;
    static equals(a: GetJobStatusRequest | PlainMessage<GetJobStatusRequest> | undefined, b: GetJobStatusRequest | PlainMessage<GetJobStatusRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobStatusResponse
 */
declare class GetJobStatusResponse extends Message<GetJobStatusResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.JobStatus status = 1;
     */
    status: JobStatus;
    constructor(data?: PartialMessage<GetJobStatusResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobStatusResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobStatusResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobStatusResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobStatusResponse;
    static equals(a: GetJobStatusResponse | PlainMessage<GetJobStatusResponse> | undefined, b: GetJobStatusResponse | PlainMessage<GetJobStatusResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobStatusRecord
 */
declare class JobStatusRecord extends Message<JobStatusRecord> {
    /**
     * @generated from field: string job_id = 1;
     */
    jobId: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobStatus status = 2;
     */
    status: JobStatus;
    constructor(data?: PartialMessage<JobStatusRecord>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobStatusRecord";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobStatusRecord;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobStatusRecord;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobStatusRecord;
    static equals(a: JobStatusRecord | PlainMessage<JobStatusRecord> | undefined, b: JobStatusRecord | PlainMessage<JobStatusRecord> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobStatusesRequest
 */
declare class GetJobStatusesRequest extends Message<GetJobStatusesRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetJobStatusesRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobStatusesRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobStatusesRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobStatusesRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobStatusesRequest;
    static equals(a: GetJobStatusesRequest | PlainMessage<GetJobStatusesRequest> | undefined, b: GetJobStatusesRequest | PlainMessage<GetJobStatusesRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobStatusesResponse
 */
declare class GetJobStatusesResponse extends Message<GetJobStatusesResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobStatusRecord statuses = 1;
     */
    statuses: JobStatusRecord[];
    constructor(data?: PartialMessage<GetJobStatusesResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobStatusesResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobStatusesResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobStatusesResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobStatusesResponse;
    static equals(a: GetJobStatusesResponse | PlainMessage<GetJobStatusesResponse> | undefined, b: GetJobStatusesResponse | PlainMessage<GetJobStatusesResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.ActivityFailure
 */
declare class ActivityFailure extends Message<ActivityFailure> {
    /**
     * @generated from field: string message = 1;
     */
    message: string;
    constructor(data?: PartialMessage<ActivityFailure>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.ActivityFailure";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ActivityFailure;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ActivityFailure;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ActivityFailure;
    static equals(a: ActivityFailure | PlainMessage<ActivityFailure> | undefined, b: ActivityFailure | PlainMessage<ActivityFailure> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.PendingActivity
 */
declare class PendingActivity extends Message<PendingActivity> {
    /**
     * @generated from field: mgmt.v1alpha1.ActivityStatus status = 1;
     */
    status: ActivityStatus;
    /**
     * @generated from field: string activity_name = 2;
     */
    activityName: string;
    /**
     * @generated from field: optional mgmt.v1alpha1.ActivityFailure last_failure = 3;
     */
    lastFailure?: ActivityFailure;
    constructor(data?: PartialMessage<PendingActivity>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.PendingActivity";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PendingActivity;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PendingActivity;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PendingActivity;
    static equals(a: PendingActivity | PlainMessage<PendingActivity> | undefined, b: PendingActivity | PlainMessage<PendingActivity> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRun
 */
declare class JobRun extends Message<JobRun> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string job_id = 2;
     */
    jobId: string;
    /**
     * @generated from field: string name = 3;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.JobRunStatus status = 4;
     */
    status: JobRunStatus;
    /**
     * @generated from field: google.protobuf.Timestamp started_at = 6;
     */
    startedAt?: Timestamp;
    /**
     * @generated from field: optional google.protobuf.Timestamp completed_at = 7;
     */
    completedAt?: Timestamp;
    /**
     * @generated from field: repeated mgmt.v1alpha1.PendingActivity pending_activities = 8;
     */
    pendingActivities: PendingActivity[];
    constructor(data?: PartialMessage<JobRun>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRun";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRun;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRun;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRun;
    static equals(a: JobRun | PlainMessage<JobRun> | undefined, b: JobRun | PlainMessage<JobRun> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRunEventTaskError
 */
declare class JobRunEventTaskError extends Message<JobRunEventTaskError> {
    /**
     * @generated from field: string message = 1;
     */
    message: string;
    /**
     * @generated from field: string retry_state = 2;
     */
    retryState: string;
    constructor(data?: PartialMessage<JobRunEventTaskError>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRunEventTaskError";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRunEventTaskError;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRunEventTaskError;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRunEventTaskError;
    static equals(a: JobRunEventTaskError | PlainMessage<JobRunEventTaskError> | undefined, b: JobRunEventTaskError | PlainMessage<JobRunEventTaskError> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRunEventTask
 */
declare class JobRunEventTask extends Message<JobRunEventTask> {
    /**
     * @generated from field: int64 id = 1;
     */
    id: bigint;
    /**
     * @generated from field: string type = 2;
     */
    type: string;
    /**
     * @generated from field: google.protobuf.Timestamp event_time = 3;
     */
    eventTime?: Timestamp;
    /**
     * @generated from field: mgmt.v1alpha1.JobRunEventTaskError error = 4;
     */
    error?: JobRunEventTaskError;
    constructor(data?: PartialMessage<JobRunEventTask>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRunEventTask";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRunEventTask;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRunEventTask;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRunEventTask;
    static equals(a: JobRunEventTask | PlainMessage<JobRunEventTask> | undefined, b: JobRunEventTask | PlainMessage<JobRunEventTask> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRunSyncMetadata
 */
declare class JobRunSyncMetadata extends Message<JobRunSyncMetadata> {
    /**
     * @generated from field: string schema = 1;
     */
    schema: string;
    /**
     * @generated from field: string table = 2;
     */
    table: string;
    constructor(data?: PartialMessage<JobRunSyncMetadata>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRunSyncMetadata";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRunSyncMetadata;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRunSyncMetadata;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRunSyncMetadata;
    static equals(a: JobRunSyncMetadata | PlainMessage<JobRunSyncMetadata> | undefined, b: JobRunSyncMetadata | PlainMessage<JobRunSyncMetadata> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRunEventMetadata
 */
declare class JobRunEventMetadata extends Message<JobRunEventMetadata> {
    /**
     * @generated from oneof mgmt.v1alpha1.JobRunEventMetadata.metadata
     */
    metadata: {
        /**
         * @generated from field: mgmt.v1alpha1.JobRunSyncMetadata sync_metadata = 1;
         */
        value: JobRunSyncMetadata;
        case: "syncMetadata";
    } | {
        case: undefined;
        value?: undefined;
    };
    constructor(data?: PartialMessage<JobRunEventMetadata>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRunEventMetadata";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRunEventMetadata;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRunEventMetadata;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRunEventMetadata;
    static equals(a: JobRunEventMetadata | PlainMessage<JobRunEventMetadata> | undefined, b: JobRunEventMetadata | PlainMessage<JobRunEventMetadata> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.JobRunEvent
 */
declare class JobRunEvent extends Message<JobRunEvent> {
    /**
     * @generated from field: int64 id = 1;
     */
    id: bigint;
    /**
     * @generated from field: string type = 2;
     */
    type: string;
    /**
     * @generated from field: google.protobuf.Timestamp start_time = 3;
     */
    startTime?: Timestamp;
    /**
     * @generated from field: google.protobuf.Timestamp close_time = 4;
     */
    closeTime?: Timestamp;
    /**
     * @generated from field: mgmt.v1alpha1.JobRunEventMetadata metadata = 5;
     */
    metadata?: JobRunEventMetadata;
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobRunEventTask tasks = 6;
     */
    tasks: JobRunEventTask[];
    constructor(data?: PartialMessage<JobRunEvent>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.JobRunEvent";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): JobRunEvent;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): JobRunEvent;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): JobRunEvent;
    static equals(a: JobRunEvent | PlainMessage<JobRunEvent> | undefined, b: JobRunEvent | PlainMessage<JobRunEvent> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunEventsRequest
 */
declare class GetJobRunEventsRequest extends Message<GetJobRunEventsRequest> {
    /**
     * @generated from field: string job_run_id = 1;
     */
    jobRunId: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetJobRunEventsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunEventsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunEventsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunEventsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunEventsRequest;
    static equals(a: GetJobRunEventsRequest | PlainMessage<GetJobRunEventsRequest> | undefined, b: GetJobRunEventsRequest | PlainMessage<GetJobRunEventsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetJobRunEventsResponse
 */
declare class GetJobRunEventsResponse extends Message<GetJobRunEventsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.JobRunEvent events = 1;
     */
    events: JobRunEvent[];
    /**
     * @generated from field: bool is_run_complete = 2;
     */
    isRunComplete: boolean;
    constructor(data?: PartialMessage<GetJobRunEventsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetJobRunEventsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetJobRunEventsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetJobRunEventsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetJobRunEventsResponse;
    static equals(a: GetJobRunEventsResponse | PlainMessage<GetJobRunEventsResponse> | undefined, b: GetJobRunEventsResponse | PlainMessage<GetJobRunEventsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobRunRequest
 */
declare class DeleteJobRunRequest extends Message<DeleteJobRunRequest> {
    /**
     * @generated from field: string job_run_id = 1;
     */
    jobRunId: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<DeleteJobRunRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobRunRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobRunRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobRunRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobRunRequest;
    static equals(a: DeleteJobRunRequest | PlainMessage<DeleteJobRunRequest> | undefined, b: DeleteJobRunRequest | PlainMessage<DeleteJobRunRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.DeleteJobRunResponse
 */
declare class DeleteJobRunResponse extends Message<DeleteJobRunResponse> {
    constructor(data?: PartialMessage<DeleteJobRunResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.DeleteJobRunResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteJobRunResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteJobRunResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteJobRunResponse;
    static equals(a: DeleteJobRunResponse | PlainMessage<DeleteJobRunResponse> | undefined, b: DeleteJobRunResponse | PlainMessage<DeleteJobRunResponse> | undefined): boolean;
}

/**
 * @generated from service mgmt.v1alpha1.JobService
 */
declare const JobService: {
    readonly typeName: "mgmt.v1alpha1.JobService";
    readonly methods: {
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobs
         */
        readonly getJobs: {
            readonly name: "GetJobs";
            readonly I: typeof GetJobsRequest;
            readonly O: typeof GetJobsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJob
         */
        readonly getJob: {
            readonly name: "GetJob";
            readonly I: typeof GetJobRequest;
            readonly O: typeof GetJobResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.CreateJob
         */
        readonly createJob: {
            readonly name: "CreateJob";
            readonly I: typeof CreateJobRequest;
            readonly O: typeof CreateJobResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.DeleteJob
         */
        readonly deleteJob: {
            readonly name: "DeleteJob";
            readonly I: typeof DeleteJobRequest;
            readonly O: typeof DeleteJobResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.IsJobNameAvailable
         */
        readonly isJobNameAvailable: {
            readonly name: "IsJobNameAvailable";
            readonly I: typeof IsJobNameAvailableRequest;
            readonly O: typeof IsJobNameAvailableResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobSchedule
         */
        readonly updateJobSchedule: {
            readonly name: "UpdateJobSchedule";
            readonly I: typeof UpdateJobScheduleRequest;
            readonly O: typeof UpdateJobScheduleResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobSourceConnection
         */
        readonly updateJobSourceConnection: {
            readonly name: "UpdateJobSourceConnection";
            readonly I: typeof UpdateJobSourceConnectionRequest;
            readonly O: typeof UpdateJobSourceConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.SetJobSourceSqlConnectionSubsets
         */
        readonly setJobSourceSqlConnectionSubsets: {
            readonly name: "SetJobSourceSqlConnectionSubsets";
            readonly I: typeof SetJobSourceSqlConnectionSubsetsRequest;
            readonly O: typeof SetJobSourceSqlConnectionSubsetsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.UpdateJobDestinationConnection
         */
        readonly updateJobDestinationConnection: {
            readonly name: "UpdateJobDestinationConnection";
            readonly I: typeof UpdateJobDestinationConnectionRequest;
            readonly O: typeof UpdateJobDestinationConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.DeleteJobDestinationConnection
         */
        readonly deleteJobDestinationConnection: {
            readonly name: "DeleteJobDestinationConnection";
            readonly I: typeof DeleteJobDestinationConnectionRequest;
            readonly O: typeof DeleteJobDestinationConnectionResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.CreateJobDestinationConnections
         */
        readonly createJobDestinationConnections: {
            readonly name: "CreateJobDestinationConnections";
            readonly I: typeof CreateJobDestinationConnectionsRequest;
            readonly O: typeof CreateJobDestinationConnectionsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.PauseJob
         */
        readonly pauseJob: {
            readonly name: "PauseJob";
            readonly I: typeof PauseJobRequest;
            readonly O: typeof PauseJobResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobRecentRuns
         */
        readonly getJobRecentRuns: {
            readonly name: "GetJobRecentRuns";
            readonly I: typeof GetJobRecentRunsRequest;
            readonly O: typeof GetJobRecentRunsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobNextRuns
         */
        readonly getJobNextRuns: {
            readonly name: "GetJobNextRuns";
            readonly I: typeof GetJobNextRunsRequest;
            readonly O: typeof GetJobNextRunsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobStatus
         */
        readonly getJobStatus: {
            readonly name: "GetJobStatus";
            readonly I: typeof GetJobStatusRequest;
            readonly O: typeof GetJobStatusResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobStatuses
         */
        readonly getJobStatuses: {
            readonly name: "GetJobStatuses";
            readonly I: typeof GetJobStatusesRequest;
            readonly O: typeof GetJobStatusesResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobRuns
         */
        readonly getJobRuns: {
            readonly name: "GetJobRuns";
            readonly I: typeof GetJobRunsRequest;
            readonly O: typeof GetJobRunsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobRunEvents
         */
        readonly getJobRunEvents: {
            readonly name: "GetJobRunEvents";
            readonly I: typeof GetJobRunEventsRequest;
            readonly O: typeof GetJobRunEventsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.GetJobRun
         */
        readonly getJobRun: {
            readonly name: "GetJobRun";
            readonly I: typeof GetJobRunRequest;
            readonly O: typeof GetJobRunResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.DeleteJobRun
         */
        readonly deleteJobRun: {
            readonly name: "DeleteJobRun";
            readonly I: typeof DeleteJobRunRequest;
            readonly O: typeof DeleteJobRunResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.CreateJobRun
         */
        readonly createJobRun: {
            readonly name: "CreateJobRun";
            readonly I: typeof CreateJobRunRequest;
            readonly O: typeof CreateJobRunResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.JobService.CancelJobRun
         */
        readonly cancelJobRun: {
            readonly name: "CancelJobRun";
            readonly I: typeof CancelJobRunRequest;
            readonly O: typeof CancelJobRunResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};

/**
 * @generated from service mgmt.v1alpha1.TransformersService
 */
declare const TransformersService: {
    readonly typeName: "mgmt.v1alpha1.TransformersService";
    readonly methods: {
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.GetSystemTransformers
         */
        readonly getSystemTransformers: {
            readonly name: "GetSystemTransformers";
            readonly I: typeof GetSystemTransformersRequest;
            readonly O: typeof GetSystemTransformersResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.GetUserDefinedTransformers
         */
        readonly getUserDefinedTransformers: {
            readonly name: "GetUserDefinedTransformers";
            readonly I: typeof GetUserDefinedTransformersRequest;
            readonly O: typeof GetUserDefinedTransformersResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.GetUserDefinedTransformerById
         */
        readonly getUserDefinedTransformerById: {
            readonly name: "GetUserDefinedTransformerById";
            readonly I: typeof GetUserDefinedTransformerByIdRequest;
            readonly O: typeof GetUserDefinedTransformerByIdResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.CreateUserDefinedTransformer
         */
        readonly createUserDefinedTransformer: {
            readonly name: "CreateUserDefinedTransformer";
            readonly I: typeof CreateUserDefinedTransformerRequest;
            readonly O: typeof CreateUserDefinedTransformerResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.DeleteUserDefinedTransformer
         */
        readonly deleteUserDefinedTransformer: {
            readonly name: "DeleteUserDefinedTransformer";
            readonly I: typeof DeleteUserDefinedTransformerRequest;
            readonly O: typeof DeleteUserDefinedTransformerResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.UpdateUserDefinedTransformer
         */
        readonly updateUserDefinedTransformer: {
            readonly name: "UpdateUserDefinedTransformer";
            readonly I: typeof UpdateUserDefinedTransformerRequest;
            readonly O: typeof UpdateUserDefinedTransformerResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.TransformersService.IsTransformerNameAvailable
         */
        readonly isTransformerNameAvailable: {
            readonly name: "IsTransformerNameAvailable";
            readonly I: typeof IsTransformerNameAvailableRequest;
            readonly O: typeof IsTransformerNameAvailableResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};

/**
 * @generated from enum mgmt.v1alpha1.UserAccountType
 */
declare enum UserAccountType {
    /**
     * @generated from enum value: USER_ACCOUNT_TYPE_UNSPECIFIED = 0;
     */
    UNSPECIFIED = 0,
    /**
     * @generated from enum value: USER_ACCOUNT_TYPE_PERSONAL = 1;
     */
    PERSONAL = 1,
    /**
     * @generated from enum value: USER_ACCOUNT_TYPE_TEAM = 2;
     */
    TEAM = 2
}
/**
 * @generated from message mgmt.v1alpha1.GetUserRequest
 */
declare class GetUserRequest extends Message<GetUserRequest> {
    constructor(data?: PartialMessage<GetUserRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserRequest;
    static equals(a: GetUserRequest | PlainMessage<GetUserRequest> | undefined, b: GetUserRequest | PlainMessage<GetUserRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserResponse
 */
declare class GetUserResponse extends Message<GetUserResponse> {
    /**
     * @generated from field: string user_id = 1;
     */
    userId: string;
    constructor(data?: PartialMessage<GetUserResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserResponse;
    static equals(a: GetUserResponse | PlainMessage<GetUserResponse> | undefined, b: GetUserResponse | PlainMessage<GetUserResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetUserRequest
 */
declare class SetUserRequest extends Message<SetUserRequest> {
    constructor(data?: PartialMessage<SetUserRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetUserRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetUserRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetUserRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetUserRequest;
    static equals(a: SetUserRequest | PlainMessage<SetUserRequest> | undefined, b: SetUserRequest | PlainMessage<SetUserRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetUserResponse
 */
declare class SetUserResponse extends Message<SetUserResponse> {
    /**
     * @generated from field: string user_id = 1;
     */
    userId: string;
    constructor(data?: PartialMessage<SetUserResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetUserResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetUserResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetUserResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetUserResponse;
    static equals(a: SetUserResponse | PlainMessage<SetUserResponse> | undefined, b: SetUserResponse | PlainMessage<SetUserResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserAccountsRequest
 */
declare class GetUserAccountsRequest extends Message<GetUserAccountsRequest> {
    constructor(data?: PartialMessage<GetUserAccountsRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserAccountsRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserAccountsRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserAccountsRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserAccountsRequest;
    static equals(a: GetUserAccountsRequest | PlainMessage<GetUserAccountsRequest> | undefined, b: GetUserAccountsRequest | PlainMessage<GetUserAccountsRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetUserAccountsResponse
 */
declare class GetUserAccountsResponse extends Message<GetUserAccountsResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.UserAccount accounts = 1;
     */
    accounts: UserAccount[];
    constructor(data?: PartialMessage<GetUserAccountsResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetUserAccountsResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetUserAccountsResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetUserAccountsResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetUserAccountsResponse;
    static equals(a: GetUserAccountsResponse | PlainMessage<GetUserAccountsResponse> | undefined, b: GetUserAccountsResponse | PlainMessage<GetUserAccountsResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.UserAccount
 */
declare class UserAccount extends Message<UserAccount> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: mgmt.v1alpha1.UserAccountType type = 3;
     */
    type: UserAccountType;
    constructor(data?: PartialMessage<UserAccount>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.UserAccount";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UserAccount;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UserAccount;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UserAccount;
    static equals(a: UserAccount | PlainMessage<UserAccount> | undefined, b: UserAccount | PlainMessage<UserAccount> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.ConvertPersonalToTeamAccountRequest
 */
declare class ConvertPersonalToTeamAccountRequest extends Message<ConvertPersonalToTeamAccountRequest> {
    constructor(data?: PartialMessage<ConvertPersonalToTeamAccountRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.ConvertPersonalToTeamAccountRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConvertPersonalToTeamAccountRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConvertPersonalToTeamAccountRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConvertPersonalToTeamAccountRequest;
    static equals(a: ConvertPersonalToTeamAccountRequest | PlainMessage<ConvertPersonalToTeamAccountRequest> | undefined, b: ConvertPersonalToTeamAccountRequest | PlainMessage<ConvertPersonalToTeamAccountRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.ConvertPersonalToTeamAccountResponse
 */
declare class ConvertPersonalToTeamAccountResponse extends Message<ConvertPersonalToTeamAccountResponse> {
    constructor(data?: PartialMessage<ConvertPersonalToTeamAccountResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.ConvertPersonalToTeamAccountResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ConvertPersonalToTeamAccountResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ConvertPersonalToTeamAccountResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ConvertPersonalToTeamAccountResponse;
    static equals(a: ConvertPersonalToTeamAccountResponse | PlainMessage<ConvertPersonalToTeamAccountResponse> | undefined, b: ConvertPersonalToTeamAccountResponse | PlainMessage<ConvertPersonalToTeamAccountResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetPersonalAccountRequest
 */
declare class SetPersonalAccountRequest extends Message<SetPersonalAccountRequest> {
    constructor(data?: PartialMessage<SetPersonalAccountRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetPersonalAccountRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetPersonalAccountRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetPersonalAccountRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetPersonalAccountRequest;
    static equals(a: SetPersonalAccountRequest | PlainMessage<SetPersonalAccountRequest> | undefined, b: SetPersonalAccountRequest | PlainMessage<SetPersonalAccountRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetPersonalAccountResponse
 */
declare class SetPersonalAccountResponse extends Message<SetPersonalAccountResponse> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<SetPersonalAccountResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetPersonalAccountResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetPersonalAccountResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetPersonalAccountResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetPersonalAccountResponse;
    static equals(a: SetPersonalAccountResponse | PlainMessage<SetPersonalAccountResponse> | undefined, b: SetPersonalAccountResponse | PlainMessage<SetPersonalAccountResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsUserInAccountRequest
 */
declare class IsUserInAccountRequest extends Message<IsUserInAccountRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<IsUserInAccountRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsUserInAccountRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsUserInAccountRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsUserInAccountRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsUserInAccountRequest;
    static equals(a: IsUserInAccountRequest | PlainMessage<IsUserInAccountRequest> | undefined, b: IsUserInAccountRequest | PlainMessage<IsUserInAccountRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.IsUserInAccountResponse
 */
declare class IsUserInAccountResponse extends Message<IsUserInAccountResponse> {
    /**
     * @generated from field: bool ok = 1;
     */
    ok: boolean;
    constructor(data?: PartialMessage<IsUserInAccountResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.IsUserInAccountResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): IsUserInAccountResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): IsUserInAccountResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): IsUserInAccountResponse;
    static equals(a: IsUserInAccountResponse | PlainMessage<IsUserInAccountResponse> | undefined, b: IsUserInAccountResponse | PlainMessage<IsUserInAccountResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountTemporalConfigRequest
 */
declare class GetAccountTemporalConfigRequest extends Message<GetAccountTemporalConfigRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetAccountTemporalConfigRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountTemporalConfigRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountTemporalConfigRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountTemporalConfigRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountTemporalConfigRequest;
    static equals(a: GetAccountTemporalConfigRequest | PlainMessage<GetAccountTemporalConfigRequest> | undefined, b: GetAccountTemporalConfigRequest | PlainMessage<GetAccountTemporalConfigRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetAccountTemporalConfigResponse
 */
declare class GetAccountTemporalConfigResponse extends Message<GetAccountTemporalConfigResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 1;
     */
    config?: AccountTemporalConfig;
    constructor(data?: PartialMessage<GetAccountTemporalConfigResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetAccountTemporalConfigResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetAccountTemporalConfigResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetAccountTemporalConfigResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetAccountTemporalConfigResponse;
    static equals(a: GetAccountTemporalConfigResponse | PlainMessage<GetAccountTemporalConfigResponse> | undefined, b: GetAccountTemporalConfigResponse | PlainMessage<GetAccountTemporalConfigResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetAccountTemporalConfigRequest
 */
declare class SetAccountTemporalConfigRequest extends Message<SetAccountTemporalConfigRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 2;
     */
    config?: AccountTemporalConfig;
    constructor(data?: PartialMessage<SetAccountTemporalConfigRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetAccountTemporalConfigRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetAccountTemporalConfigRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetAccountTemporalConfigRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetAccountTemporalConfigRequest;
    static equals(a: SetAccountTemporalConfigRequest | PlainMessage<SetAccountTemporalConfigRequest> | undefined, b: SetAccountTemporalConfigRequest | PlainMessage<SetAccountTemporalConfigRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.SetAccountTemporalConfigResponse
 */
declare class SetAccountTemporalConfigResponse extends Message<SetAccountTemporalConfigResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountTemporalConfig config = 1;
     */
    config?: AccountTemporalConfig;
    constructor(data?: PartialMessage<SetAccountTemporalConfigResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.SetAccountTemporalConfigResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): SetAccountTemporalConfigResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): SetAccountTemporalConfigResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): SetAccountTemporalConfigResponse;
    static equals(a: SetAccountTemporalConfigResponse | PlainMessage<SetAccountTemporalConfigResponse> | undefined, b: SetAccountTemporalConfigResponse | PlainMessage<SetAccountTemporalConfigResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AccountTemporalConfig
 */
declare class AccountTemporalConfig extends Message<AccountTemporalConfig> {
    /**
     * @generated from field: string url = 1;
     */
    url: string;
    /**
     * @generated from field: string namespace = 2;
     */
    namespace: string;
    /**
     * @generated from field: string sync_job_queue_name = 3;
     */
    syncJobQueueName: string;
    constructor(data?: PartialMessage<AccountTemporalConfig>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AccountTemporalConfig";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AccountTemporalConfig;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AccountTemporalConfig;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AccountTemporalConfig;
    static equals(a: AccountTemporalConfig | PlainMessage<AccountTemporalConfig> | undefined, b: AccountTemporalConfig | PlainMessage<AccountTemporalConfig> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateTeamAccountRequest
 */
declare class CreateTeamAccountRequest extends Message<CreateTeamAccountRequest> {
    /**
     * @generated from field: string name = 1;
     */
    name: string;
    constructor(data?: PartialMessage<CreateTeamAccountRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateTeamAccountRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateTeamAccountRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateTeamAccountRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateTeamAccountRequest;
    static equals(a: CreateTeamAccountRequest | PlainMessage<CreateTeamAccountRequest> | undefined, b: CreateTeamAccountRequest | PlainMessage<CreateTeamAccountRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.CreateTeamAccountResponse
 */
declare class CreateTeamAccountResponse extends Message<CreateTeamAccountResponse> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<CreateTeamAccountResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.CreateTeamAccountResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateTeamAccountResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateTeamAccountResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateTeamAccountResponse;
    static equals(a: CreateTeamAccountResponse | PlainMessage<CreateTeamAccountResponse> | undefined, b: CreateTeamAccountResponse | PlainMessage<CreateTeamAccountResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AccountUser
 */
declare class AccountUser extends Message<AccountUser> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string name = 2;
     */
    name: string;
    /**
     * @generated from field: string image = 3;
     */
    image: string;
    /**
     * @generated from field: string email = 4;
     */
    email: string;
    constructor(data?: PartialMessage<AccountUser>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AccountUser";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AccountUser;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AccountUser;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AccountUser;
    static equals(a: AccountUser | PlainMessage<AccountUser> | undefined, b: AccountUser | PlainMessage<AccountUser> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetTeamAccountMembersRequest
 */
declare class GetTeamAccountMembersRequest extends Message<GetTeamAccountMembersRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetTeamAccountMembersRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetTeamAccountMembersRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTeamAccountMembersRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTeamAccountMembersRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTeamAccountMembersRequest;
    static equals(a: GetTeamAccountMembersRequest | PlainMessage<GetTeamAccountMembersRequest> | undefined, b: GetTeamAccountMembersRequest | PlainMessage<GetTeamAccountMembersRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetTeamAccountMembersResponse
 */
declare class GetTeamAccountMembersResponse extends Message<GetTeamAccountMembersResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.AccountUser users = 1;
     */
    users: AccountUser[];
    constructor(data?: PartialMessage<GetTeamAccountMembersResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetTeamAccountMembersResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTeamAccountMembersResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTeamAccountMembersResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTeamAccountMembersResponse;
    static equals(a: GetTeamAccountMembersResponse | PlainMessage<GetTeamAccountMembersResponse> | undefined, b: GetTeamAccountMembersResponse | PlainMessage<GetTeamAccountMembersResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RemoveTeamAccountMemberRequest
 */
declare class RemoveTeamAccountMemberRequest extends Message<RemoveTeamAccountMemberRequest> {
    /**
     * @generated from field: string user_id = 1;
     */
    userId: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    constructor(data?: PartialMessage<RemoveTeamAccountMemberRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RemoveTeamAccountMemberRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveTeamAccountMemberRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveTeamAccountMemberRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveTeamAccountMemberRequest;
    static equals(a: RemoveTeamAccountMemberRequest | PlainMessage<RemoveTeamAccountMemberRequest> | undefined, b: RemoveTeamAccountMemberRequest | PlainMessage<RemoveTeamAccountMemberRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RemoveTeamAccountMemberResponse
 */
declare class RemoveTeamAccountMemberResponse extends Message<RemoveTeamAccountMemberResponse> {
    constructor(data?: PartialMessage<RemoveTeamAccountMemberResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RemoveTeamAccountMemberResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveTeamAccountMemberResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveTeamAccountMemberResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveTeamAccountMemberResponse;
    static equals(a: RemoveTeamAccountMemberResponse | PlainMessage<RemoveTeamAccountMemberResponse> | undefined, b: RemoveTeamAccountMemberResponse | PlainMessage<RemoveTeamAccountMemberResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.InviteUserToTeamAccountRequest
 */
declare class InviteUserToTeamAccountRequest extends Message<InviteUserToTeamAccountRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    /**
     * @generated from field: string email = 2;
     */
    email: string;
    constructor(data?: PartialMessage<InviteUserToTeamAccountRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.InviteUserToTeamAccountRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): InviteUserToTeamAccountRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): InviteUserToTeamAccountRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): InviteUserToTeamAccountRequest;
    static equals(a: InviteUserToTeamAccountRequest | PlainMessage<InviteUserToTeamAccountRequest> | undefined, b: InviteUserToTeamAccountRequest | PlainMessage<InviteUserToTeamAccountRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AccountInvite
 */
declare class AccountInvite extends Message<AccountInvite> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    /**
     * @generated from field: string account_id = 2;
     */
    accountId: string;
    /**
     * @generated from field: string sender_user_id = 3;
     */
    senderUserId: string;
    /**
     * @generated from field: string email = 4;
     */
    email: string;
    /**
     * @generated from field: string token = 5;
     */
    token: string;
    /**
     * @generated from field: bool accepted = 6;
     */
    accepted: boolean;
    /**
     * @generated from field: google.protobuf.Timestamp created_at = 7;
     */
    createdAt?: Timestamp;
    /**
     * @generated from field: google.protobuf.Timestamp updated_at = 8;
     */
    updatedAt?: Timestamp;
    /**
     * @generated from field: google.protobuf.Timestamp expires_at = 9;
     */
    expiresAt?: Timestamp;
    constructor(data?: PartialMessage<AccountInvite>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AccountInvite";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AccountInvite;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AccountInvite;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AccountInvite;
    static equals(a: AccountInvite | PlainMessage<AccountInvite> | undefined, b: AccountInvite | PlainMessage<AccountInvite> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.InviteUserToTeamAccountResponse
 */
declare class InviteUserToTeamAccountResponse extends Message<InviteUserToTeamAccountResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.AccountInvite invite = 1;
     */
    invite?: AccountInvite;
    constructor(data?: PartialMessage<InviteUserToTeamAccountResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.InviteUserToTeamAccountResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): InviteUserToTeamAccountResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): InviteUserToTeamAccountResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): InviteUserToTeamAccountResponse;
    static equals(a: InviteUserToTeamAccountResponse | PlainMessage<InviteUserToTeamAccountResponse> | undefined, b: InviteUserToTeamAccountResponse | PlainMessage<InviteUserToTeamAccountResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetTeamAccountInvitesRequest
 */
declare class GetTeamAccountInvitesRequest extends Message<GetTeamAccountInvitesRequest> {
    /**
     * @generated from field: string account_id = 1;
     */
    accountId: string;
    constructor(data?: PartialMessage<GetTeamAccountInvitesRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetTeamAccountInvitesRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTeamAccountInvitesRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTeamAccountInvitesRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTeamAccountInvitesRequest;
    static equals(a: GetTeamAccountInvitesRequest | PlainMessage<GetTeamAccountInvitesRequest> | undefined, b: GetTeamAccountInvitesRequest | PlainMessage<GetTeamAccountInvitesRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.GetTeamAccountInvitesResponse
 */
declare class GetTeamAccountInvitesResponse extends Message<GetTeamAccountInvitesResponse> {
    /**
     * @generated from field: repeated mgmt.v1alpha1.AccountInvite invites = 1;
     */
    invites: AccountInvite[];
    constructor(data?: PartialMessage<GetTeamAccountInvitesResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.GetTeamAccountInvitesResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetTeamAccountInvitesResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetTeamAccountInvitesResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetTeamAccountInvitesResponse;
    static equals(a: GetTeamAccountInvitesResponse | PlainMessage<GetTeamAccountInvitesResponse> | undefined, b: GetTeamAccountInvitesResponse | PlainMessage<GetTeamAccountInvitesResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RemoveTeamAccountInviteRequest
 */
declare class RemoveTeamAccountInviteRequest extends Message<RemoveTeamAccountInviteRequest> {
    /**
     * @generated from field: string id = 1;
     */
    id: string;
    constructor(data?: PartialMessage<RemoveTeamAccountInviteRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RemoveTeamAccountInviteRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveTeamAccountInviteRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveTeamAccountInviteRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveTeamAccountInviteRequest;
    static equals(a: RemoveTeamAccountInviteRequest | PlainMessage<RemoveTeamAccountInviteRequest> | undefined, b: RemoveTeamAccountInviteRequest | PlainMessage<RemoveTeamAccountInviteRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.RemoveTeamAccountInviteResponse
 */
declare class RemoveTeamAccountInviteResponse extends Message<RemoveTeamAccountInviteResponse> {
    constructor(data?: PartialMessage<RemoveTeamAccountInviteResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.RemoveTeamAccountInviteResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): RemoveTeamAccountInviteResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): RemoveTeamAccountInviteResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): RemoveTeamAccountInviteResponse;
    static equals(a: RemoveTeamAccountInviteResponse | PlainMessage<RemoveTeamAccountInviteResponse> | undefined, b: RemoveTeamAccountInviteResponse | PlainMessage<RemoveTeamAccountInviteResponse> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AcceptTeamAccountInviteRequest
 */
declare class AcceptTeamAccountInviteRequest extends Message<AcceptTeamAccountInviteRequest> {
    /**
     * @generated from field: string token = 1;
     */
    token: string;
    constructor(data?: PartialMessage<AcceptTeamAccountInviteRequest>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AcceptTeamAccountInviteRequest";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AcceptTeamAccountInviteRequest;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AcceptTeamAccountInviteRequest;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AcceptTeamAccountInviteRequest;
    static equals(a: AcceptTeamAccountInviteRequest | PlainMessage<AcceptTeamAccountInviteRequest> | undefined, b: AcceptTeamAccountInviteRequest | PlainMessage<AcceptTeamAccountInviteRequest> | undefined): boolean;
}
/**
 * @generated from message mgmt.v1alpha1.AcceptTeamAccountInviteResponse
 */
declare class AcceptTeamAccountInviteResponse extends Message<AcceptTeamAccountInviteResponse> {
    /**
     * @generated from field: mgmt.v1alpha1.UserAccount account = 1;
     */
    account?: UserAccount;
    constructor(data?: PartialMessage<AcceptTeamAccountInviteResponse>);
    static readonly runtime: typeof proto3;
    static readonly typeName = "mgmt.v1alpha1.AcceptTeamAccountInviteResponse";
    static readonly fields: FieldList;
    static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): AcceptTeamAccountInviteResponse;
    static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): AcceptTeamAccountInviteResponse;
    static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): AcceptTeamAccountInviteResponse;
    static equals(a: AcceptTeamAccountInviteResponse | PlainMessage<AcceptTeamAccountInviteResponse> | undefined, b: AcceptTeamAccountInviteResponse | PlainMessage<AcceptTeamAccountInviteResponse> | undefined): boolean;
}

/**
 * @generated from service mgmt.v1alpha1.UserAccountService
 */
declare const UserAccountService: {
    readonly typeName: "mgmt.v1alpha1.UserAccountService";
    readonly methods: {
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUser
         */
        readonly getUser: {
            readonly name: "GetUser";
            readonly I: typeof GetUserRequest;
            readonly O: typeof GetUserResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.SetUser
         */
        readonly setUser: {
            readonly name: "SetUser";
            readonly I: typeof SetUserRequest;
            readonly O: typeof SetUserResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.GetUserAccounts
         */
        readonly getUserAccounts: {
            readonly name: "GetUserAccounts";
            readonly I: typeof GetUserAccountsRequest;
            readonly O: typeof GetUserAccountsResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.SetPersonalAccount
         */
        readonly setPersonalAccount: {
            readonly name: "SetPersonalAccount";
            readonly I: typeof SetPersonalAccountRequest;
            readonly O: typeof SetPersonalAccountResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.ConvertPersonalToTeamAccount
         */
        readonly convertPersonalToTeamAccount: {
            readonly name: "ConvertPersonalToTeamAccount";
            readonly I: typeof ConvertPersonalToTeamAccountRequest;
            readonly O: typeof ConvertPersonalToTeamAccountResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.CreateTeamAccount
         */
        readonly createTeamAccount: {
            readonly name: "CreateTeamAccount";
            readonly I: typeof CreateTeamAccountRequest;
            readonly O: typeof CreateTeamAccountResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.IsUserInAccount
         */
        readonly isUserInAccount: {
            readonly name: "IsUserInAccount";
            readonly I: typeof IsUserInAccountRequest;
            readonly O: typeof IsUserInAccountResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.GetAccountTemporalConfig
         */
        readonly getAccountTemporalConfig: {
            readonly name: "GetAccountTemporalConfig";
            readonly I: typeof GetAccountTemporalConfigRequest;
            readonly O: typeof GetAccountTemporalConfigResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.SetAccountTemporalConfig
         */
        readonly setAccountTemporalConfig: {
            readonly name: "SetAccountTemporalConfig";
            readonly I: typeof SetAccountTemporalConfigRequest;
            readonly O: typeof SetAccountTemporalConfigResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountMembers
         */
        readonly getTeamAccountMembers: {
            readonly name: "GetTeamAccountMembers";
            readonly I: typeof GetTeamAccountMembersRequest;
            readonly O: typeof GetTeamAccountMembersResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountMember
         */
        readonly removeTeamAccountMember: {
            readonly name: "RemoveTeamAccountMember";
            readonly I: typeof RemoveTeamAccountMemberRequest;
            readonly O: typeof RemoveTeamAccountMemberResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.InviteUserToTeamAccount
         */
        readonly inviteUserToTeamAccount: {
            readonly name: "InviteUserToTeamAccount";
            readonly I: typeof InviteUserToTeamAccountRequest;
            readonly O: typeof InviteUserToTeamAccountResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.GetTeamAccountInvites
         */
        readonly getTeamAccountInvites: {
            readonly name: "GetTeamAccountInvites";
            readonly I: typeof GetTeamAccountInvitesRequest;
            readonly O: typeof GetTeamAccountInvitesResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.RemoveTeamAccountInvite
         */
        readonly removeTeamAccountInvite: {
            readonly name: "RemoveTeamAccountInvite";
            readonly I: typeof RemoveTeamAccountInviteRequest;
            readonly O: typeof RemoveTeamAccountInviteResponse;
            readonly kind: MethodKind.Unary;
        };
        /**
         * @generated from rpc mgmt.v1alpha1.UserAccountService.AcceptTeamAccountInvite
         */
        readonly acceptTeamAccountInvite: {
            readonly name: "AcceptTeamAccountInvite";
            readonly I: typeof AcceptTeamAccountInviteRequest;
            readonly O: typeof AcceptTeamAccountInviteResponse;
            readonly kind: MethodKind.Unary;
        };
    };
};

type NeosyncClient = NeosyncV1alpha1Client;
type ClientVersion = "v1alpha1" | "latest";
interface NeosyncV1alpha1Client {
    connections: PromiseClient<typeof ConnectionService>;
    users: PromiseClient<typeof UserAccountService>;
    jobs: PromiseClient<typeof JobService>;
    transformers: PromiseClient<typeof TransformersService>;
    apikeys: PromiseClient<typeof ApiKeyService>;
}
/**
 * Function that returns the access token either as a string or a string promise
 */
type GetAccessTokenFn = () => string | Promise<string>;
interface ClientConfig {
    /**
     * Return the access token to be used for authenticating against Neosync API
     * This will either be a JWT, or an API Key
     * It will be used to construct the Authorization Header in the format: Authorization: Bearer <access token>
     */
    getAccessToken?: GetAccessTokenFn;
    /**
     * Return the connect transport for the appropriate environment (connect, grpc, web)
     * @param interceptors - A list of interceptors that have been pre-compuled. If `getAccessToken` is provided, this will include the auth interceptor
     */
    getTransport(interceptors: Interceptor[]): Transport;
}
/**
 * Returns the latest version of the Neosync Client
 */
declare function getNeosyncClient(config: ClientConfig): NeosyncClient;
/**
 * Returns the latest version of the Neosync Client
 */
declare function getNeosyncClient(config: ClientConfig, version: "latest"): NeosyncClient;
/**
 * Returns the v1alpha1 version of the Neosync Client
 */
declare function getNeosyncClient(config: ClientConfig, version: "v1alpha1"): NeosyncV1alpha1Client;
/**
 * Returns the v1alpha1 version of the Neosync client
 * @returns
 */
declare function getNeosyncV1alpha1Client(config: ClientConfig): NeosyncV1alpha1Client;

export { AcceptTeamAccountInviteRequest, AcceptTeamAccountInviteResponse, AccountApiKey, AccountInvite, AccountTemporalConfig, AccountUser, ActivityFailure, ActivityStatus, ApiKeyService, AwsS3ConnectionConfig, AwsS3Credentials, AwsS3DestinationConnectionOptions, AwsS3SourceConnectionOptions, CancelJobRunRequest, CancelJobRunResponse, CheckConnectionConfigRequest, CheckConnectionConfigResponse, CheckSqlQueryRequest, CheckSqlQueryResponse, type ClientConfig, type ClientVersion, Connection, ConnectionConfig, ConnectionService, ConvertPersonalToTeamAccountRequest, ConvertPersonalToTeamAccountResponse, CreateAccountApiKeyRequest, CreateAccountApiKeyResponse, CreateConnectionRequest, CreateConnectionResponse, CreateJobDestination, CreateJobDestinationConnectionsRequest, CreateJobDestinationConnectionsResponse, CreateJobRequest, CreateJobResponse, CreateJobRunRequest, CreateJobRunResponse, CreateTeamAccountRequest, CreateTeamAccountResponse, CreateUserDefinedTransformerRequest, CreateUserDefinedTransformerResponse, DatabaseColumn, DeleteAccountApiKeyRequest, DeleteAccountApiKeyResponse, DeleteConnectionRequest, DeleteConnectionResponse, DeleteJobDestinationConnectionRequest, DeleteJobDestinationConnectionResponse, DeleteJobRequest, DeleteJobResponse, DeleteJobRunRequest, DeleteJobRunResponse, DeleteUserDefinedTransformerRequest, DeleteUserDefinedTransformerResponse, ForeignConstraintTables, GenerateBool, GenerateCardNumber, GenerateCity, GenerateDefault, GenerateE164Number, GenerateEmail, GenerateFirstName, GenerateFloat, GenerateFullAddress, GenerateFullName, GenerateGender, GenerateInt, GenerateInt64Phone, GenerateLastName, GenerateRealisticEmail, GenerateSSN, GenerateSha256Hash, GenerateSourceOptions, GenerateSourceSchemaOption, GenerateSourceTableOption, GenerateState, GenerateStreetAddress, GenerateString, GenerateStringPhone, GenerateUnixTimestamp, GenerateUsername, GenerateUtcTimestamp, GenerateUuid, GenerateZipcode, type GetAccessTokenFn, GetAccountApiKeyRequest, GetAccountApiKeyResponse, GetAccountApiKeysRequest, GetAccountApiKeysResponse, GetAccountTemporalConfigRequest, GetAccountTemporalConfigResponse, GetConnectionDataStreamRequest, GetConnectionDataStreamResponse, GetConnectionForeignConstraintsRequest, GetConnectionForeignConstraintsResponse, GetConnectionRequest, GetConnectionResponse, GetConnectionSchemaRequest, GetConnectionSchemaResponse, GetConnectionsRequest, GetConnectionsResponse, GetJobNextRunsRequest, GetJobNextRunsResponse, GetJobRecentRunsRequest, GetJobRecentRunsResponse, GetJobRequest, GetJobResponse, GetJobRunEventsRequest, GetJobRunEventsResponse, GetJobRunRequest, GetJobRunResponse, GetJobRunsRequest, GetJobRunsResponse, GetJobStatusRequest, GetJobStatusResponse, GetJobStatusesRequest, GetJobStatusesResponse, GetJobsRequest, GetJobsResponse, GetSystemTransformersRequest, GetSystemTransformersResponse, GetTeamAccountInvitesRequest, GetTeamAccountInvitesResponse, GetTeamAccountMembersRequest, GetTeamAccountMembersResponse, GetUserAccountsRequest, GetUserAccountsResponse, GetUserDefinedTransformerByIdRequest, GetUserDefinedTransformerByIdResponse, GetUserDefinedTransformersRequest, GetUserDefinedTransformersResponse, GetUserRequest, GetUserResponse, InviteUserToTeamAccountRequest, InviteUserToTeamAccountResponse, IsConnectionNameAvailableRequest, IsConnectionNameAvailableResponse, IsJobNameAvailableRequest, IsJobNameAvailableResponse, IsTransformerNameAvailableRequest, IsTransformerNameAvailableResponse, IsUserInAccountRequest, IsUserInAccountResponse, Job, JobDestination, JobDestinationOptions, JobMapping, JobMappingTransformer, JobNextRuns, JobRecentRun, JobRun, JobRunEvent, JobRunEventMetadata, JobRunEventTask, JobRunEventTaskError, JobRunStatus, JobRunSyncMetadata, JobService, JobSource, JobSourceOptions, JobSourceSqlSubetSchemas, JobStatus, JobStatusRecord, MysqlConnection, MysqlConnectionConfig, MysqlDestinationConnectionOptions, MysqlSourceConnectionOptions, MysqlSourceSchemaOption, MysqlSourceSchemaSubset, MysqlSourceTableOption, MysqlTruncateTableConfig, type NeosyncClient, type NeosyncV1alpha1Client, Null, Passthrough, PauseJobRequest, PauseJobResponse, PendingActivity, PostgresConnection, PostgresConnectionConfig, PostgresDestinationConnectionOptions, PostgresSourceConnectionOptions, PostgresSourceSchemaOption, PostgresSourceSchemaSubset, PostgresSourceTableOption, PostgresTruncateTableConfig, RegenerateAccountApiKeyRequest, RegenerateAccountApiKeyResponse, RemoveTeamAccountInviteRequest, RemoveTeamAccountInviteResponse, RemoveTeamAccountMemberRequest, RemoveTeamAccountMemberResponse, SetAccountTemporalConfigRequest, SetAccountTemporalConfigResponse, SetJobSourceSqlConnectionSubsetsRequest, SetJobSourceSqlConnectionSubsetsResponse, SetPersonalAccountRequest, SetPersonalAccountResponse, SetUserRequest, SetUserResponse, SystemTransformer, TransformE164Phone, TransformEmail, TransformFirstName, TransformFloat, TransformFullName, TransformInt, TransformIntPhone, TransformLastName, TransformPhone, TransformString, TransformerConfig, TransformersService, UpdateConnectionRequest, UpdateConnectionResponse, UpdateJobDestinationConnectionRequest, UpdateJobDestinationConnectionResponse, UpdateJobScheduleRequest, UpdateJobScheduleResponse, UpdateJobSourceConnectionRequest, UpdateJobSourceConnectionResponse, UpdateUserDefinedTransformerRequest, UpdateUserDefinedTransformerResponse, UserAccount, UserAccountService, UserAccountType, UserDefinedTransformer, UserDefinedTransformerConfig, getNeosyncClient, getNeosyncV1alpha1Client };
