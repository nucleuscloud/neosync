'use client';
import MongoDbConnectionForm from '@/components/connections/forms/mongodb/MongoDbConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { DiMongodb } from 'react-icons/di';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';
export default function NewMongoDBConnectionPage(): ReactElement<any> {
  const onSuccess = useOnCreateSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="MongoDB"
          subHeadings="Configure a MongoDB database as a connection"
          leftIcon={<DiMongodb className=" w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <MongoDbConnectionForm mode="create" onSuccess={onSuccess} />
    </OverviewContainer>
  );
}
