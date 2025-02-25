'use client';
import MysqlConnectionForm from '@/components/connections/forms/mysql/MysqlConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { DiMysql } from 'react-icons/di';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';
export default function NewMysqlConnectionPage(): ReactElement {
  const onSuccess = useOnCreateSuccess();

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
      <MysqlConnectionForm mode="create" onSuccess={onSuccess} />
    </OverviewContainer>
  );
}
