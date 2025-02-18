'use client';
import OpenAiConnectionForm from '@/components/connections/forms/openai/OpenAiConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { useTheme } from 'next-themes';
import { useRouter } from 'next/navigation';
import { ReactElement, useMemo } from 'react';
import { OpenAiLogo } from '../../../new/connection/openai/OpenAiLogo';
export default function CloneConnectionPage({
  params,
}: PageProps): ReactElement {
  const { resolvedTheme } = useTheme();
  const logoBg = useMemo(
    () => (resolvedTheme === 'dark' ? 'white' : '#272F30'),
    [resolvedTheme]
  );

  const router = useRouter();
  const { account } = useAccount();

  const id = params?.id ?? ''; // todo: check this

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
      {/* todo: wrap in permission check */}
      <OpenAiConnectionForm
        mode="clone"
        onSuccess={async (conn) => {
          router.push(`/${account?.name}/connections/${conn.id}`);
        }}
        connectionId={id}
      />
    </OverviewContainer>
  );
}
