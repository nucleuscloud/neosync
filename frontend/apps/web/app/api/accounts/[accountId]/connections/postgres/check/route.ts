import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  ClientTlsFormValues,
  POSTGRES_CONNECTION,
  SSH_TUNNEL_FORM_SCHEMA,
} from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  ClientTlsConfig,
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

    const pgconfig = new PostgresConnectionConfig();

    if (body.url) {
      pgconfig.connectionConfig = {
        case: 'url',
        value: body.url,
      };
    } else if (body.db) {
      const db = await POSTGRES_CONNECTION.validate(body.db);
      pgconfig.connectionConfig = {
        case: 'connection',
        value: new PostgresConnection(db),
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
      ? await SSH_TUNNEL_FORM_SCHEMA.validate(body.tunnel)
      : null;

    const clientTls = body.clientTls
      ? await ClientTlsFormValues.validate(body.clientTls)
      : null;

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
    if (clientTls?.rootCert || clientTls?.clientKey || clientTls?.clientCert) {
      pgconfig.clientTls = new ClientTlsConfig({
        rootCert: clientTls.rootCert ? clientTls.rootCert : undefined,
        clientKey: clientTls.clientKey ? clientTls.clientKey : undefined,
        clientCert: clientTls.clientCert ? clientTls.clientCert : undefined,
      });
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
