import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SubNav, { ITEMS } from './temporal/components/SubNav';

export default function Settings() {
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page mx-24"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        <p>There is nothing here yet...</p>
      </div>
    </OverviewContainer>
  );
}
