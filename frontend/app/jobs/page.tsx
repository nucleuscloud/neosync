'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useGetJobs } from '@/libs/hooks/useGetJobs';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { getColumns } from './components/DataTable/columns';
import { DataTable } from './components/DataTable/data-table';

export default function Jobs() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Jobs"
          description="Jobs are asynchronous tasks that move, transform, or scan data"
          extraHeading={<NewJobButton />}
        />
      }
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
  const account = useAccount();
  const { isLoading, data, mutate } = useGetJobs(account?.id ?? '');

  if (isLoading) {
    return <SkeletonTable />;
  }

  const jobs = data?.jobs ?? [];

  const columns = getColumns({
    onDeleted() {
      mutate();
    },
  });

  return (
    <div>
      <DataTable columns={columns} data={jobs} />
    </div>
  );
}

interface NewJobButtonProps {}

function NewJobButton(props: NewJobButtonProps): ReactElement {
  const {} = props;
  return (
    <NextLink href={'/new/job'}>
      <Button>New Job </Button>
    </NextLink>
  );
}
