import { isAuthEnabled } from '@/api-only/auth-config';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import MemberManagementSettings from './components/MemberManagementSettings';
import SubNav, { ITEMS } from './temporal/components/SubNav';

export default function Settings() {
  return (
    <OverviewContainer
      Header={<PageHeader header="Settings" />}
      containerClassName="settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        {isAuthEnabled() && <MemberManagementSettings />}
      </div>
    </OverviewContainer>
  );
}
