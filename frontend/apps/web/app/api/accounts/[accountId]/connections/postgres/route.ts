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
  UpdateConnectionRequest,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest): Promise<NextResponse> {
  return withNeosyncContext(async (ctx) => {
    const jsonbody = await NEW_POSTGRES_CONNECTION.validate(await req.json());
    return ctx.client.connections.createConnection(
      new CreateConnectionRequest({
        name: jsonbody.connectionName,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: new PostgresConnectionConfig({
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
            }),
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
    return ctx.client.connections.updateConnection(
      new UpdateConnectionRequest({
        id: jsonbody.id,
        connectionConfig: new ConnectionConfig({
          config: {
            case: 'pgConfig',
            value: new PostgresConnectionConfig({
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
            }),
          },
        }),
      })
    );
  })(req);
}
