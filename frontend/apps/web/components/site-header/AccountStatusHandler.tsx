'use client';
import { SystemAppConfig } from '@/app/config/app-config';
import { useQuery } from '@connectrpc/connect-query';
import { UserAccountType } from '@neosync/sdk';
import { isAccountStatusValid } from '@neosync/sdk/connectquery';
import { useAccount } from '../providers/account-provider';
import { Skeleton } from '../ui/skeleton';
import RecordsProgressBar from './RecordsProgressBar';
import Upgrade from './Upgrade';

interface Props {
  systemAppConfig: SystemAppConfig;
}

export function AccountStatusHandler(props: Props) {
  const { account } = useAccount();
  const { systemAppConfig } = props;

  const showRecordsUsage =
    systemAppConfig.isNeosyncCloud &&
    systemAppConfig.isStripeEnabled &&
    account?.type === UserAccountType.PERSONAL;

  const { data: data, isLoading } = useQuery(
    isAccountStatusValid,
    { accountId: account?.id },
    { enabled: !!account?.id && showRecordsUsage }
  );

  if (isLoading) {
    return <Skeleton className="w-[100px] h-8" />;
  }

  return (
    <div className="flex flex-row items-center gap-2">
      {showRecordsUsage && (
        <RecordsProgressBar
          count={Number(data?.usedRecordCount)}
          allowedRecords={Number(data?.allowedRecordCount)}
          identifier={account?.id ?? ''}
          idType={'accountId'}
        />
      )}
      <Upgrade
        calendlyLink={systemAppConfig.calendlyUpgradeLink}
        isNeosyncCloud={systemAppConfig.isNeosyncCloud}
        count={Number(data?.usedRecordCount)}
        allowedRecords={Number(data?.allowedRecordCount)}
      />
    </div>
  );
}
