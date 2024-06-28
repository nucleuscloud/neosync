'use client';

import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useSession } from 'next-auth/react';
import { ReactElement, ReactNode, useMemo } from 'react';

interface Props {
  children: ReactNode;
  apiBaseUrl: string;
}
export default function ConnectProvider(props: Props): ReactElement {
  const { children, apiBaseUrl } = props;
  const { data } = useSession();
  const connectTransport = useMemo(() => {
    return createConnectTransport({
      baseUrl: apiBaseUrl,
      interceptors: [
        (next) => async (req) => {
          if (data?.accessToken) {
            req.header.set('Authorization', `Bearer ${data.accessToken}`);
          }
          return next(req);
        },
      ],
    });
  }, [apiBaseUrl, data?.accessToken]);

  const queryClient = useMemo(() => new QueryClient(), []);

  return (
    <TransportProvider transport={connectTransport}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TransportProvider>
  );
}
