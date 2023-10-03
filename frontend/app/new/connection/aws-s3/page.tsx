import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import AwsS3Form from './AwsS3Form';

export default async function Postgres() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="AWS S3"
          description="Configure an AWS S3 data connection"
        />
      }
    >
      <AwsS3Form />
    </OverviewContainer>
  );
}
