'use client';
import OpenAiConnectionForm from '@/components/connections/forms/openai/OpenAiConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Connection } from '@neosync/sdk';
import { useTheme } from 'next-themes';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useMemo } from 'react';
import { OpenAiLogo } from './OpenAiLogo';

export default function OpenAi(): ReactElement {
  const { resolvedTheme } = useTheme();
  const logoBg = useMemo(
    () => (resolvedTheme === 'dark' ? 'white' : '#272F30'),
    [resolvedTheme]
  );

  const onSuccess = useOnSuccess();

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="OpenAI"
          subHeadings="Configure an OpenAI SDK to be used for AI-focused data generation jobs."
          leftIcon={<OpenAiLogo bg={logoBg} />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <OpenAiConnectionForm mode="create" onSuccess={onSuccess} />
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
