import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  POSTGRES_CONNECTION,
  SshTunnelFormValues,
} from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  ConnectionConfig,
  MysqlConnection,
  MysqlConnectionConfig,
  SSHAuthentication,
  SSHPassphrase,
  SSHPrivateKey,
  SSHTunnel,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const body = (await req.json()) ?? {};

    const mysqlconfig = new MysqlConnectionConfig();

    if (body.url) {
      mysqlconfig.connectionConfig = {
        case: 'url',
        value: body.url,
      };
    } else if (body.db) {
      const db = await POSTGRES_CONNECTION.validate(body.db);
      mysqlconfig.connectionConfig = {
        case: 'connection',
        value: new MysqlConnection(db),
      };
    } else {
      return new NextResponse(
        JSON.stringify({
          error:
            'The Postgres connection must be either a connection or url case',
        }),
        {
          status: 500,
          headers: {
            'Content-Type': 'application/json',
          },
        }
      );
    }
    const tunnel = body.tunnel
      ? await SshTunnelFormValues.validate(body.tunnel)
      : null;

    if (tunnel && tunnel.host && tunnel.port && tunnel.user) {
      mysqlconfig.tunnel = new SSHTunnel({
        host: tunnel.host,
        port: tunnel.port,
        user: tunnel.user,
        knownHostPublicKey: tunnel.knownHostPublicKey
          ? tunnel.knownHostPublicKey
          : undefined,
      });
      if (tunnel.privateKey) {
        mysqlconfig.tunnel.authentication = new SSHAuthentication({
          authConfig: {
            case: 'privateKey',
            value: new SSHPrivateKey({
              value: tunnel.privateKey,
              passphrase: tunnel.passphrase,
            }),
          },
        });
      } else if (tunnel.passphrase) {
        mysqlconfig.tunnel.authentication = new SSHAuthentication({
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
            case: 'mysqlConfig',
            value: mysqlconfig,
          },
        }),
      })
    );
  })(req);
}
