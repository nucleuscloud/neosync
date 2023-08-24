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
    description: 'Amazon S3',
  },
];

export default function NewConnection(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new Connection"
          description="Connection a new datasource to use in jobs or other synchronizations."
        />
      }
    >
      <div className="items-start justify-center gap-6 rounded-lg p-8 md:grid lg:grid-cols-2 xl:grid-cols-3">
        {CONNECTIONS.map((connection) => (
          <ConnectionCard key={connection.urlSlug} connection={connection} />
        ))}
      </div>
    </OverviewContainer>
  );
}
