'use client';

import { TransportProvider } from '@connectrpc/connect-query';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ReactElement, ReactNode, useMemo } from 'react';

interface Props {
  children: ReactNode;
  apiBaseUrl: string;
}

export default function ConnectProvider(props: Props): ReactElement<any> {
  const { children, apiBaseUrl } = props;
  const connectTransport = useMemo(() => {
    return createConnectTransport({
      baseUrl: apiBaseUrl,
    });
  }, [apiBaseUrl]);

  return (
    <TransportProvider transport={connectTransport}>
      {children}
    </TransportProvider>
  );
}
