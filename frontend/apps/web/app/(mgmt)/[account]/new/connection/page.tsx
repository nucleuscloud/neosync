'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';
import ConnectionCard, { ConnectionMeta } from './components/ConnectionCard';

const CONNECTIONS: ConnectionMeta[] = [
  {
    urlSlug: 'postgres',
    name: 'Postgres',
    description:
      'PostgreSQL is a free and open-source relational database manageent system emphasizing extensibility and SQL compliance.',
    connectionType: 'postgres',
  },
  {
    urlSlug: 'mysql',
    name: 'MySQL',
    description:
      'MySQL is an open-source relational database management system.',
    connectionType: 'mysql',
  },
  {
    urlSlug: 'aws-s3',
    name: 'AWS S3',
    description:
      'Amazon Simple Storage Service (Amazon S3) is an object storage service used to store and retrieve any data.',
    connectionType: 'aws-s3',
  },
  {
    urlSlug: 'gcp-cloud-storage',
    name: 'GCP Cloud Storage',
    description:
      'GCP Cloud Storage is an object storage service used to store and retrieve any data.',
    connectionType: 'gcp-cloud-storage',
  },
  {
    urlSlug: 'neon',
    name: 'Neon',
    description:
      'Neon is a serverless Postgres database that separates storage and copmuyte to offer autoscaling, branching and bottomless storage.',
    connectionType: 'postgres',
    connectionTypeVariant: 'neon',
  },
  {
    urlSlug: 'supabase',
    name: 'Supabase',
    description:
      'Supabase is an open source Firebase alternative that ships with Authentication, Instant APIs, Edge functions and more.',
    connectionType: 'postgres',
    connectionTypeVariant: 'supabase',
  },
  {
    urlSlug: 'openai',
    name: 'OpenAI',
    description:
      'OpenAI (or equivalent interface) Chat API for generating synthetic data and inserting it directly into a destination datasource.',
    connectionType: 'openai',
  },
  {
    urlSlug: 'mongodb',
    name: 'MongoDB',
    description:
      'MongoDB is a source-available, cross-platform, document-oriented database program.',
    connectionType: 'mongodb',
    isExperimental: true,
  },
];

export default function NewConnectionPage(): ReactElement {
  const searchParams = useSearchParams();
  const { data: sytemAppConfigData } = useGetSystemAppConfig();
  const connectionTypes = new Set(searchParams.getAll('connectionType'));

  const connections = getConnectionsMetadata(
    connectionTypes,
    sytemAppConfigData?.isGcpCloudStorageConnectionsEnabled ?? false
  );
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

function getConnectionsMetadata(
  connectionTypes: Set<string>,
  isGcpCloudStorageConnectionsEnabled: boolean
): ConnectionMeta[] {
  let connections = CONNECTIONS;
  if (!isGcpCloudStorageConnectionsEnabled) {
    connections = connections.filter(
      (c) => c.connectionType !== 'gcp-cloud-storage'
    );
  }
  if (connectionTypes.size > 0) {
    connections = connections.filter((c) =>
      connectionTypes.has(c.connectionType)
    );
  }
  return connections;
}
