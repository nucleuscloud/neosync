'use client';
import { cn } from '@/libs/utils';
import { GetMetricCountRequest } from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useAccount } from '../providers/account-provider';
import { Button } from '../ui/button';
import { Progress } from '../ui/progress';

interface Props {
  count: number;
  idtype: MetricsIdentifierCase;
  identifier: string;
}

export default function RecordsProgressBar(props: Props): ReactElement {
  const { count, idtype, identifier } = props;
  const { account } = useAccount();

  const router = useRouter();

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
      className={cn(count > totalRecords && 'bg-orange-200 dark:bg-orange-700')}
    >
      <div className="flex flex-row items-center gap-2 sm:w-60">
        <span className="text-sm text-nowrap dark:text-gray-900">
          Records used
        </span>
        <Progress value={percentageUsed} className="w-[60%] hidden sm:flex" />
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

export function formatNumber(num: number): string {
  const browserLanguages = navigator.languages;
  const formatter = new Intl.NumberFormat(browserLanguages, {
    notation: 'compact',
    compactDisplay: 'short',
    maximumFractionDigits: 1,
  });
  return formatter.format(num);
}
