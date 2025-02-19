import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import {
  AwsCredentialsFormValues,
  AwsFormValues,
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
import { create } from '@bufbuild/protobuf';
import {
  AwsS3ConnectionConfigSchema,
  AwsS3Credentials,
  AwsS3CredentialsSchema,
  ClientTlsConfig,
  ClientTlsConfigSchema,
  Connection,
  ConnectionConfig,
  ConnectionConfigSchema,
  DynamoDBConnectionConfigSchema,
  GcpCloudStorageConnectionConfig,
  GcpCloudStorageConnectionConfigSchema,
  MongoConnectionConfig,
  MongoConnectionConfigSchema,
  MssqlConnectionConfig,
  MssqlConnectionConfigSchema,
  MysqlConnectionConfig,
  MysqlConnectionConfigSchema,
  MysqlConnectionSchema,
  OpenAiConnectionConfigSchema,
  PostgresConnectionConfig,
  PostgresConnectionConfigSchema,
  PostgresConnectionSchema,
  SqlConnectionOptions,
  SqlConnectionOptionsSchema,
  SSHAuthenticationSchema,
  SSHPassphraseSchema,
  SSHPrivateKeySchema,
  SSHTunnel,
  SSHTunnelSchema,
} from '@neosync/sdk';

export interface ConnectionMeta {
  name: string;
  description: string;
  urlSlug: string;
  connectionType: ConnectionConfigCase;
  connectionTypeVariant?: ConnectionTypeVariant;
  isExperimental?: boolean;
  isLicenseOnly?: boolean;
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
    isLicenseOnly: true,
  },
  {
    urlSlug: 'gcp-cloud-storage',
    name: 'GCP Cloud Storage',
    description:
      'GCP Cloud Storage is an object storage service used to store and retrieve any data.',
    connectionType: 'gcpCloudstorageConfig',
    isLicenseOnly: true,
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
  },
  {
    urlSlug: 'dynamodb',
    name: 'DynamoDB',
    description:
      'Amazon DynamoDB is a fully managed proprietary NoSQL database offered by Amazon.com as part of the Amazon Web Services portfolio',
    connectionType: 'dynamodbConfig',
  },
  {
    urlSlug: 'mssql',
    name: 'Microsoft SQL Server',
    description:
      'Microsoft SQL Server is a proprietary relational database management system developed by Microsoft.',
    connectionType: 'mssqlConfig',
  },
];

export function useGetConnectionsMetadata(
  allowedConnectionTypes: Set<string>
): ConnectionMeta[] {
  const { data: systemAppConfigData } = useGetSystemAppConfig();

  return getConnectionsMetadata(
    allowedConnectionTypes,
    systemAppConfigData?.isGcpCloudStorageConnectionsEnabled ?? false
  );
}

function getConnectionsMetadata(
  connectionTypes: Set<string>,
  isGcpCloudStorageConnectionsEnabled: boolean
): ConnectionMeta[] {
  let connections = CONNECTIONS_METADATA;
  if (!isGcpCloudStorageConnectionsEnabled) {
    connections = connections.filter(
      (c) => c.connectionType !== 'gcpCloudstorageConfig'
    );
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
  connectionConfig: ConnectionConfig
): ConnectionConfigCase | null {
  return connectionConfig.config.case ?? null;
}

export function getConnectionUrlSlugName(
  connectionConfig: ConnectionConfig
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
export function getCategory(cc?: ConnectionConfig): string {
  if (!cc || !cc.config.case) {
    return '-';
  }
  const connType = getConnectionType(cc);
  return connType ? CONNECTION_CATEGORY_MAP[connType] : '-';
}

export function buildConnectionConfigDynamoDB(
  values: DynamoDbFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'dynamodbConfig',
      value: create(DynamoDBConnectionConfigSchema, {
        endpoint: values.advanced?.endpoint,
        region: values.advanced?.region,
        credentials: values.credentials
          ? buildAwsCredentials(values.credentials)
          : undefined,
      }),
    },
  });
}

export function buildConnectionConfigAwsS3(
  values: AwsFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'awsS3Config',
      value: create(AwsS3ConnectionConfigSchema, {
        bucket: values.s3.bucket,
        pathPrefix: values.s3.pathPrefix,
        region: values.advanced?.region,
        endpoint: values.advanced?.endpoint,
        credentials: values.credentials
          ? buildAwsCredentials(values.credentials)
          : undefined,
      }),
    },
  });
}

function buildAwsCredentials(
  values: AwsCredentialsFormValues
): AwsS3Credentials {
  return create(AwsS3CredentialsSchema, {
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
  return create(ConnectionConfigSchema, {
    config: {
      case: 'gcpCloudstorageConfig',
      value: buildGcpCloudStorageConnectionConfig(values),
    },
  });
}

export function buildConnectionConfigPostgres(
  values: PostgresFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'pgConfig',
      value: buildPostgresConnectionConfig(values),
    },
  });
}

export function buildConnectionConfigOpenAi(
  values: OpenAiFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'openaiConfig',
      value: create(OpenAiConnectionConfigSchema, {
        apiUrl: values.sdk.url,
        apiKey: values.sdk.apiKey,
      }),
    },
  });
}

export function buildConnectionConfigMysql(
  values: MysqlFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'mysqlConfig',
      value: buildMysqlConnectionConfig(values),
    },
  });
}

export function buildConnectionConfigMssql(
  values: MssqlFormValues
): ConnectionConfig {
  return create(ConnectionConfigSchema, {
    config: {
      case: 'mssqlConfig',
      value: buildMssqlConnectionConfig(values),
    },
  });
}

function buildMssqlConnectionConfig(
  values: MssqlFormValues
): MssqlConnectionConfig {
  const mssqlconfig = create(MssqlConnectionConfigSchema, {
    connectionOptions: create(SqlConnectionOptionsSchema, {
      ...values.options,
    }),
    tunnel: getTunnelConfig(values.tunnel),
    clientTls: getClientTlsConfig(values.clientTls),
  });

  if (values.activeTab === 'url' && values.url) {
    mssqlconfig.connectionConfig = {
      case: 'url',
      value: values.url,
    };
  } else if (values.activeTab === 'url-env' && values.envVar) {
    mssqlconfig.connectionConfig = {
      case: 'urlFromEnv',
      value: values.envVar,
    };
  }
  return mssqlconfig;
}

function buildGcpCloudStorageConnectionConfig(
  values: GcpCloudStorageFormValues
): GcpCloudStorageConnectionConfig {
  return create(GcpCloudStorageConnectionConfigSchema, {
    bucket: values.gcp.bucket,
    pathPrefix: values.gcp.pathPrefix,
  });
}

function buildMysqlConnectionConfig(
  values: MysqlFormValues
): MysqlConnectionConfig {
  const mysqlconfig = create(MysqlConnectionConfigSchema, {
    connectionOptions: create(SqlConnectionOptionsSchema, {
      ...values.options,
    }),
    tunnel: getTunnelConfig(values.tunnel),
    clientTls: getClientTlsConfig(values.clientTls),
  });

  if (values.activeTab === 'url' && values.url) {
    mysqlconfig.connectionConfig = {
      case: 'url',
      value: values.url,
    };
  } else if (values.activeTab === 'url-env' && values.envVar) {
    mysqlconfig.connectionConfig = {
      case: 'urlFromEnv',
      value: values.envVar,
    };
  } else if (values.activeTab === 'host' && values.db) {
    mysqlconfig.connectionConfig = {
      case: 'connection',
      value: create(MysqlConnectionSchema, {
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
  return create(SqlConnectionOptionsSchema, {
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
  const pgconfig = create(PostgresConnectionConfigSchema, {
    clientTls: getClientTlsConfig(values.clientTls),
    tunnel: getTunnelConfig(values.tunnel),
    connectionOptions: getSqlConnectionOptions(values.options),
  });

  if (values.activeTab === 'url' && values.url) {
    pgconfig.connectionConfig = {
      case: 'url',
      value: values.url,
    };
  } else if (values.activeTab === 'url-env' && values.envVar) {
    pgconfig.connectionConfig = {
      case: 'urlFromEnv',
      value: values.envVar,
    };
  } else if (values.activeTab === 'host' && values.db) {
    pgconfig.connectionConfig = {
      case: 'connection',
      value: create(PostgresConnectionSchema, {
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
  return create(ClientTlsConfigSchema, {
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
  const tunnel = create(SSHTunnelSchema, {
    host: values.host,
    port: values.port,
    user: values.user,
    knownHostPublicKey: values.knownHostPublicKey
      ? values.knownHostPublicKey
      : undefined,
  });

  if (values.privateKey) {
    tunnel.authentication = create(SSHAuthenticationSchema, {
      authConfig: {
        case: 'privateKey',
        value: create(SSHPrivateKeySchema, {
          value: values.privateKey,
          passphrase: values.passphrase,
        }),
      },
    });
  } else if (values.passphrase) {
    tunnel.authentication = create(SSHAuthenticationSchema, {
      authConfig: {
        case: 'passphrase',
        value: create(SSHPassphraseSchema, {
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
  return create(ConnectionConfigSchema, {
    config: {
      case: 'mongoConfig',
      value: buildMongoConnectionConfig(values),
    },
  });
}

function buildMongoConnectionConfig(
  values: MongoDbFormValues
): MongoConnectionConfig {
  const mongoconfig = create(MongoConnectionConfigSchema, {
    connectionConfig: {
      case: 'url',
      value: values.url,
    },

    clientTls: getClientTlsConfig(values.clientTls),
  });

  return mongoconfig;
}
