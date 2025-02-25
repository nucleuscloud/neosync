'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import DailyMetricCount from '@/components/usage/DailyMetricCount';
import MetricCount from '@/components/usage/MetricCount';
import UsagePeriodSelector from '@/components/usage/UsagePeriodSelector';
import {
  UsagePeriod,
  getDateRangeLabel,
  getPeriodLabel,
  periodToDateRange,
} from '@/components/usage/util';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { RangedMetricName } from '@neosync/sdk';
import { ReactElement, useState } from 'react';

export default function UsagePage(): ReactElement<any> {
  const [period, setPeriod] = useState<UsagePeriod>('current');
  const { data: configData, isLoading } = useGetSystemAppConfig();
  const { account } = useAccount();
  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }
  if (!configData?.isMetricsServiceEnabled) {
    return (
      <div>
        <Alert variant="warning">
          <AlertTitle>Metrics are not currently enabled</AlertTitle>
          <AlertDescription>
            To enable them, please update Neosync configuration or contact your
            system administrator.
          </AlertDescription>
        </Alert>
      </div>
    );
  }
  const [start, end] = periodToDateRange(period);
  return (
    <div className="flex flex-col gap-5">
      <SubPageHeader
        header="Usage"
        description="See periodic usage for this account"
        subHeadings={`${getPeriodLabel(period)}: ${getDateRangeLabel(start, end)}`}
        extraHeading={
          <UsagePeriodSelector period={period} setPeriod={setPeriod} />
        }
      />
      <div className="flex">
        <MetricCount
          period={period}
          metric={RangedMetricName.INPUT_RECEIVED}
          idtype="accountId"
          identifier={account?.id ?? ''}
        />
      </div>
      <div>
        <DailyMetricCount
          period={period}
          metric={RangedMetricName.INPUT_RECEIVED}
          idtype="accountId"
          identifier={account?.id ?? ''}
        />
      </div>
    </div>
  );
}
