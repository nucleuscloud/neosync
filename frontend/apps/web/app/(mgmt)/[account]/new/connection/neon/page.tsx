import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import NeonForm from './NeonForm';
import { NeonLogo } from './NeonLogo';

export default async function Neon() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Neon"
          subHeadings="Configure a Neon database as a connection"
          leftIcon={<NeonLogo />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <NeonForm />
    </OverviewContainer>
  );
}
