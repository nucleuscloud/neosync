'use client';
import GcpCloudStorageConnectionForm from '@/components/connections/forms/gcp-cloud-storage/GcpCloudStorageConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Connection } from '@neosync/sdk';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement } from 'react';
import { SiGooglecloud } from 'react-icons/si';
export default function NewGCPCloudStoragePage(): ReactElement {
  const onSuccess = useOnSuccess();

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

function useOnSuccess(): (conn: Connection) => Promise<void> {
  const router = useRouter();
  const { account } = useAccount();
  // const posthog = usePostHog();
  const searchParams = useSearchParams();
  const returnTo = searchParams.get('returnTo');

  return async (conn: Connection): Promise<void> => {
    if (!account) {
      return;
    }
    if (returnTo) {
      router.push(returnTo);
    } else if (conn.id) {
      router.push(`/${account.name}/connections/${conn.id}`);
    }

    // try {
    //   // toast.success('Successfully created OpenAI Connection!');
    //   // posthog.capture('New Connection Created', { type: 'openai' });

    //   if (returnTo) {
    //     router.push(returnTo);
    //   } else if (conn.id) {
    //     router.push(`/${account.name}/connections/${conn.id}`);
    //   }
    // } catch (err) {
    //   console.error(err);
    //   toast.error('Unable to create OpenAI Connection', {
    //     description: getErrorMessage(err),
    //   });
    // }
  };
}
