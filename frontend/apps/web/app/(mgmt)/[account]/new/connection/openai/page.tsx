'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useTheme } from 'next-themes';
import OpenAiForm from './OpenAiForm';
import { OpenAiLogo } from './OpenAiLogo';

export default function OpenAi() {
  const { resolvedTheme } = useTheme();
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="OpenAI"
          subHeadings="Configure an OpenAI SDK to be used for AI-focused data generation jobs."
          leftIcon={
            <OpenAiLogo bg={resolvedTheme === 'dark' ? 'white' : '#272F30'} />
          }
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <OpenAiForm />
    </OverviewContainer>
  );
}
