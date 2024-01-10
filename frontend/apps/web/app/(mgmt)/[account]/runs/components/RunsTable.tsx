import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  JobRunsAutoRefreshInterval,
  onJobRunsAutoRefreshInterval,
  onJobRunsPaused,
  useGetJobRuns,
} from '@/libs/hooks/useGetJobRuns';
import { useGetJobs } from '@/libs/hooks/useGetJobs';
import { ReactElement, useMemo, useState } from 'react';
import { getColumns } from './JobRunsTable/columns';
import { DataTable } from './JobRunsTable/data-table';

const INTERVAL_SELECT_OPTIONS: JobRunsAutoRefreshInterval[] = [
  'off',
  '10s',
  '30s',
  '1m',
  '5m',
];

interface RunsTableProps {}

export default function RunsTable(props: RunsTableProps): ReactElement {
  const {} = props;
  const { account } = useAccount();
  const [refreshInterval, setAutoRefreshInterval] =
    useState<JobRunsAutoRefreshInterval>('1m');
  const { isLoading, data, mutate, isValidating } = useGetJobRuns(
    account?.id ?? '',
    {
      refreshIntervalFn: () => onJobRunsAutoRefreshInterval(refreshInterval),
      isPaused: () => onJobRunsPaused(refreshInterval),
    }
  );

  const { data: jobsData, mutate: jobsMutate } = useGetJobs(account?.id ?? '');

  const jobNameMap =
    jobsData?.jobs.reduce(
      (prev, curr) => {
        return { ...prev, [curr.id]: curr.name };
      },
      {} as Record<string, string>
    ) || {};

  const columns = useMemo(
    () =>
      getColumns({
        onDeleted() {
          mutate();
        },
        accountId: account?.id ?? '',
        accountName: account?.name ?? '',
        jobNameMap: jobNameMap,
      }),
    [account?.id ?? '', account?.name ?? '', jobNameMap]
  );

  if (isLoading) {
    return <SkeletonTable />;
  }

  const runs = data?.jobRuns ?? [];

  function refreshClick(): void {
    mutate();
    jobsMutate();
  }

  return (
    <div>
      <DataTable
        columns={columns}
        data={runs}
        refreshInterval={refreshInterval}
        onAutoRefreshIntervalChange={(newVal: JobRunsAutoRefreshInterval) =>
          setAutoRefreshInterval(newVal)
        }
        autoRefreshIntervalOptions={INTERVAL_SELECT_OPTIONS}
        onRefreshClick={refreshClick}
        isRefreshing={isValidating}
      />
    </div>
  );
}
