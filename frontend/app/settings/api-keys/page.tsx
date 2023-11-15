'use client';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { PlusIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';
import SubNav, { ITEMS } from '../temporal/components/SubNav';

export default function ApiKeys(): ReactElement {
  return (
    <OverviewContainer
      Header={
        <PageHeader header="API Keys" extraHeading={<NewApiKeyButton />} />
      }
      containerClassName="apikeys-settings-page"
    >
      <div className="flex flex-col gap-4">
        <div>
          <SubNav items={ITEMS} />
        </div>
        <ApiKeyTable />
      </div>
    </OverviewContainer>
  );
}

interface ApiKeyTableProps {}
function ApiKeyTable(props: ApiKeyTableProps): ReactElement {
  const {} = props;
  const { account } = useAccount();

  return <div>TODO Table</div>;
}

interface NewApiKeyButtonProps {}

function NewApiKeyButton(props: NewApiKeyButtonProps): ReactElement {
  const {} = props;
  return (
    <Link href={'/new/api-key'}>
      <Button type="button">
        <ButtonText leftIcon={<PlusIcon />} text="New API Key" />
      </Button>
    </Link>
  );
}
