import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SupabaseForm from './SupabaseForm';
import { SupabaseLogo } from './SupabaseLogo';

export default async function Supabase() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Supabase"
          subHeadings="Configure a Supabase database as a connection"
          leftIcon={<SupabaseLogo />}
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <SupabaseForm />
    </OverviewContainer>
  );
}
