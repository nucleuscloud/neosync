import { withNeosyncContext } from '@/api-only/neosync-context';
import {
  CheckConnectionConfigRequest,
  ConnectionConfig,
  PostgresConnection,
  PostgresConnectionConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { POSTGRES_CONNECTION } from '@/yup-validations/connections';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const jsonbody = await POSTGRES_CONNECTION.validate(await req.json());
    return ctx.connectionClient.checkConnectionConfig(
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
