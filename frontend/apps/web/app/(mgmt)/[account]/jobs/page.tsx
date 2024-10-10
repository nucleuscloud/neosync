'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import EmptyState from '@/components/EmptyState';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import { JobStatus } from '@neosync/sdk';
import { getJobs, getJobStatuses } from '@neosync/sdk/connectquery';
import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useMemo } from 'react';
import { GrPowerCycle } from 'react-icons/gr';
import { getColumns } from './components/DataTable/columns';
import { DataTable } from './components/DataTable/data-table';

export default function Jobs() {
  return (
    <OverviewContainer
      Header={<PageHeader header="Jobs" extraHeading={<NewJobButton />} />}
      containerClassName="jobs-page"
    >
      <div>
        <JobTable />
      </div>
    </OverviewContainer>
  );
}

interface JobTableProps {}

function JobTable(props: JobTableProps): ReactElement {
  const {} = props;
  const { account } = useAccount();
  const {
    isLoading,
    data,
    refetch: mutate,
  } = useQuery(getJobs, { accountId: account?.id }, { enabled: !!account?.id });
  const { data: statusData } = useQuery(
    getJobStatuses,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const columns = useMemo(
    () =>
      getColumns({
        accountName: account?.name ?? '',
        onDeleted() {
          mutate();
        },
      }),
    [account?.name ?? '']
  );

  if (isLoading) {
    return <SkeletonTable />;
  }

  const jobs = data?.jobs ?? [];
  const statusJobMap =
    statusData?.statuses.reduce(
      (prev, curr) => {
        return { ...prev, [curr.jobId]: curr.status };
      },
      {} as Record<string, JobStatus>
    ) || {};

  const jobData = jobs.map((j) => {
    let jobtype = 'Sync';
    if (j.source?.options?.config.case === 'generate') {
      jobtype = 'Generate';
    } else if (j.source?.options?.config.case === 'aiGenerate') {
      jobtype = 'AI Generate';
    }
    return {
      ...j,
      status: statusJobMap[j.id] || JobStatus.UNSPECIFIED,
      type: jobtype,
    };
  });

  return (
    <div>
      {jobData.length == 0 ? (
        <EmptyState
          title="No Jobs yet"
          description="Jobs are async workflows that transform data and sync it between source and destination systems."
          buttonText="Create your first Job"
          icon={<GrPowerCycle className="w-8 h-8 text-primary" />}
          href={`/${account?.name}/new/job`}
        />
      ) : (
        <DataTable columns={columns} data={jobData} />
      )}
    </div>
  );
}

function NewJobButton(): ReactElement {
  const { account } = useAccount();
  const posthog = usePostHog();
  return (
    <NextLink
      href={`/${account?.name}/new/job`}
      onClick={() => {
        posthog.capture('clicked_new_job_button');
      }}
    >
      <Button onClick={() => {}}>
        <ButtonText leftIcon={<PlusIcon />} text="New Job" />
      </Button>
    </NextLink>
  );
}
