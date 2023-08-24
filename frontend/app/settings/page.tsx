import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';

export default function Settings() {
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Settings"
          description="Configure account-wide settings here."
        />
      }
      containerClassName="settings-page"
    >
      <div>
        <p>There is nothing here yet...</p>
      </div>
    </OverviewContainer>
  );
}
