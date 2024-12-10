import {
  AwsCredentialsFormValues,
  AWSFormValues,
  ClientTlsFormValues,
  DynamoDbFormValues,
  GcpCloudStorageFormValues,
  MongoDbFormValues,
  MssqlFormValues,
  MysqlFormValues,
  OpenAiFormValues,
  PostgresFormValues,
  SqlOptionsFormValues,
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
  MssqlConnectionConfig,
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

export interface ConnectionMeta {
  name: string;
  description: string;
  urlSlug: string;
  connectionType: ConnectionConfigCase;
  connectionTypeVariant?: ConnectionTypeVariant;
  isExperimental?: boolean;
}

const CONNECTIONS_METADATA: ConnectionMeta[] = [
  {
    urlSlug: 'postgres',
    name: 'Postgres',
    description:
      'PostgreSQL is a free and open-source relational database manageent system emphasizing extensibility and SQL compliance.',
    connectionType: 'pgConfig',
  },
  {
    urlSlug: 'mysql',
    name: 'MySQL',
    description:
      'MySQL is an open-source relational database management system.',
    connectionType: 'mysqlConfig',
  },
  {
    urlSlug: 'aws-s3',
    name: 'AWS S3',
    description:
      'Amazon Simple Storage Service (Amazon S3) is an object storage service used to store and retrieve any data.',
    connectionType: 'awsS3Config',
  },
  {
    urlSlug: 'gcp-cloud-storage',
    name: 'GCP Cloud Storage',
    description:
      'GCP Cloud Storage is an object storage service used to store and retrieve any data.',
    connectionType: 'gcpCloudstorageConfig',
  },
  {
    urlSlug: 'neon',
    name: 'Neon',
    description:
      'Neon is a serverless Postgres database that separates storage and compute to offer autoscaling, branching and bottomless storage.',
    connectionType: 'pgConfig',
    connectionTypeVariant: 'neon',
  },
  {
    urlSlug: 'supabase',
    name: 'Supabase',
    description:
      'Supabase is an open source Firebase alternative that ships with Authentication, Instant APIs, Edge functions and more.',
    connectionType: 'pgConfig',
    connectionTypeVariant: 'supabase',
  },
  {
    urlSlug: 'openai',
    name: 'OpenAI',
    description:
      'OpenAI (or equivalent interface) Chat API for generating synthetic data and inserting it directly into a destination datasource.',
    connectionType: 'openaiConfig',
  },
  {
    urlSlug: 'mongodb',
    name: 'MongoDB',
    description:
      'MongoDB is a source-available, cross-platform, document-oriented database program.',
    connectionType: 'mongoConfig',
    isExperimental: true,
  },
  {
    urlSlug: 'dynamodb',
    name: 'DynamoDB',
    description:
      'Amazon DynamoDB is a fully managed proprietary NoSQL database offered by Amazon.com as part of the Amazon Web Services portfolio',
    connectionType: 'dynamodbConfig',
    isExperimental: true,
  },
  {
    urlSlug: 'mssql',
    name: 'Microsoft SQL Server',
    description:
      'Microsoft SQL Server is a proprietary relational database management system developed by Microsoft.',
    connectionType: 'mssqlConfig',
    isExperimental: true,
  },
];

export function getConnectionsMetadata(
  connectionTypes: Set<string>,
  isGcpCloudStorageConnectionsEnabled: boolean,
  isDynamoDbConnectionsEnabled: boolean,
  isMssqlConnectionsEnabled: boolean
): ConnectionMeta[] {
  let connections = CONNECTIONS_METADATA;
  if (!isGcpCloudStorageConnectionsEnabled) {
    connections = connections.filter(
      (c) => c.connectionType !== 'gcpCloudstorageConfig'
    );
  }
  if (!isDynamoDbConnectionsEnabled) {
    connections = connections.filter(
      (c) => c.connectionType !== 'dynamodbConfig'
    );
  }
  if (!isMssqlConnectionsEnabled) {
    connections = connections.filter((c) => c.connectionType !== 'mssqlConfig');
  }
  if (connectionTypes.size > 0) {
    connections = connections.filter((c) =>
      connectionTypes.has(c.connectionType)
    );
  }
  return connections;
}

// Helper function to extract the 'case' property from a config type
type ExtractCase<T> = T extends { case: infer U } ? U : never;

// Extraction type that pulls out all of the connection config cases
export type ConnectionConfigCase = NonNullable<
  ExtractCase<ConnectionConfig['config']>
>;

// Key is Source config, set of values are allowed destinations
const ALLOWED_SYNC_SOURCE_CONNECTION_PAIRS: Record<
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
  openaiConfig: new Set<ConnectionConfigCase>(),
  pgConfig: new Set<ConnectionConfigCase>([
    'pgConfig',
    'awsS3Config',
    'gcpCloudstorageConfig',
  ]),
  mssqlConfig: new Set<ConnectionConfigCase>(['mssqlConfig']),
};
function reverseConfigCaseMap(
  originalMap: Record<ConnectionConfigCase, Set<ConnectionConfigCase>>
): Record<ConnectionConfigCase, Set<ConnectionConfigCase>> {
  const reversedMap: Record<ConnectionConfigCase, Set<ConnectionConfigCase>> = {
    awsS3Config: new Set<ConnectionConfigCase>(),
    dynamodbConfig: new Set<ConnectionConfigCase>(),
    gcpCloudstorageConfig: new Set<ConnectionConfigCase>(),
    localDirConfig: new Set<ConnectionConfigCase>(),
    mongoConfig: new Set<ConnectionConfigCase>(),
    mysqlConfig: new Set<ConnectionConfigCase>(),
    openaiConfig: new Set<ConnectionConfigCase>(),
    pgConfig: new Set<ConnectionConfigCase>(),
    mssqlConfig: new Set<ConnectionConfigCase>(),
  };

  Object.entries(originalMap).forEach(([source, destinations]) => {
    destinations.forEach((destination) => {
      if (!reversedMap[destination]) {
        reversedMap[destination] = new Set<ConnectionConfigCase>();
      }
      reversedMap[destination].add(source as ConnectionConfigCase);
    });
  });

  return reversedMap;
}
const ALLOWED_SYNC_DEST_CONNECTION_PAIRS = reverseConfigCaseMap(
  ALLOWED_SYNC_SOURCE_CONNECTION_PAIRS
);

export function getAllowedSyncDestinationTypes(
  sourceType?: ConnectionConfigCase
): Set<ConnectionConfigCase> {
  if (sourceType) {
    return new Set(ALLOWED_SYNC_SOURCE_CONNECTION_PAIRS[sourceType]);
  }

  return new Set(
    Object.values(ALLOWED_SYNC_SOURCE_CONNECTION_PAIRS).flatMap((set) =>
      Array.from(set)
    )
  );
}

export function getAllowedSyncSourceTypes(
  destTypes: ConnectionConfigCase[] = []
): Set<ConnectionConfigCase> {
  const filteredDests = destTypes.filter(
    (dt) => !DESTINATION_ONLY_CONNECTION_TYPES.has(dt)
  );
  return new Set(
    filteredDests
      .map((dest) => ALLOWED_SYNC_DEST_CONNECTION_PAIRS[dest])
      .flatMap((set) => Array.from(set))
  );
}

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
    ALLOWED_SYNC_SOURCE_CONNECTION_PAIRS[
      sourceConn.connectionConfig.config.case
    ];

  return (
    allowsPairs.size > 0 &&
    allowsPairs.has(destConn.connectionConfig.config.case)
  );
}

// Variant of a connection type.
export type ConnectionTypeVariant = 'neon' | 'supabase';

const DESTINATION_ONLY_CONNECTION_TYPES = new Set<ConnectionConfigCase>([
  'awsS3Config',
  'gcpCloudstorageConfig',
]);

export function getConnectionType(
  connectionConfig: PlainMessage<ConnectionConfig>
): ConnectionConfigCase | null {
  return connectionConfig.config.case ?? null;
}

export function getConnectionUrlSlugName(
  connectionConfig: PlainMessage<ConnectionConfig>
): string {
  const connType = getConnectionType(connectionConfig);
  if (!connType) {
    return '';
  }
  const metadata = CONNECTIONS_METADATA.find(
    (md) => md.connectionType === connType
  );
  return metadata ? metadata.urlSlug : '';
}

const CONNECTION_CATEGORY_MAP: Record<ConnectionConfigCase, string> = {
  awsS3Config: 'AWS S3',
  dynamodbConfig: 'DynamoDB',
  gcpCloudstorageConfig: 'GCP Cloud Storage',
  localDirConfig: 'Local Directory',
  mongoConfig: 'MongoDB',
  mysqlConfig: 'MySQL',
  openaiConfig: 'OpenAI',
  pgConfig: 'PostgreSQL',
  mssqlConfig: 'Microsoft SQL Server',
};

// Used for the connections data table
export function getCategory(cc?: PlainMessage<ConnectionConfig>): string {
  if (!cc || !cc.config.case) {
    return '-';
  }
  const connType = getConnectionType(cc);
  return connType ? CONNECTION_CATEGORY_MAP[connType] : '-';
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

export function buildConnectionConfigMssql(
  values: MssqlFormValues
): ConnectionConfig {
  return new ConnectionConfig({
    config: {
      case: 'mssqlConfig',
      value: buildMssqlConnectionConfig(values),
    },
  });
}

function buildMssqlConnectionConfig(
  values: MssqlFormValues
): MssqlConnectionConfig {
  return new MssqlConnectionConfig({
    connectionConfig: {
      case: 'url',
      value: values.db.url,
    },
    connectionOptions: new SqlConnectionOptions({
      ...values.options,
    }),
    tunnel: getTunnelConfig(values.tunnel),
    clientTls: getClientTlsConfig(values.clientTls),
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
    clientTls: getClientTlsConfig(values.clientTls),
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

function getSqlConnectionOptions(
  values: SqlOptionsFormValues
): SqlConnectionOptions {
  return new SqlConnectionOptions({
    maxConnectionLimit:
      values.maxConnectionLimit != null && values.maxConnectionLimit >= 0
        ? values.maxConnectionLimit
        : undefined,
    maxIdleConnections:
      values.maxIdleLimit != null && values.maxIdleLimit >= 0
        ? values.maxIdleLimit
        : undefined,
    maxIdleDuration: values.maxIdleDuration
      ? values.maxIdleDuration
      : undefined,
    maxOpenDuration: values.maxOpenDuration
      ? values.maxOpenDuration
      : undefined,
  });
}

function buildPostgresConnectionConfig(
  values: PostgresFormValues
): PostgresConnectionConfig {
  const pgconfig = new PostgresConnectionConfig({
    clientTls: getClientTlsConfig(values.clientTls),
    tunnel: getTunnelConfig(values.tunnel),
    connectionOptions: getSqlConnectionOptions(values.options),
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
    (!values.rootCert &&
      !values.clientKey &&
      !values.clientCert &&
      !values.serverName)
  ) {
    return undefined;
  }
  return new ClientTlsConfig({
    rootCert: values.rootCert ? values.rootCert : undefined,
    clientKey: values.clientKey ? values.clientKey : undefined,
    clientCert: values.clientCert ? values.clientCert : undefined,
    serverName: values.serverName ? values.serverName : undefined,
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
