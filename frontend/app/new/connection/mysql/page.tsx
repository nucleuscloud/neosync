import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import MysqlForm from './MysqlForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Mysql"
          description="Configure a Mysql data connection"
          pageHeaderContainerClassName="mx-64"
        />
      }
    >
      <MysqlForm />
    </OverviewContainer>
  );
}
