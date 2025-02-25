'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import EmptyState, { EmptyStateLinkButton } from '@/components/EmptyState';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useQuery } from '@connectrpc/connect-query';
import { TransformersService } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReadonlyURLSearchParams, useSearchParams } from 'next/navigation';
import { ReactElement, useMemo } from 'react';
import { IoMdCode } from 'react-icons/io';
import { getSystemTransformerColumns } from './components/SystemTransformersTable/columns';
import { SystemTransformersDataTable } from './components/SystemTransformersTable/data-table';
import { getUserDefinedTransformerColumns } from './components/UserDefinedTransformersTable/columns';
import { UserDefinedTransformersDataTable } from './components/UserDefinedTransformersTable/data-table';

export default function Transformers(): ReactElement<any> {
  const searchParams = useSearchParams();
  const defaultTab = getTableTabFromParams(searchParams);
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Transformers"
          extraHeading={<NewTransformerButton />}
        />
      }
      containerClassName="transformer-page"
    >
      <TransformersTable defaultTab={defaultTab} />
    </OverviewContainer>
  );
}

function getTableTabFromParams(
  searchParams: ReadonlyURLSearchParams
): TableTab {
  const tab = searchParams.get('tab');
  return tab && isTableTab(tab) ? tab : 'ud';
}

function isTableTab(input: string): input is TableTab {
  return input === 'ud' || input === 'system';
}

type TableTab = 'ud' | 'system';

interface TransformersTableProps {
  defaultTab: TableTab;
}

function TransformersTable(props: TransformersTableProps): ReactElement<any> {
  const { defaultTab } = props;
  const { data, isLoading: isSystemTransformersLoading } = useQuery(
    TransformersService.method.getSystemTransformers
  );
  const { account } = useAccount();
  const {
    data: udTransformers,
    isLoading: userDefinedTransformersLoading,
    refetch: userDefinedTransformerRefetch,
  } = useQuery(
    TransformersService.method.getUserDefinedTransformers,
    { accountId: account?.id ?? '' },
    { enabled: !!account?.id }
  );

  const systemTransformers = data?.transformers ?? [];
  const userDefinedTransformers = udTransformers?.transformers ?? [];

  // memoizing these columns to prevent infinite re-render when hovering over next link
  const systemTransformerColumns = useMemo(
    () =>
      getSystemTransformerColumns({
        accountName: account?.name ?? '',
      }),
    [account?.name]
  );
  // memoizing these columns to prevent infinite re-render when hovering over next link
  const userDefinedTransformerColumns = useMemo(
    () =>
      getUserDefinedTransformerColumns({
        onTransformerDeleted() {
          userDefinedTransformerRefetch();
        },
        accountName: account?.name ?? '',
      }),
    [account?.name]
  );

  if (isSystemTransformersLoading || userDefinedTransformersLoading) {
    return <SkeletonTable />;
  }

  return (
    <div>
      <Tabs defaultValue={defaultTab}>
        <TabsList>
          <TabsTrigger value="ud">User Defined Transformers</TabsTrigger>
          <TabsTrigger value="system">System Transformers</TabsTrigger>
        </TabsList>
        <TabsContent value="ud">
          {userDefinedTransformers.length == 0 ? (
            <EmptyState
              title="No User Defined Transformers yet"
              description="Create a User Defined Transformer to implement data transformation logic. "
              icon={<IoMdCode className="w-8 h-8 text-primary" />}
              extra={
                <EmptyStateLinkButton
                  buttonText="Create your first Transformer"
                  href={`/${account?.name}/new/transformer`}
                />
              }
            />
          ) : (
            <UserDefinedTransformersDataTable
              columns={userDefinedTransformerColumns}
              data={userDefinedTransformers}
            />
          )}
        </TabsContent>
        <TabsContent value="system">
          <SystemTransformersDataTable
            columns={systemTransformerColumns}
            data={systemTransformers}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function NewTransformerButton(): ReactElement<any> {
  const { account } = useAccount();
  return (
    <NextLink href={`/${account?.name}/new/transformer`}>
      <Button>
        <ButtonText leftIcon={<PlusIcon />} text="New Transformer" />
      </Button>
    </NextLink>
  );
}
