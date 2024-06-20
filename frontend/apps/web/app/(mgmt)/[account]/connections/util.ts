import {
  ClientTlsFormValues,
  MongoDbFormValues,
  MysqlFormValues,
  PostgresFormValues,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  CheckConnectionConfigResponse,
  ClientTlsConfig,
  ConnectionConfig,
  CreateConnectionRequest,
  CreateConnectionResponse,
  IsConnectionNameAvailableResponse,
  MongoConnectionConfig,
  MysqlConnection,
  MysqlConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
  SqlConnectionOptions,
  UpdateConnectionRequest,
  UpdateConnectionResponse,
} from '@neosync/sdk';

export type ConnectionType =
  | 'postgres'
  | 'mysql'
  | 'aws-s3'
  | 'openai'
  | 'mongodb';

export const DESTINATION_ONLY_CONNECTION_TYPES = new Set<ConnectionType>([
  'aws-s3',
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
    default:
      return null;
  }
}

export async function isConnectionNameAvailable(
  name: string,
  accountId: string
): Promise<IsConnectionNameAvailableResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/is-connection-name-available?connectionName=${name}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return IsConnectionNameAvailableResponse.fromJson(await res.json());
}

export async function createMysqlConnection(
  values: MysqlFormValues,
  accountId: string
): Promise<CreateConnectionResponse> {
  return createConnection(
    new CreateConnectionRequest({
      name: values.connectionName,
      accountId: accountId,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mysqlConfig',
          value: buildMysqlConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function updateMysqlConnection(
  values: MysqlFormValues,
  accountId: string,
  resourceId: string
): Promise<UpdateConnectionResponse> {
  return updateConnection(
    new UpdateConnectionRequest({
      id: resourceId,
      name: values.connectionName,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mysqlConfig',
          value: buildMysqlConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function checkMysqlConnection(
  values: MysqlFormValues,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mysqlConfig',
          value: buildMysqlConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
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

export async function createPostgresConnection(
  values: PostgresFormValues,
  accountId: string
): Promise<CreateConnectionResponse> {
  return createConnection(
    new CreateConnectionRequest({
      name: values.connectionName,
      accountId: accountId,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'pgConfig',
          value: buildPostgresConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function updatePostgresConnection(
  values: PostgresFormValues,
  accountId: string,
  resourceId: string
): Promise<UpdateConnectionResponse> {
  return updateConnection(
    new UpdateConnectionRequest({
      id: resourceId,
      name: values.connectionName,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'pgConfig',
          value: buildPostgresConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function checkPostgresConnection(
  values: PostgresFormValues,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'pgConfig',
          value: buildPostgresConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
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

export async function createMongoConnection(
  values: MongoDbFormValues,
  accountId: string
): Promise<CreateConnectionResponse> {
  return createConnection(
    new CreateConnectionRequest({
      name: values.connectionName,
      accountId: accountId,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mongoConfig',
          value: buildMongoConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function updateMongoConnection(
  values: MongoDbFormValues,
  accountId: string,
  resourceId: string
): Promise<UpdateConnectionResponse> {
  return updateConnection(
    new UpdateConnectionRequest({
      id: resourceId,
      name: values.connectionName,
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mongoConfig',
          value: buildMongoConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

export async function checkMongoConnection(
  values: MongoDbFormValues,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mongoConfig',
          value: buildMongoConnectionConfig(values),
        },
      }),
    }),
    accountId
  );
}

function buildMongoConnectionConfig(
  values: MongoDbFormValues
): MongoConnectionConfig {
  const mongoconfig = new MongoConnectionConfig({
    connectionConfig: {
      case: 'url',
      value: values.url,
    },
    tunnel: getTunnelConfig(values.tunnel),
  });

  return mongoconfig;
}

async function checkConnection(
  input: CheckConnectionConfigRequest,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  const res = await fetch(buildCheckConnectionKey(accountId), {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CheckConnectionConfigResponse.fromJson(await res.json());
}

export function buildCheckConnectionKey(accountId: string): string {
  return `/api/accounts/${accountId}/connections/check`;
}

async function createConnection(
  input: CreateConnectionRequest,
  accountId: string
): Promise<CreateConnectionResponse> {
  const res = await fetch(`/api/accounts/${accountId}/connections`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(input),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateConnectionResponse.fromJson(await res.json());
}

async function updateConnection(
  input: UpdateConnectionRequest,
  accountId: string
): Promise<UpdateConnectionResponse> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${input.id}`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(input),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateConnectionResponse.fromJson(await res.json());
}

export async function removeConnection(
  accountId: string,
  connectionId: string
): Promise<void> {
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
