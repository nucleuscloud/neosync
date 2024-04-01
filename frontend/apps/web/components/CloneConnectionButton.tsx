import { getConnectionType } from '@/app/(mgmt)/[account]/connections/components/ConnectionsTable/data-table-row-actions';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import ButtonText from './ButtonText';
import { useAccount } from './providers/account-provider';
import { Button } from './ui/button';

interface CloneConnectionProps {
  connectionType: string;
  id: string;
}

export function CloneConnectionButton(
  props: CloneConnectionProps
): ReactElement {
  const { connectionType, id } = props;
  const { account } = useAccount();

  return (
    <NextLink
      href={`/${account?.name}/new/connection/${getConnectionType(connectionType)}?sourceId=${id}`}
    >
      <Button>
        <ButtonText
          text="Clone Connection"
          leftIcon={<GrClone className="mr-1" />}
        />
      </Button>
    </NextLink>
  );
}
