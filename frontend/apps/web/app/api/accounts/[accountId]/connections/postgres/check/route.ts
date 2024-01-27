import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  POSTGRES_CONNECTION,
  SSH_TUNNEL_FORM_SCHEMA,
} from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  ConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = (await req.json()) ?? {};
    const db = await POSTGRES_CONNECTION.validate(body.db ?? {});
    const tunnel = await SSH_TUNNEL_FORM_SCHEMA.validate(body.tunnel ?? {});

    const pgconfig = new PostgresConnectionConfig({
      connectionConfig: {
        case: 'connection',
        value: new PostgresConnection({
          host: db.host,
          name: db.name,
          user: db.user,
          pass: db.pass,
          port: db.port,
          sslMode: db.sslMode,
        }),
      },
    });

    if (tunnel && tunnel.host && tunnel.port && tunnel.user) {
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

    return ctx.client.connections.checkConnectionConfig(
      new CheckConnectionConfigRequest({
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
