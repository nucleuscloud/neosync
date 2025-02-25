'use client';
import GcpCloudStorageConnectionForm from '@/components/connections/forms/gcp-cloud-storage/GcpCloudStorageConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { SiGooglecloud } from 'react-icons/si';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';

export default function NewGCPCloudStoragePage(): ReactElement<any> {
  const onSuccess = useOnCreateSuccess();
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
      <GcpCloudStorageConnectionForm mode="create" onSuccess={onSuccess} />
    </OverviewContainer>
  );
}
