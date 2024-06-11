import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { DiMongodb } from 'react-icons/di';
import MongoDBForm from './MongoDBForm';

export default async function Postgres() {
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
      <MongoDBForm />
    </OverviewContainer>
  );
}
