import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiMysql } from 'react-icons/di';
import MysqlForm from './MysqlForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Mysql"
          description="Configure a Mysql database as a connection"
          pageHeaderContainerClassName="mx-64"
          leftIcon={<DiMysql className="w-[40px] h-[40px]" />}
        />
      }
    >
      <MysqlForm />
    </OverviewContainer>
  );
}
