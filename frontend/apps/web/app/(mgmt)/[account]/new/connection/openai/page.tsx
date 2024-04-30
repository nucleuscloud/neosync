import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { OpenAiLogo } from './OpenAiLogo';
import OpenAiForm from './OpenAiForm';

export default async function Supabase() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="OpenAI"
          subHeadings="Configure an OpenAI SDK to be used for AI-focused data generation jobs."
          leftIcon={<OpenAiLogo />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <OpenAiForm />
    </OverviewContainer>
  );
}
