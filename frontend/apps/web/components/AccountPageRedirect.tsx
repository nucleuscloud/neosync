'use client';

import Error from 'next/error';
import { useRouter } from 'next/navigation';
import { ReactNode, useEffect, type JSX } from 'react';
import { useAccount } from './providers/account-provider';
import { Skeleton } from './ui/skeleton';

interface Props {
  children: ReactNode;
}

export default function AccountPageRedirect(props: Props): JSX.Element {
  const { children } = props;

  const router = useRouter();
  const { account, isLoading } = useAccount();

  useEffect(() => {
    if (isLoading || !account?.name) {
      return;
    }
    router.push(`/${account.name}/jobs`);
  }, [isLoading, account?.name, account?.id]);

  if (isLoading) {
    return <Skeleton className="w-full h-full py-2" />;
  }

  if (!account) {
    return <Error statusCode={404} />;
  }

  return <>{children}</>;
}
