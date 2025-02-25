'use client';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { PageProps } from '@/components/types';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import DailyMetricCount from '@/components/usage/DailyMetricCount';
import MetricCount from '@/components/usage/MetricCount';
import UsagePeriodSelector from '@/components/usage/UsagePeriodSelector';
import {
  UsagePeriod,
  getPeriodLabel,
  periodToDateRange,
} from '@/components/usage/util';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { RangedMetricName } from '@neosync/sdk';
import { format } from 'date-fns';
import { ReactElement, useState, use } from 'react';

export default function UsagePage(props: PageProps): ReactElement<any> {
  const params = use(props.params);
  const id = params?.id ?? '';
  const [period, setPeriod] = useState<UsagePeriod>('current');
  const { data: configData, isLoading } = useGetSystemAppConfig();
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
    <div className="job-details-usage-container flex flex-col gap-5">
      <SubPageHeader
        header="Usage"
        description={`${getPeriodLabel(period)}: ${getDateRangeLabel(start, end)}`}
        extraHeading={
          <UsagePeriodSelector period={period} setPeriod={setPeriod} />
        }
      />
      <div className="flex">
        <MetricCount
          period={period}
          metric={RangedMetricName.INPUT_RECEIVED}
          idtype="jobId"
          identifier={id}
        />
      </div>
      <div>
        <DailyMetricCount
          period={period}
          metric={RangedMetricName.INPUT_RECEIVED}
          idtype="jobId"
          identifier={id}
        />
      </div>
    </div>
  );
}

function getDateRangeLabel(start: Date, end: Date): string {
  return `${format(start, 'MM/dd/yy')} - ${format(end, 'MM/dd/yy')}`;
}
