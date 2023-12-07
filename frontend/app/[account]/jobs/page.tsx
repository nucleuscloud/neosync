'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useGetJobStatuses } from '@/libs/hooks/useGetJobStatuses';
import { useGetJobs } from '@/libs/hooks/useGetJobs';
import { JobStatus } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement } from 'react';
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
  const { isLoading, data, mutate } = useGetJobs(account?.id ?? '');
  const { data: statusData } = useGetJobStatuses(account?.id ?? '');

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
    return {
      ...j,
      status: statusJobMap[j.id] || JobStatus.UNSPECIFIED,
      type: j.source?.options?.config.case == 'generate' ? 'Generate' : 'Sync',
    };
  });

  const columns = getColumns({
    accountName: account?.name ?? '',
    onDeleted() {
      mutate();
    },
  });

  return (
    <div>
      <DataTable columns={columns} data={jobData} />
    </div>
  );
}

function NewJobButton(): ReactElement {
  const { account } = useAccount();
  return (
    <NextLink href={`/${account?.name}/new/job`}>
      <Button>
        <ButtonText leftIcon={<PlusIcon />} text="New Job" />
      </Button>
    </NextLink>
  );
}
