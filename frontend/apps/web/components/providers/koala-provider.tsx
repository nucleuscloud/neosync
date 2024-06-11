'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import ko, { KoalaProvider } from '@getkoala/react';
import { useSession } from 'next-auth/react';
import { ReactElement, ReactNode, useEffect } from 'react';
import { useAccount } from './account-provider';

export default function KoalaIdentifier(): ReactElement {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: session } = useSession();
  const { account, isLoading: isAccountLoading } = useAccount();
  const user = session?.user;

  useEffect(() => {
    if (
      isAccountLoading ||
      (!isSystemAppConfigLoading &&
        typeof window !== 'undefined' &&
        systemAppConfig?.koala.enabled &&
        systemAppConfig?.koala?.key)
    ) {
      return;
    }

    ko!.identify(user?.email, {
      account: account?.name,
      name: user?.name,
      neosyncCloud: systemAppConfig?.isNeosyncCloud ?? false,
    });
  }, [user?.name, systemAppConfig?.isNeosyncCloud]);

  return <></>;
}

interface KProps {
  children: ReactNode;
}

export function KProvider({ children }: KProps) {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();

  if (
    typeof window !== 'undefined' &&
    !isSystemAppConfigLoading &&
    systemAppConfig?.koala?.enabled &&
    systemAppConfig?.koala?.key
  ) {
    return children;
  }

  return (
    <KoalaProvider publicApiKey={systemAppConfig?.koala.key}>
      {children}
    </KoalaProvider>
  );
}
