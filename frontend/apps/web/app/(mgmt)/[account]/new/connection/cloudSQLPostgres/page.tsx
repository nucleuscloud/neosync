import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { CloudSQLLogo } from './CloudSQLLogo';
import ClouDSQLPostgresForm from './CloudSQLPostgresForm';

export default async function CloudSQLPostgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="CloudSQL (Postgres)"
          subHeadings="Configure a CloudSQL (Postgres) database as a connection"
          leftIcon={<CloudSQLLogo />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <ClouDSQLPostgresForm />
    </OverviewContainer>
  );
}
