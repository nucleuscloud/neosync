import PostgresConnectionForm from '@/components/connections/forms/postgres/PostgresConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { DiPostgresql } from 'react-icons/di';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';

export default function NewPostgresPage(): ReactElement {
  const onSuccess = useOnCreateSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="PostgreSQL"
          subHeadings="Configure a PostgreSQL database as a connection"
          leftIcon={<DiPostgresql className=" w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <PostgresConnectionForm onSuccess={onSuccess} mode="create" />
    </OverviewContainer>
  );
}
