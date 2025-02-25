import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { ReactElement } from 'react';
import NewApiKeyForm from './NewApiKeyForm';

export default function NewAccountApiKey(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Create a new account API Key"
          subHeadings="API Keys can be used to interact with Neosync programmatically"
        />
      }
      containerClassName="mx-24"
    >
      <NewApiKeyForm />
    </OverviewContainer>
  );
}
