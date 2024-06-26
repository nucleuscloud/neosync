import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { SiGooglecloud } from 'react-icons/si';
import GcpCloudStorageForm from './GcpCloudStorageForm';

export default async function GCPCloudStoragePage() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="GCP Cloud Storage"
          subHeadings="Configure a GCP Cloud Storage bucket as a connection"
          leftIcon={<SiGooglecloud className="w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <GcpCloudStorageForm />
    </OverviewContainer>
  );
}
