'use client';
import PostgresConnectionForm from '@/components/connections/forms/postgres/PostgresConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';
import { NeonLogo } from './NeonLogo';

export default function NewPostgresNeonPage(): ReactElement<any> {
  const onSuccess = useOnCreateSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Neon"
          subHeadings="Configure a Neon database as a connection"
          leftIcon={<NeonLogo />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <PostgresConnectionForm onSuccess={onSuccess} mode="create" />
    </OverviewContainer>
  );
}
