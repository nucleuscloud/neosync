'use client';
import ButtonText from '@/components/ButtonText';
import SubPageHeader from '@/components/headers/SubPageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import { ApiKeyService } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement, useMemo } from 'react';
import { getColumns } from './components/ApiKeysTable/columns';
import { DataTable } from './components/ApiKeysTable/data-table';

export default function ApiKeys(): ReactElement<any> {
  return (
    <div className="flex flex-col gap-5">
      <SubPageHeader
        header="API Keys"
        description="Create and manage API keys"
        extraHeading={<NewApiKeyButton />}
      />
      <div>
        <ApiKeyTable />
      </div>
    </div>
  );
}

interface ApiKeyTableProps {}
function ApiKeyTable(props: ApiKeyTableProps): ReactElement<any> {
  const {} = props;
  const { account } = useAccount();

  const { data, isLoading, refetch } = useQuery(
    ApiKeyService.method.getAccountApiKeys,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );

  const columns = useMemo(
    () =>
      getColumns({
        accountName: account?.name ?? '',
        onDeleted() {
          refetch();
        },
      }),
    [account?.name ?? '']
  );

  if (isLoading) {
    return <SkeletonTable />;
  }

  const apiKeys = data?.apiKeys ?? [];

  return (
    <div>
      <DataTable columns={columns} data={apiKeys} />
    </div>
  );
}

interface NewApiKeyButtonProps {}

function NewApiKeyButton(props: NewApiKeyButtonProps): ReactElement<any> {
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
