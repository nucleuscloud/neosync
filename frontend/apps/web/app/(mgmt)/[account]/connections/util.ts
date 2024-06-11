import { MongoDbFormValues } from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  CheckConnectionConfigResponse,
  ConnectionConfig,
  MongoConnectionConfig,
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
