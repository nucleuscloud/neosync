import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useGetAuthEnabled } from '@/libs/hooks/useGetAuthEnabled';
import MemberManagementSettings from './components/MemberManagementSettings';
import SubNav, { ITEMS } from './temporal/components/SubNav';

export default function Settings() {
  const authEnabled = useGetAuthEnabled();
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page mx-24"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        {authEnabled && <MemberManagementSettings />}
      </div>
    </OverviewContainer>
  );
}
