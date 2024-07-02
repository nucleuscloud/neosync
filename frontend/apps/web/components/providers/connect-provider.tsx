'use client';

import { getErrorMessage } from '@/util/util';
import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import {
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query';
import { useSession } from 'next-auth/react';
import { ReactElement, ReactNode, useMemo } from 'react';
import { useToast } from '../ui/use-toast';

interface Props {
  children: ReactNode;
  apiBaseUrl: string;
}

export default function ConnectProvider(props: Props): ReactElement {
  const { children, apiBaseUrl } = props;
  const { data } = useSession();
  const { toast } = useToast();
  const connectTransport = useMemo(() => {
    return createConnectTransport({
      baseUrl: apiBaseUrl,
    });
  }, [apiBaseUrl, data?.accessToken]);

  const queryClient = useMemo(
    () =>
      new QueryClient({
        queryCache: new QueryCache({
          // good blog here: https://tkdodo.eu/blog/react-query-error-handling
          onError(error, query) {
            toast({
              title: 'Something went wrong',
              description: getErrorMessage(error),
              variant: 'destructive',
              key: query.queryKey.toString(),
            });
          },
        }),
        defaultOptions: {},
      }),
    []
  );

  return (
    <TransportProvider transport={connectTransport}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TransportProvider>
  );
}
