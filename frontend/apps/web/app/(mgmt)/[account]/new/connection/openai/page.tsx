'use client';
import OpenAiConnectionForm from '@/components/connections/forms/openai/OpenAiConnectionForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useTheme } from 'next-themes';
import { ReactElement, useMemo } from 'react';
import { useOnCreateSuccess } from '../components/useOnCreateSuccess';
import { OpenAiLogo } from './OpenAiLogo';

export default function OpenAi(): ReactElement {
  const { resolvedTheme } = useTheme();
  const logoBg = useMemo(
    () => (resolvedTheme === 'dark' ? 'white' : '#272F30'),
    [resolvedTheme]
  );

  const onSuccess = useOnCreateSuccess();
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
