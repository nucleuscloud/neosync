'use client';
import { SystemAppConfig } from '@/app/config/app-config';
import { useQuery } from '@connectrpc/connect-query';
import { RangedMetricName, UserAccountType } from '@neosync/sdk';
import { getMetricCount } from '@neosync/sdk/connectquery';
import { useState } from 'react';
import { useAccount } from '../providers/account-provider';
import { Skeleton } from '../ui/skeleton';
import { dateToNeoDate, periodToDateRange, UsagePeriod } from '../usage/util';
import RecordsProgressBar from './RecordsProgressBar';
import Upgrade from './Upgrade';

interface Props {
  systemAppConfig: SystemAppConfig;
}

export function AccountStatusHandler(props: Props) {
  const { account } = useAccount();
  const idtype = 'accountId';
  const { systemAppConfig } = props;
  const [period, _] = useState<UsagePeriod>('current');
  const metric = RangedMetricName.INPUT_RECEIVED;
  const [start, end] = periodToDateRange(period);

  const identifier = account?.id ?? '';

  const showRecordsUsage =
    systemAppConfig.isNeosyncCloud &&
    systemAppConfig.isStripeEnabled &&
    account?.type === UserAccountType.PERSONAL;

  const { data: metricCountData, isLoading } = useQuery(
    getMetricCount,
    {
      metric,
      startDay: dateToNeoDate(start),
      endDay: dateToNeoDate(end),
      identifier:
        idtype === 'accountId'
          ? { case: 'accountId', value: identifier }
          : idtype === 'jobId'
            ? { case: 'jobId', value: identifier }
            : idtype === 'runId'
              ? { case: 'runId', value: identifier }
              : undefined,
    },
    {
      enabled:
        showRecordsUsage && !!metric && !!identifier && !!idtype && !!period,
    }
  );

  const count =
    metricCountData?.count !== undefined ? Number(metricCountData.count) : 0;

  if (isLoading) {
    return <Skeleton className="w-[100px] h-8" />;
  }

  return (
    <div className="flex flex-row items-center gap-2">
      {showRecordsUsage && (
        <RecordsProgressBar
          count={count}
          identifier={identifier}
          idtype={idtype}
        />
      )}
      <Upgrade
        calendlyLink={systemAppConfig.calendlyUpgradeLink}
        isNeosyncCloud={systemAppConfig.isNeosyncCloud}
        count={count}
      />
    </div>
  );
}
