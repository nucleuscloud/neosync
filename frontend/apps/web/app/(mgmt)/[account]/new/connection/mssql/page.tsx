import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiMysql } from 'react-icons/di';
import MssqlForm from './MssqlForm';

export default async function Mssql() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Microsoft SQL Server"
          subHeadings="Configure a Microsoft SQL Server database as a connection"
          leftIcon={<DiMysql className="w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <MssqlForm />
    </OverviewContainer>
  );
}
