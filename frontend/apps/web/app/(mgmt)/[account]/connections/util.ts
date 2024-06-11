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

export async function checkMysqlConnection(
  values: MysqlFormValues,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
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

  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mysqlConfig',
          value: mysqlconfig,
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

  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'pgConfig',
          value: pgconfig,
        },
      }),
    }),
    accountId
  );
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
  if (!values || (!values.host && !values.port && !values.user)) {
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

export async function checkMongoConnection(
  values: MongoDbFormValues,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  // let tunnel: SSHTunnel | undefined = undefined;
  // if (tunnelForm && tunnelForm.host && tunnelForm.port && tunnelForm.user) {
  //   tunnel = new SSHTunnel({
  //     host: tunnelForm.host,
  //     port: tunnelForm.port,
  //     user: tunnelForm.user,
  //     knownHostPublicKey: tunnelForm.knownHostPublicKey
  //       ? tunnelForm.knownHostPublicKey
  //       : undefined,
  //   });
  //   if (tunnelForm.privateKey) {
  //     tunnel.authentication = new SSHAuthentication({
  //       authConfig: {
  //         case: 'privateKey',
  //         value: new SSHPrivateKey({
  //           value: tunnelForm.privateKey,
  //           passphrase: tunnelForm.passphrase,
  //         }),
  //       },
  //     });
  //   } else if (tunnelForm.passphrase) {
  //     tunnel.authentication = new SSHAuthentication({
  //       authConfig: {
  //         case: 'passphrase',
  //         value: new SSHPassphrase({
  //           value: tunnelForm.passphrase,
  //         }),
  //       },
  //     });
  //   }
  // }
  return checkConnection(
    new CheckConnectionConfigRequest({
      connectionConfig: new ConnectionConfig({
        config: {
          case: 'mongoConfig',
          value: new MongoConnectionConfig({
            connectionConfig: {
              case: 'url',
              value: values.url,
            },
            tunnel: undefined,
          }),
        },
      }),
    }),
    accountId
  );
}

async function checkConnection(
  input: CheckConnectionConfigRequest,
  accountId: string
): Promise<CheckConnectionConfigResponse> {
  const res = await fetch(`/api/accounts/${accountId}/connections/check`, {
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
