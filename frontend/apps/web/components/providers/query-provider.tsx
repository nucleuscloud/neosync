'use client';
import { getErrorMessage } from '@/util/util';
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

export function TanstackQueryProvider404(props: Props): ReactElement {
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

  // if (!browserQueryClient) {
  browserQueryClient = new QueryClient({
    queryCache: new QueryCache({
      // good blog here: https://tkdodo.eu/blog/react-query-error-handling
      onError(error, query) {
        console.log('error', error);
        // toast.error('Something went wrong', {
        //   description: getErrorMessage(error),
        //   id: query.queryKey.toString(),
        // });
      },
    }),
    defaultOptions: {},
  });
  // }

  return (
    <QueryClientProvider client={browserQueryClient}>
      {children}
    </QueryClientProvider>
  );
}
