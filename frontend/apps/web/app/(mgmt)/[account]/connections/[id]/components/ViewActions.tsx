import ButtonText from '@/components/ButtonText';
import { CloneConnectionButton } from '@/components/CloneConnectionButton';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import {
  Connection,
  ResourcePermission_Action,
  ResourcePermission_Type,
  UserAccountService,
} from '@neosync/sdk';
import { Pencil1Icon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import RemoveConnectionButton from './RemoveConnectionButton';

interface Props {
  connection: Connection;
}

export default function ViewActions(props: Props): ReactElement<any> | null {
  const { connection } = props;
  const router = useRouter();
  const { account } = useAccount();

  const { data: permissions } = useQuery(
    UserAccountService.method.hasPermissions,
    {
      accountId: account?.id ?? '',
      resources: [
        {
          type: ResourcePermission_Type.CONNECTION,
          id: connection?.id ?? '',
          action: ResourcePermission_Action.UPDATE,
        },
        {
          type: ResourcePermission_Type.CONNECTION,
          id: connection?.id ?? '',
          action: ResourcePermission_Action.DELETE,
        },
        {
          type: ResourcePermission_Type.CONNECTION,
          id: connection?.id ?? '',
          action: ResourcePermission_Action.CREATE,
        },
      ],
    }
  );

  if (!connection?.id) {
    return null;
  }

  const [canCreateConnections, canDeleteConnections, canEditConnections] =
    permissions?.assertions ?? [false, false, false];

  return (
    <div className="flex flex-row items-center gap-4">
      {canCreateConnections && <CloneConnectionButton id={connection.id} />}
      {canDeleteConnections && (
        <RemoveConnectionButton connectionId={connection?.id ?? ''} />
      )}
      {canEditConnections && (
        <Button
          type="button"
          onClick={() => {
            router.push(`/${account?.name}/connections/${connection?.id}/edit`);
          }}
        >
          <ButtonText leftIcon={<Pencil1Icon />} text="Edit" />
        </Button>
      )}
    </div>
  );
}
