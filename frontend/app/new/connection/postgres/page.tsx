import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import PostgresForm from './PostgresForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="PostgreSQL"
          description="Configure a PostgreSQL data connection"
        />
      }
    >
      <PostgresForm />
    </OverviewContainer>
  );
}
