'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { useGetUserDefinedTransformers } from '@/libs/hooks/useGetUserDefinedTransformers';
import { PlusIcon } from '@radix-ui/react-icons';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { getSystemTransformerColumns } from './components/SystemTransformersTable/columns';
import { SystemTransformersDataTable } from './components/SystemTransformersTable/data-table';
import { getUserDefinedTransformerColumns } from './components/UserDefinedTransformersTable/columns';
import { UserDefinedTransformersDataTable } from './components/UserDefinedTransformersTable/data-table';

export default function Transformers(): ReactElement {
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
      <TransformersTable />
    </OverviewContainer>
  );
}

function TransformersTable(): ReactElement {
  const { data, isLoading: transformersIsLoading } = useGetSystemTransformers();
  const { account } = useAccount();
  const {
    data: udTransformers,
    isLoading: userDefinedTransformersLoading,
    mutate: userDefinedTransformerMutate,
  } = useGetUserDefinedTransformers(account?.id ?? '');

  const systemTransformers = data?.transformers ?? [];
  const userDefinedTransformers = udTransformers?.transformers ?? [];

  if (transformersIsLoading || userDefinedTransformersLoading) {
    return <SkeletonTable />;
  }

  const systemTransformerColumns = getSystemTransformerColumns();
  const userDefinedTransformerColumns = getUserDefinedTransformerColumns({
    onTransformerDeleted() {
      userDefinedTransformerMutate();
    },
  });

  return (
    <div>
      <Tabs defaultValue="udtransformers" className="">
        <TabsList>
          <TabsTrigger value="udtransformers">
            User Defined Transformers
          </TabsTrigger>
          <TabsTrigger value="system">System Transformers</TabsTrigger>
        </TabsList>
        <TabsContent value="udtransformers">
          <UserDefinedTransformersDataTable
            columns={userDefinedTransformerColumns}
            data={userDefinedTransformers}
          />
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

function NewTransformerButton(): ReactElement {
  const { account } = useAccount();
  return (
    <NextLink href={`/${account?.name}/new/transformer`}>
      <Button>
        <ButtonText leftIcon={<PlusIcon />} text="New Transformer" />
      </Button>
    </NextLink>
  );
}
