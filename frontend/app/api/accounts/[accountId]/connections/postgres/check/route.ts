import { withNeosyncContext } from '@/api-only/neosync-context';
import { POSTGRES_CONNECTION } from '@/yup-validations/connections';
import {
  CheckConnectionConfigRequest,
  ConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const jsonbody = await POSTGRES_CONNECTION.validate(await req.json());
    return ctx.client.connections.checkConnectionConfig(
      new CheckConnectionConfigRequest({
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: new PostgresConnectionConfig({
              connectionConfig: {
                case: 'connection',
                value: new PostgresConnection({
                  host: jsonbody.host,
                  name: jsonbody.name,
                  user: jsonbody.user,
                  pass: jsonbody.pass,
                  port: jsonbody.port,
                  sslMode: jsonbody.sslMode,
                }),
              },
            }),
          },
        }),
      })
    );
  })(req);
}
