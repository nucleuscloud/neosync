import {
  AwsCredentialsFormValues,
  AWSFormValues,
  ClientTlsFormValues,
  DynamoDbFormValues,
  GcpCloudStorageFormValues,
  MongoDbFormValues,
  MysqlFormValues,
  OpenAiFormValues,
  PostgresFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  AwsS3ConnectionConfig,
  AwsS3Credentials,
  ClientTlsConfig,
  Connection,
  ConnectionConfig,
  DynamoDBConnectionConfig,
  GcpCloudStorageConnectionConfig,
  MongoConnectionConfig,
  MysqlConnection,
  MysqlConnectionConfig,
  OpenAiConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
  SqlConnectionOptions,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
} from '@neosync/sdk';

// Helper function to extract the 'case' property from a config type
type ExtractCase<T> = T extends { case: infer U } ? U : never;

// Extraction type that pulls out all of the connection config cases
type ConnectionConfigCase = NonNullable<
  ExtractCase<ConnectionConfig['config']>
>;

// Key is Source config, set of values are allowed destinations
const ALLOWED_SOURCE_CONNECTION_PAIRS: Record<
  ConnectionConfigCase,
  Set<ConnectionConfigCase>
> = {
  awsS3Config: new Set<ConnectionConfigCase>(),
  dynamodbConfig: new Set<ConnectionConfigCase>(['dynamodbConfig']),
  gcpCloudstorageConfig: new Set<ConnectionConfigCase>(),
  localDirConfig: new Set<ConnectionConfigCase>(),
  mongoConfig: new Set<ConnectionConfigCase>(['mongoConfig']),
  mysqlConfig: new Set<ConnectionConfigCase>([
    'mysqlConfig',
    'awsS3Config',
    'gcpCloudstorageConfig',
  ]),
  openaiConfig: new Set<ConnectionConfigCase>([
    'pgConfig',
    'mysqlConfig',
    'awsS3Config',
    'gcpCloudstorageConfig',
  ]),
  pgConfig: new Set<ConnectionConfigCase>([
    'pgConfig',
    'awsS3Config',
    'gcpCloudstorageConfig',
  ]),
};

// Given two connections, determines if they are a valid pair.
export function isValidConnectionPair(
  sourceConn: Connection,
  destConn: Connection
): boolean {
  if (
    !sourceConn.connectionConfig?.config.case ||
    !destConn.connectionConfig?.config.case
  ) {
    return false;
  }

  const allowsPairs =
    ALLOWED_SOURCE_CONNECTION_PAIRS[sourceConn.connectionConfig.config.case];

  return (
    allowsPairs.size > 0 &&
    allowsPairs.has(destConn.connectionConfig.config.case)
  );
}

export type ConnectionType =
  | 'postgres'
  | 'mysql'
  | 'aws-s3'
  | 'openai'
  | 'mongodb'
  | 'gcp-cloud-storage'
  | 'dynamodb';

// Variant of a connection type.
export type ConnectionTypeVariant = 'neon' | 'supabase';

export const DESTINATION_ONLY_CONNECTION_TYPES = new Set<ConnectionType>([
  'aws-s3',
  'gcp-cloud-storage',
]);

export function getConnectionType(
  connectionConfig: ConnectionConfig
): ConnectionType | null {
  switch (connectionConfig.config.case) {
    case 'pgConfig':
      return 'postgres';
    case 'mysqlConfig':
      return 'mysql';
    case 'awsS3Config':
      return 'aws-s3';
    case 'openaiConfig':
      return 'openai';
    case 'mongoConfig':
      return 'mongodb';
    case 'gcpCloudstorageConfig':
      return 'gcp-cloud-storage';
    case 'dynamodbConfig':
      return 'dynamodb';
    default:
      return null;
  }
}

// Used for the connections data table
export function getCategory(cc?: PlainMessage<ConnectionConfig>): string {
  if (!cc) {
    return '-';
  }
  switch (cc.config.case) {
    case 'pgConfig':
      return 'Postgres';
    case 'mysqlConfig':
      return 'MySQL';
    case 'awsS3Config':
      return 'AWS S3';
    case 'openaiConfig':
      return 'OpenAI';
    case 'localDirConfig':
      return 'Local Dir';
    case 'mongoConfig':
      return 'MongoDB';
    case 'gcpCloudstorageConfig':
      return 'GCP Cloud Storage';
    case 'dynamodbConfig':
      return 'DynamoDB';
    default:
      return '-';
  }
}

export function buildConnectionConfigDynamoDB(
  values: DynamoDbFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'dynamodbConfig',
      value: new DynamoDBConnectionConfig({
        endpoint: values.db.endpoint,
        region: values.db.region,
        credentials: values.db.credentials
          ? buildAwsCredentials(values.db.credentials)
          : undefined,
      }),
    },
  });
}

export function buildConnectionConfigAwsS3(
  values: AWSFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'awsS3Config',
      value: new AwsS3ConnectionConfig({
        bucket: values.s3.bucket,
        pathPrefix: values.s3.pathPrefix,
        region: values.s3.region,
        endpoint: values.s3.endpoint,
        credentials: values.s3.credentials
          ? buildAwsCredentials(values.s3.credentials)
          : undefined,
      }),
    },
  });
}

function buildAwsCredentials(
  values: AwsCredentialsFormValues
): AwsS3Credentials {
  return new AwsS3Credentials({
    profile: values.profile,
    accessKeyId: values.accessKeyId,
    secretAccessKey: values.secretAccessKey,
    fromEc2Role: values.fromEc2Role,
    roleArn: values.roleArn,
    roleExternalId: values.roleExternalId,
    sessionToken: values.sessionToken,
  });
}

export function buildConnectionConfigGcpCloudStorage(
  values: GcpCloudStorageFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'gcpCloudstorageConfig',
      value: buildGcpCloudStorageConnectionConfig(values),
    },
  });
}

export function buildConnectionConfigPostgres(
  values: PostgresFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'pgConfig',
      value: buildPostgresConnectionConfig(values),
    },
  });
}

export function buildConnectionConfigOpenAi(
  values: OpenAiFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'openaiConfig',
      value: new OpenAiConnectionConfig({
        apiUrl: values.sdk.url,
        apiKey: values.sdk.apiKey,
      }),
    },
  });
}

export function buildConnectionConfigMysql(
  values: MysqlFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'mysqlConfig',
      value: buildMysqlConnectionConfig(values),
    },
  });
}

function buildGcpCloudStorageConnectionConfig(
  values: GcpCloudStorageFormValues
): GcpCloudStorageConnectionConfig {
  return new GcpCloudStorageConnectionConfig({
    bucket: values.gcp.bucket,
    pathPrefix: values.gcp.pathPrefix,
  });
}

function buildMysqlConnectionConfig(
  values: MysqlFormValues
): MysqlConnectionConfig {
  const mysqlconfig = new MysqlConnectionConfig({
    connectionOptions: new SqlConnectionOptions({
      ...values.options,
    }),
    tunnel: getTunnelConfig(values.tunnel),
  });

  if (values.url) {
    mysqlconfig.connectionConfig = {
      case: 'url',
      value: values.url,
    };
  } else {
    mysqlconfig.connectionConfig = {
      case: 'connection',
      value: new MysqlConnection({
        host: values.db.host,
        name: values.db.name,
        pass: values.db.pass,
        port: values.db.port,
        protocol: values.db.protocol,
        user: values.db.user,
      }),
    };
  }
  return mysqlconfig;
}

function buildPostgresConnectionConfig(
  values: PostgresFormValues
): PostgresConnectionConfig {
  const pgconfig = new PostgresConnectionConfig({
    clientTls: getClientTlsConfig(values.clientTls),
    tunnel: getTunnelConfig(values.tunnel),
    connectionOptions: new SqlConnectionOptions({
      ...values.options,
    }),
  });

  if (values.url) {
    pgconfig.connectionConfig = {
      case: 'url',
      value: values.url,
    };
  } else {
    pgconfig.connectionConfig = {
      case: 'connection',
      value: new PostgresConnection({
        host: values.db.host,
        port: values.db.port,
        name: values.db.name,
        pass: values.db.pass,
        sslMode: values.db.sslMode,
        user: values.db.user,
      }),
    };
  }
  return pgconfig;
}

function getClientTlsConfig(
  values?: ClientTlsFormValues
): ClientTlsConfig | undefined {
  if (
    !values ||
    (!values.rootCert && !values.clientKey && !values.clientCert)
  ) {
    return undefined;
  }
  return new ClientTlsConfig({
    rootCert: values.rootCert ? values.rootCert : undefined,
    clientKey: values.clientKey ? values.clientKey : undefined,
    clientCert: values.clientCert ? values.clientCert : undefined,
  });
}

function getTunnelConfig(values?: SshTunnelFormValues): SSHTunnel | undefined {
  if (!values || !values.host) {
    return undefined;
  }
  const tunnel = new SSHTunnel({
    host: values.host,
    port: values.port,
    user: values.user,
    knownHostPublicKey: values.knownHostPublicKey
      ? values.knownHostPublicKey
      : undefined,
  });

  if (values.privateKey) {
    tunnel.authentication = new SSHAuthentication({
      authConfig: {
        case: 'privateKey',
        value: new SSHPrivateKey({
          value: values.privateKey,
          passphrase: values.passphrase,
        }),
      },
    });
  } else if (values.passphrase) {
    tunnel.authentication = new SSHAuthentication({
      authConfig: {
        case: 'passphrase',
        value: new SSHPassphrase({
          value: values.passphrase,
        }),
      },
    });
  }
  return tunnel;
}

export function buildConnectionConfigMongo(
  values: MongoDbFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'mongoConfig',
      value: buildMongoConnectionConfig(values),
    },
  });
}

function buildMongoConnectionConfig(
  values: MongoDbFormValues
): MongoConnectionConfig {
  const mongoconfig = new MongoConnectionConfig({
    connectionConfig: {
      case: 'url',
      value: values.url,
    },

    clientTls: getClientTlsConfig(values.clientTls),
  });

  return mongoconfig;
}
