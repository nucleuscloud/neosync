'use client';

import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactElement, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}
export default function ConnectProvider(props: Props): ReactElement {
  const { children } = props;
  const connectTransport = createConnectTransport({
    baseUrl: 'http://localhost:8080',
    // credentials: 'include',
    // credentials: 'omit',
    interceptors: [
      (next) => async (req) => {
        // const accessToken = await getAccessToken();
        req.header.set('Authorization', `Bearer ${'my-access-token'}`);
        return next(req);
      },
    ],
  });
  const queryClient = new QueryClient();

  return (
    <TransportProvider transport={connectTransport}>
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </TransportProvider>
  );
}
