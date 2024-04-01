import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import ConnectionCard, { ConnectionMeta } from './components/ConnectionCard';

const CONNECTIONS: ConnectionMeta[] = [
  {
    urlSlug: 'postgres',
    name: 'Postgres',
    description:
      'PostgreSQL is a free and open-source relational database manageent system emphasizing extensibility and SQL compliance.',
  },
  {
    urlSlug: 'mysql',
    name: 'MySQL',
    description:
      'MySQL is an open-source relational database management system.',
  },
  {
    urlSlug: 'aws-s3',
    name: 'AWS S3',
    description:
      'Amazon Simple Storage Service (Amazon S3) is an object storage service used to store and retrieve any data.',
  },
  {
    urlSlug: 'neon',
    name: 'Neon',
    description:
      'Neon is a serverless Postgres database that separates storage and copmuyte to offer autoscaling, branching and bottomless storage.',
  },
  {
    urlSlug: 'supabase',
    name: 'Supabase',
    description:
      'Supabase is an open source Firebase alternative that ships with Authentication, Instant APIs, Edge functions and more. ',
  },
];

export default function NewConnection(): ReactElement {
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
        {CONNECTIONS.map((connection) => (
          <ConnectionCard key={connection.urlSlug} connection={connection} />
        ))}
      </div>
    </OverviewContainer>
  );
}
