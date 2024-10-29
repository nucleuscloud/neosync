'use client';
import { SystemAppConfig } from '@/app/config/app-config';
import { cn } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { AccountStatus, IsAccountStatusValidResponse } from '@neosync/sdk';
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

  const showTrialCountdown =
    systemAppConfig.isNeosyncCloud &&
    (data?.accountStatus == AccountStatus.ACCOUNT_TRIAL_ACTIVE ||
      data?.accountStatus == AccountStatus.ACCOUNT_TRIAL_EXPIRED);

  return (
    <div className="flex flex-row items-center gap-2">
      {showTrialCountdown && (
        <TrialCountdown
          trialEndDate={new Date(
            data?.trialExpiresAt?.toDate() ?? Date.now()
          ).getTime()}
          isAccountStatusValidResp={data}
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
  trialEndDate: number;
  isAccountStatusValidResp: IsAccountStatusValidResponse | undefined;
}

function TrialCountdown(props: TrialCountdownProps) {
  const { trialEndDate, isAccountStatusValidResp } = props;

  const now = Date.now();
  const daysRemaining = Math.max(
    0,
    Math.ceil((trialEndDate - now) / (1000 * 60 * 60 * 24))
  );

  const isExpired =
    isAccountStatusValidResp?.accountStatus ==
    AccountStatus.ACCOUNT_TRIAL_EXPIRED;
  const isAlmostExpired = daysRemaining <= 3;

  if (isExpired)
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
