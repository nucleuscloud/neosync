'use client';
import NewOpenAiForm from '@/components/connections/forms/openai/NewOpenAiForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { ConnectionService, CreateConnectionRequest } from '@neosync/sdk';
import { useTheme } from 'next-themes';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useMemo } from 'react';
import { toast } from 'sonner';
import { OpenAiLogo } from './OpenAiLogo';

export default function OpenAi(): ReactElement {
  const { resolvedTheme } = useTheme();
  const logoBg = useMemo(
    () => (resolvedTheme === 'dark' ? 'white' : '#272F30'),
    [resolvedTheme]
  );

  const onSubmit = useOnSubmit();

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
      <NewOpenAiForm onSubmit={onSubmit} />
    </OverviewContainer>
  );
}

function useOnSubmit(): (values: CreateConnectionRequest) => Promise<void> {
  const { mutateAsync: createConnection } = useMutation(
    ConnectionService.method.createConnection
  );
  const router = useRouter();
  const { account } = useAccount();
  const posthog = usePostHog();
  const searchParams = useSearchParams();
  const returnTo = searchParams.get('returnTo');

  return async (values: CreateConnectionRequest): Promise<void> => {
    if (!account) {
      return;
    }

    try {
      const connectionResp = await createConnection(values);
      toast.success('Successfully created OpenAI Connection!');
      posthog.capture('New Connection Created', { type: 'openai' });

      if (returnTo) {
        router.push(returnTo);
      } else if (connectionResp.connection?.id) {
        router.push(
          `/${account.name}/connections/${connectionResp.connection.id}`
        );
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to create OpenAI Connection', {
        description: getErrorMessage(err),
      });
    }
  };
}
