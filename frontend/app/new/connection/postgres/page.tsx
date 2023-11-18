import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiPostgresql } from 'react-icons/di';
import PostgresForm from './PostgresForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="PostgreSQL"
          description="Configure a PostgreSQL database as a connection"
          leftIcon={<DiPostgresql className="w-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <PostgresForm />
    </OverviewContainer>
  );
}
