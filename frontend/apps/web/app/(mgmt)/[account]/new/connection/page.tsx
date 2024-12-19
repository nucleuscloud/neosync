'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';
import { useGetConnectionsMetadata } from '../../connections/util';
import ConnectionCard from './components/ConnectionCard';

export default function NewConnectionPage(): ReactElement {
  const searchParams = useSearchParams();
  const connectionTypes = new Set(searchParams.getAll('connectionType'));

  const connections = useGetConnectionsMetadata(connectionTypes);
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Connection"
          subHeadings="Connect a new datasource to use in sync or generate jobs."
          pageHeaderContainerClassName="mx-24"
        />
      }
    >
      <div className="gap-6 rounded-lg md:grid lg:grid-cols-2 xl:grid-cols-3 content-stretch mx-24">
        {connections.map((connection) => (
          <ConnectionCard key={connection.urlSlug} connection={connection} />
        ))}
      </div>
    </OverviewContainer>
  );
}
