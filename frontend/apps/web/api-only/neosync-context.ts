import { getSystemAppConfig } from '@/app/api/config/config';
import { createConnectTransport } from '@connectrpc/connect-node';
import {
  Code,
  ConnectError,
  GetAccessTokenFn,
  NeosyncClient,
  getNeosyncClient,
} from '@neosync/sdk';
import { NextRequest, NextResponse } from 'next/server';
import { auth } from '../app/api/auth/[...nextauth]/auth';

interface NeosyncContext {
  client: NeosyncClient;
}

type NeosyncApiHandler<T = unknown> = (ctx: NeosyncContext) => Promise<T>;

interface ErrorMessageResponse {
  message: string;
}

export function withNeosyncContext<T = unknown>(
  handler: NeosyncApiHandler<T>
): (req: NextRequest) => Promise<NextResponse<T | ErrorMessageResponse>> {
  const systemAppConfig = getSystemAppConfig();
  return async (_req) => {
    try {
      const neosyncClient = getNeosyncClient({
        getAccessToken: getAccessTokenFn(systemAppConfig.isAuthEnabled),
        getTransport(interceptors) {
          return createConnectTransport({
            baseUrl: getApiBaseUrlFromEnv(),
            httpVersion: '2',
            interceptors: interceptors,
          });
        },
      });
      const output = await handler({ client: neosyncClient });
      if (output instanceof NextResponse) {
        return output;
      }
      return NextResponse.json(output);
    } catch (err) {
      if (err instanceof ConnectError) {
        return NextResponse.json(
          { message: err.message },
          {
            status: translateGrpcCodeToHttpCode(err.code),
          }
        );
      }
      return NextResponse.json(
        {
          message: `unknown error type: ${err}`,
        },
        {
          status: 500,
        }
      );
    }
  };
}

function getAccessTokenFn(
  isAuthEnabled: boolean
): GetAccessTokenFn | undefined {
  if (!isAuthEnabled) {
    return undefined;
  }
  return async (): Promise<string> => {
    const session = await auth();
    const accessToken = session?.accessToken;
    if (!accessToken) {
      throw new Error('no session provided');
    }
    return accessToken;
  };
}

function getApiBaseUrlFromEnv(): string {
  const apiUrl = process.env.NEOSYNC_API_BASE_URL;
  if (!apiUrl) {
    throw new Error('must provide NEOSYNC_API_BASE_URL');
  }
  return apiUrl;
}

function translateGrpcCodeToHttpCode(code: Code): number {
  switch (code) {
    case Code.InvalidArgument:
    case Code.FailedPrecondition:
    case Code.OutOfRange: {
      return 400;
    }
    case Code.Unauthenticated: {
      return 401;
    }
    case Code.PermissionDenied: {
      return 403;
    }
    case Code.Unimplemented:
    case Code.NotFound: {
      return 404;
    }
    case Code.AlreadyExists: {
      return 409;
    }
    case Code.Unavailable: {
      return 503;
    }
    default: {
      return 500;
    }
  }
}
