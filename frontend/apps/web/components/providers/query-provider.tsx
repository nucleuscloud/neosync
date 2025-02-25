'use client';
import { getErrorMessage } from '@/util/util';
import { Code, ConnectError } from '@neosync/sdk';
import {
  isServer,
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query';
import { ReactElement, ReactNode } from 'react';
import { toast } from 'sonner';

interface Props {
  children: ReactNode;
}

let browserQueryClient: QueryClient | undefined = undefined;

export default function TanstackQueryProvider(props: Props): ReactElement {
  const { children } = props;

  if (isServer) {
    const client = new QueryClient({
      defaultOptions: {
        queries: { staleTime: 60 * 1000 },
      },
    });
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  }

  if (!browserQueryClient) {
    browserQueryClient = new QueryClient({
      queryCache: new QueryCache({
        // good blog here: https://tkdodo.eu/blog/react-query-error-handling
        onError(error, query) {
          toast.error('Something went wrong', {
            description: getErrorMessage(error),
            id: query.queryKey.toString(),
          });
        },
      }),
      defaultOptions: {},
    });
  }

  return (
    <QueryClientProvider client={browserQueryClient}>
      {children}
    </QueryClientProvider>
  );
}

// This is a query provider that ignores 404 errors.
export function TanstackQueryProviderIgnore404Errors(
  props: Props
): ReactElement {
  const { children } = props;

  if (isServer) {
    const client = new QueryClient({
      defaultOptions: {
        queries: { staleTime: 60 * 1000 },
      },
    });
    return (
      <QueryClientProvider client={client}>{children}</QueryClientProvider>
    );
  }

  const browserQueryClient404 = new QueryClient({
    queryCache: new QueryCache({
      // good blog here: https://tkdodo.eu/blog/react-query-error-handling
      onError(error, query) {
        if (error instanceof ConnectError && error.code !== Code.NotFound) {
          toast.error('Something went wrong', {
            description: getErrorMessage(error),
            id: query.queryKey.toString(),
          });
        }
      },
    }),
    defaultOptions: {},
  });

  return (
    <QueryClientProvider client={browserQueryClient404}>
      {children}
    </QueryClientProvider>
  );
}
