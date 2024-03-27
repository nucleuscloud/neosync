'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
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

export default function UsagePage(): ReactElement {
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
    <OverviewContainer
      Header={
        <PageHeader
          header="Usage"
          subHeadings={`${getPeriodLabel(period)}: ${getDateRangeLabel(start, end)}`}
          extraHeading={
            <UsagePeriodSelector period={period} setPeriod={setPeriod} />
          }
        />
      }
      containerClassName="usage-page"
    >
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
    </OverviewContainer>
  );
}
