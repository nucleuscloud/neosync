'use client';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import ko from '@getkoala/react';
import { useSession } from 'next-auth/react';
import { ReactElement, useEffect } from 'react';
import { useAccount } from './account-provider';

export default function KoalaIdentifier(): ReactElement {
  const { data: systemAppConfig, isLoading: isSystemAppConfigLoading } =
    useGetSystemAppConfig();
  const { data: session } = useSession();
  const { account, isLoading: isAccountLoading } = useAccount();
  const user = session?.user;

  useEffect(() => {
    if (
      typeof window !== 'undefined' &&
      !isSystemAppConfigLoading &&
      systemAppConfig?.koala?.enabled
    ) {
      ko.init(process.env.KOALA_KEY);
    }
  }, [
    systemAppConfig?.koala?.enabled,
    systemAppConfig?.koala?.key,
    isSystemAppConfigLoading,
  ]);

  useEffect(() => {
    if (isAccountLoading || isSystemAppConfigLoading) {
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
