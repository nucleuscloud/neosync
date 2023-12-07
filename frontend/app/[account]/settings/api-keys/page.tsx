'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useGetAccountApiKeys } from '@/libs/hooks/useGetAccountApiKeys';
import { PlusIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import SubNav, { getNavSettings } from '../temporal/components/SubNav';
import { getColumns } from './components/ApiKeysTable/columns';
import { DataTable } from './components/ApiKeysTable/data-table';

export default function ApiKeys(): ReactElement {
  const { account } = useAccount();
  return (
    <OverviewContainer
      Header={
        <PageHeader header="API Keys" extraHeading={<NewApiKeyButton />} />
      }
      containerClassName="apikeys-settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={getNavSettings(account?.name ?? '')} />
        </div>
        <ApiKeyTable />
      </div>
    </OverviewContainer>
  );
}

interface ApiKeyTableProps {}
function ApiKeyTable(props: ApiKeyTableProps): ReactElement {
  const {} = props;
  const { account } = useAccount();
  const { isLoading, data, mutate } = useGetAccountApiKeys(account?.id ?? '');

  if (isLoading) {
    return <SkeletonTable />;
  }

  const apiKeys = data?.apiKeys ?? [];

  const columns = getColumns({
    accountName: account?.name ?? '',
    onDeleted() {
      mutate();
    },
  });

  return (
    <div>
      <DataTable columns={columns} data={apiKeys} />
    </div>
  );
}

interface NewApiKeyButtonProps {}

function NewApiKeyButton(props: NewApiKeyButtonProps): ReactElement {
  const {} = props;
  const { account } = useAccount();

  return (
    <Link href={`/${account?.name}/new/api-key`}>
      <Button type="button">
        <ButtonText leftIcon={<PlusIcon />} text="New API Key" />
      </Button>
    </Link>
  );
}
