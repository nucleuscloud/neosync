'use client';
import AwsS3ConnectionForm from '@/components/connections/forms/s3/AwsS3ConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import { FaAws } from 'react-icons/fa';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';

export default function NewAwsS3ConnectionPage(): ReactElement<any> {
  const onSuccess = useOnCreateSuccess();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="AWS S3"
          subHeadings="Configure an AWS S3 bucket as a connection"
          leftIcon={<FaAws className="w-[40px] h-[40px]" />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <AwsS3ConnectionForm mode="create" onSuccess={onSuccess} />
    </OverviewContainer>
  );
}
