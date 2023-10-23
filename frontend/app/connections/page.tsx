'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useGetConnections } from '@/libs/hooks/useGetConnections';
import { PlusIcon } from '@radix-ui/react-icons';
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
  const { isLoading, data, mutate } = useGetConnections(account?.id ?? '');

  if (isLoading) {
    return <SkeletonTable />;
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
      <Button>
        <ButtonText leftIcon={<PlusIcon />} text="New Connection" />
      </Button>
    </NextLink>
  );
}
