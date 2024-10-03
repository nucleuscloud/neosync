'use client';
import { cn } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { GetMetricCountRequest, RangedMetricName } from '@neosync/sdk';
import { getMetricCount } from '@neosync/sdk/connectquery';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useAccount } from '../providers/account-provider';
import { Button } from '../ui/button';
import { Progress } from '../ui/progress';
import { Skeleton } from '../ui/skeleton';
import { dateToNeoDate, periodToDateRange, UsagePeriod } from '../usage/util';

interface Props {
  identifier: string;
  idtype: MetricsIdentifierCase;
}

export default function RecordsProgressBar(props: Props): ReactElement {
  const { identifier, idtype } = props;
  const [period, _] = useState<UsagePeriod>('current');
  const metric = RangedMetricName.INPUT_RECEIVED;
  const { account } = useAccount();

  const router = useRouter();

  const [start, end] = periodToDateRange(period);

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
      enabled: !!metric && !!identifier && !!idtype && !!period,
    }
  );

  const formatNumber = (num: number): string => {
    const browserLanguages = navigator.languages;
    const formatter = new Intl.NumberFormat(browserLanguages, {
      notation: 'compact',
      compactDisplay: 'short',
      maximumFractionDigits: 1,
    });
    return formatter.format(num);
  };

  if (isLoading) {
    return <Skeleton className="w-[100px] h-8" />;
  }

  const count =
    metricCountData?.count !== undefined ? Number(metricCountData.count) : 0;

  const totalRecords = 20000;
  const percentageUsed = (count / totalRecords) * 100;

  return (
    <Button
      onClick={() => {
        const link = getUsageLink(
          `/${account?.name ?? ''}`,
          idtype,
          identifier
        );
        if (link) {
          return router.push(link);
        }
      }}
      variant="outline"
      className={cn(count > totalRecords && 'bg-orange-200 dark:bg-orange-500')}
    >
      <div className="flex flex-row items-center gap-2 w-60">
        <span className="text-sm text-nowrap">Records used </span>
        <Progress value={percentageUsed} className="w-[60%]" />
        <span className="text-sm">
          {formatNumber(count)}/{formatNumber(totalRecords)}
        </span>
      </div>
    </Button>
  );
}

function getUsageLink(
  basePath: string,
  idtype: MetricsIdentifierCase,
  identifier: string
): string | null {
  if (idtype === 'accountId') {
    return `${basePath}/settings/usage`;
  }
  if (idtype === 'jobId') {
    return `${basePath}/jobs/${identifier}/usage`;
  }
  if (idtype === 'runId') {
    return `${basePath}/runs/${identifier}/usage`;
  }
  return null;
}

// helper fund to extract case types for metric identifiers
type ExtractCase<T> = T extends { case: infer U } ? U : never;

type MetricsIdentifierCase = NonNullable<
  ExtractCase<GetMetricCountRequest['identifier']>
>;
