import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiMysql } from 'react-icons/di';
import MysqlForm from './MysqlForm';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="MySQL"
          subHeadings="Configure a MySQL database as a connection"
          leftIcon={<DiMysql className="w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <MysqlForm />
    </OverviewContainer>
  );
}
