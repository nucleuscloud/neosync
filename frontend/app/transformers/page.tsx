'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useGetCustomTransformers } from '@/libs/hooks/useGetCustomTransformers';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { getCustomTransformerColumns } from './components/CustomTransformersTable/columns';
import { CustomTransformersDataTable } from './components/CustomTransformersTable/data-table';
import { getSystemTransformerColumns } from './components/SystemTransformersTable/columns';
import { SystemTransformersDataTable } from './components/SystemTransformersTable/data-table';

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
    data: cTransformers,
    isLoading: customTransformersLoading,
    mutate: customTransformerMutate,
  } = useGetCustomTransformers(account?.id ?? '');

  const systemTransformers = data?.transformers ?? [];
  const customTransformers = cTransformers?.transformers ?? [];

  if (transformersIsLoading || customTransformersLoading) {
    return <SkeletonTable />;
  }

  const systemTransformerColumns = getSystemTransformerColumns();
  const customTransformerColumns = getCustomTransformerColumns({
    onTransformerDeleted() {
      customTransformerMutate();
    },
  });

  return (
    <div>
      <Tabs defaultValue="custom" className="">
        <TabsList>
          <TabsTrigger value="custom">Custom Transformers</TabsTrigger>
          <TabsTrigger value="system">System Transformers</TabsTrigger>
        </TabsList>
        <TabsContent value="custom">
          <CustomTransformersDataTable
            columns={customTransformerColumns}
            data={customTransformers}
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
  return (
    <NextLink href={'/new/transformer'}>
      <Button> + New Transformer</Button>
    </NextLink>
  );
}
