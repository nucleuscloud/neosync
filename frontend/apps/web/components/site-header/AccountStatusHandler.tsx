'use client';
import { SystemAppConfig } from '@/app/config/app-config';
import { cn } from '@/libs/utils';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useQuery } from '@connectrpc/connect-query';
import { AccountStatus, UserAccountService } from '@neosync/sdk';
import { differenceInDays } from 'date-fns';
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
    UserAccountService.method.isAccountStatusValid,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );

  if (isLoading) {
    return <Skeleton className="w-[100px] h-8" />;
  }

  const showTrialCountdown =
    systemAppConfig.isNeosyncCloud &&
    (data?.accountStatus == AccountStatus.ACCOUNT_TRIAL_ACTIVE ||
      data?.accountStatus == AccountStatus.ACCOUNT_TRIAL_EXPIRED);

  const trialEndDate = data?.trialExpiresAt
    ? timestampDate(data.trialExpiresAt)
    : new Date();

  const daysRemaining = Math.max(differenceInDays(trialEndDate, new Date()));

  return (
    <div className="flex flex-row items-center gap-2">
      {showTrialCountdown && (
        <TrialCountdown
          isExpired={data?.accountStatus == AccountStatus.ACCOUNT_TRIAL_EXPIRED}
          isAlmostExpired={daysRemaining <= 3}
          daysRemaining={daysRemaining}
        />
      )}
      <Upgrade
        calendlyLink={systemAppConfig.calendlyUpgradeLink}
        isNeosyncCloud={systemAppConfig.isNeosyncCloud}
        isAccountStatusValidResp={data}
        isLoading={isLoading}
      />
    </div>
  );
}

interface TrialCountdownProps {
  isExpired: boolean;
  isAlmostExpired: boolean;
  daysRemaining: number;
}

function TrialCountdown(props: TrialCountdownProps) {
  const { isExpired, isAlmostExpired, daysRemaining } = props;

  return (
    <div
      className={cn(
        isExpired
          ? 'border-red-700'
          : isAlmostExpired
            ? 'border-yellow-500'
            : ' border-blue-400 dark:border-blue-700',
        'border flex items-center gap-2 h-8  rounded-md px-2 py-1'
      )}
    >
      <div className="relative flex items-center">
        <div
          className={cn(
            isExpired
              ? 'bg-red-600'
              : isAlmostExpired
                ? 'bg-yellow-600'
                : ' border-blue-400 dark:border-blue-700',
            'absolute animate-ping h-2.5 w-2.5 rounded-full bg-blue-400 opacity-75'
          )}
        />
        <div
          className={cn(
            isExpired
              ? 'bg-red-600'
              : isAlmostExpired
                ? 'bg-yellow-600'
                : 'bg-blue-700',
            'relative h-2.5 w-2.5 rounded-full'
          )}
        />
      </div>
      <div className="text-xs ">
        {isExpired
          ? 'Trial Expired'
          : `${daysRemaining} day${daysRemaining !== 1 ? 's' : ''} left in your trial`}
      </div>
    </div>
  );
}
