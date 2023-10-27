'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { getColumns } from './components/TransformersTable/columns';
import { DataTable } from './components/TransformersTable/data-table';

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

  const transformers = data?.transformers ?? [];

  if (transformersIsLoading) {
    return <SkeletonTable />;
  }

  const columns = getColumns();

  return (
    <div>
      <DataTable columns={columns} data={transformers} />
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
