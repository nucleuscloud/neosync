'use client';
import SqlServerConnectionForm from '@/components/connections/forms/sql-server/SqlServerConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiMysql } from 'react-icons/di';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';
export default function Mssql() {
  const onSuccess = useOnCreateSuccess();
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
      <SqlServerConnectionForm onSuccess={onSuccess} mode="create" />
    </OverviewContainer>
  );
}
