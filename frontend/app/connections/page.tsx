'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { getColumns } from './components/ConnectionsTable/columns';
import { DataTable } from './components/ConnectionsTable/data-table';

export default function Connections(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Connections"
          description="Create and manage connections to send and receive data"
          extraHeading={<NewConnectionButton />}
        />
      }
      containerClassName="connections-page"
    >
      <ConnectionTable />
    </OverviewContainer>
  );
}

interface ConnectionTableProps {}

function ConnectionTable(props: ConnectionTableProps): ReactElement {
  const {} = props;
  const account = useAccount();
  console.log(account);
  const { isLoading, data, mutate } = useGetConnections(account?.id ?? '');

  if (isLoading) {
    return <Skeleton />;
  }

  const connections = data?.connections ?? [];

  const columns = getColumns({
    onConnectionDeleted() {
      mutate();
    },
  });

  return (
    <div>
      <DataTable columns={columns} data={connections} />
    </div>
  );
}

interface NewConnectionButtonprops {}

function NewConnectionButton(props: NewConnectionButtonprops): ReactElement {
  const {} = props;
  return (
    <NextLink href={'/new/connection'}>
      <Button>New Connection</Button>
    </NextLink>
  );
}
