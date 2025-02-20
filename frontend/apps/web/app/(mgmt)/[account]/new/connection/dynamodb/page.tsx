'use client';
import DynamoDbConnectionForm from '@/components/connections/forms/dynamodb/DynamoDbConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { FaAws } from 'react-icons/fa';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';

export default function NewDynamoDBConnection(): ReactElement {
  const onSuccess = useOnCreateSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="DynamoDB"
          subHeadings="Configure DynamoDB as a connection"
          leftIcon={<FaAws className=" w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <DynamoDbConnectionForm mode="create" onSuccess={onSuccess} />
    </OverviewContainer>
  );
}
