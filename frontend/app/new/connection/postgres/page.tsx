import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { withPageAuthRequired } from '@auth0/nextjs-auth0';
import PostgresForm from './PostgresForm';

export default withPageAuthRequired(async function Postgres() {
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
});
