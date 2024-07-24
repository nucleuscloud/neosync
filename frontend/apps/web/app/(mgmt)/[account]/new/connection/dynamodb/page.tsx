import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiAws } from 'react-icons/di';
import DynamoDBForm from './DynamoDBForm';

export default async function DynamoDB() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="DynamoDB"
          subHeadings="Configure DynamoDB as a connection"
          leftIcon={<DiAws className=" w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <DynamoDBForm />
    </OverviewContainer>
  );
}
