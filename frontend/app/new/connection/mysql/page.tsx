import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import MysqlForm from './MysqlForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Myql"
          description="Configure a Mysql data connection"
        />
      }
    >
      <MysqlForm />
    </OverviewContainer>
  );
}
