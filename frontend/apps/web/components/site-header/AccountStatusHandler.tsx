'use client';
import { SystemAppConfig } from '@/app/config/app-config';
import { useQuery } from '@connectrpc/connect-query';
import { isAccountStatusValid } from '@neosync/sdk/connectquery';
import { useAccount } from '../providers/account-provider';
import { Skeleton } from '../ui/skeleton';
import Upgrade from './Upgrade';

interface Props {
  systemAppConfig: SystemAppConfig;
}

export function AccountStatusHandler(props: Props) {
  const { systemAppConfig } = props;
  const { account } = useAccount();

  const { data: data, isLoading } = useQuery(
    isAccountStatusValid,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );

  if (isLoading) {
    return <Skeleton className="w-[100px] h-8" />;
  }

  return (
    <div className="flex flex-row items-center gap-2">
      <Upgrade
        calendlyLink={systemAppConfig.calendlyUpgradeLink}
        isNeosyncCloud={systemAppConfig.isNeosyncCloud}
        isAccountStatusValidResp={data}
        isLoading={isLoading}
      />
    </div>
  );
}
