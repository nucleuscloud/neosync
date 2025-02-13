import { getConnectionUrlSlugName } from '@/app/(mgmt)/[account]/connections/util';
import { ConnectionConfig } from '@neosync/sdk';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { GrClone } from 'react-icons/gr';
import ButtonText from './ButtonText';
import { useAccount } from './providers/account-provider';
import { Button } from './ui/button';

interface CloneConnectionProps {
  connectionConfig: ConnectionConfig;
  id: string;
}

export function CloneConnectionButton(
  props: CloneConnectionProps
): ReactElement {
  const { connectionConfig, id } = props;
  const { account } = useAccount();

  return (
    <NextLink
      href={`/${account?.name}/new/connection/${getConnectionUrlSlugName(connectionConfig)}?sourceId=${id}`}
    >
      <Button>
        <ButtonText text="Clone" leftIcon={<GrClone className="mr-1" />} />
      </Button>
    </NextLink>
  );
}
