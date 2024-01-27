import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  EXISTING_POSTGRES_CONNECTION,
  NEW_POSTGRES_CONNECTION,
} from '@/yup-validations/connections';
import {
  ConnectionConfig,
  CreateConnectionRequest,
  PostgresConnection,
  PostgresConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
  UpdateConnectionRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const jsonbody = await NEW_POSTGRES_CONNECTION.validate(await req.json());

    const pgconfig = new PostgresConnectionConfig({
      connectionConfig: {
        case: 'connection',
        value: new PostgresConnection({
          host: jsonbody.connection.host,
          name: jsonbody.connection.name,
          user: jsonbody.connection.user,
          pass: jsonbody.connection.pass,
          port: jsonbody.connection.port,
          sslMode: jsonbody.connection.sslMode,
        }),
      },
    });
    const tunnel = jsonbody.tunnel;

    if (tunnel) {
      pgconfig.tunnel = new SSHTunnel({
        host: tunnel.host,
        port: tunnel.port,
        user: tunnel.user,
        knownHostPublicKey: tunnel.knownHostPublicKey
          ? tunnel.knownHostPublicKey
          : undefined,
      });
      if (tunnel.privateKey) {
        pgconfig.tunnel.authentication = new SSHAuthentication({
          authConfig: {
            case: 'privateKey',
            value: new SSHPrivateKey({
              value: tunnel.privateKey,
              passphrase: tunnel.passphrase,
            }),
          },
        });
      } else if (tunnel.passphrase) {
        pgconfig.tunnel.authentication = new SSHAuthentication({
          authConfig: {
            case: 'passphrase',
            value: new SSHPassphrase({
              value: tunnel.passphrase,
            }),
          },
        });
      }
    }

    return ctx.client.connections.createConnection(
      new CreateConnectionRequest({
        name: jsonbody.connectionName,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: pgconfig,
          },
        }),
      })
    );
  })(req);
}

export async function PUT(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const jsonbody = await EXISTING_POSTGRES_CONNECTION.validate(
      await req.json()
    );
    const pgconfig = new PostgresConnectionConfig({
      connectionConfig: {
        case: 'connection',
        value: new PostgresConnection({
          host: jsonbody.connection.host,
          name: jsonbody.connection.name,
          user: jsonbody.connection.user,
          pass: jsonbody.connection.pass,
          port: jsonbody.connection.port,
          sslMode: jsonbody.connection.sslMode,
        }),
      },
    });
    const tunnel = jsonbody.tunnel;

    if (tunnel) {
      pgconfig.tunnel = new SSHTunnel({
        host: tunnel.host,
        port: tunnel.port,
        user: tunnel.user,
        knownHostPublicKey: tunnel.knownHostPublicKey
          ? tunnel.knownHostPublicKey
          : undefined,
      });
      if (tunnel.privateKey) {
        pgconfig.tunnel.authentication = new SSHAuthentication({
          authConfig: {
            case: 'privateKey',
            value: new SSHPrivateKey({
              value: tunnel.privateKey,
              passphrase: tunnel.passphrase,
            }),
          },
        });
      } else if (tunnel.passphrase) {
        pgconfig.tunnel.authentication = new SSHAuthentication({
          authConfig: {
            case: 'passphrase',
            value: new SSHPassphrase({
              value: tunnel.passphrase,
            }),
          },
        });
      }
    }
    return ctx.client.connections.updateConnection(
      new UpdateConnectionRequest({
        id: jsonbody.id,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: pgconfig,
          },
        }),
      })
    );
  })(req);
}
