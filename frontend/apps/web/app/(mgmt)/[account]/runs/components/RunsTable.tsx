import EmptyState, { EmptyStateLinkButton } from '@/components/EmptyState';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import {
  JobRunsAutoRefreshInterval,
  onJobRunsAutoRefreshInterval,
  onJobRunsPaused,
} from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { getJobRuns, getJobs } from '@neosync/sdk/connectquery';
import { ArrowTopRightIcon } from '@radix-ui/react-icons';
import { ReactElement, useMemo, useState } from 'react';
import { GoStack } from 'react-icons/go';
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
  const {
    isLoading,
    data,
    refetch: mutate,
    isFetching: isValidating,
  } = useQuery(
    getJobRuns,
    { id: { case: 'accountId', value: account?.id ?? '' } },
    {
      enabled() {
        return !!account?.id && !onJobRunsPaused(refreshInterval);
      },
      refetchInterval() {
        return onJobRunsAutoRefreshInterval(refreshInterval);
      },
    }
  );

  const {
    data: jobsData,
    refetch: jobsMutate,
    isLoading: isJobsLoading,
    isFetching: isJobsValidating,
  } = useQuery(getJobs, { accountId: account?.id }, { enabled: !!account?.id });

  const jobs = jobsData?.jobs ?? [];

  // must be memoized otherwise it causes columns to re-render endlessly when hovering over links within the table
  const jobNameMap = useMemo(() => {
    return jobs.reduce(
      (prev, curr) => {
        return { ...prev, [curr.id]: curr.name };
      },
      {} as Record<string, string>
    );
  }, [isJobsLoading, isJobsValidating]);

  const columns = useMemo(
    () =>
      getColumns({
        onDeleted() {
          mutate();
        },
        accountName: account?.name ?? '',
        jobNameMap: jobNameMap,
      }),
    [account?.name ?? '', jobNameMap]
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
      {runs.length == 0 ? (
        <EmptyState
          title="No Job Runs yet"
          description="Runs are instances of a job that have been executed."
          icon={<GoStack className="w-8 h-8 text-primary" />}
          extra={
            <EmptyStateLinkButton
              buttonText="Trigger a Job to see a Run"
              buttonIcon={<ArrowTopRightIcon />}
              buttonIconSide="right"
              href={`/${account?.name}/jobs`}
            />
          }
        />
      ) : (
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
      )}
    </div>
  );
}
