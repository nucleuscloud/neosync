import ButtonText from '@/components/ButtonText';
import { CloneConnectionButton } from '@/components/CloneConnectionButton';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { useQuery } from '@connectrpc/connect-query';
import {
  Connection,
  HasPermissionRequest_Permission,
  UserAccountService,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import RemoveConnectionButton from './RemoveConnectionButton';

interface Props {
  connection: Connection;
}

export default function ViewActions(props: Props): ReactElement | null {
  const { connection } = props;
  const router = useRouter();
  const { account } = useAccount();

  const { data: canEditConnections } = useQuery(
    UserAccountService.method.hasPermission,
    {
      accountId: account?.id ?? '',
      resourceId: connection?.id ?? '',
      permission: HasPermissionRequest_Permission.UPDATE,
    },
    { enabled: !!account?.id && !!connection?.id }
  );

  const { data: canDeleteConnections } = useQuery(
    UserAccountService.method.hasPermission,
    {
      accountId: account?.id ?? '',
      resourceId: connection?.id ?? '',
      permission: HasPermissionRequest_Permission.DELETE,
    },
    { enabled: !!account?.id && !!connection?.id }
  );

  const { data: canCreateConnections } = useQuery(
    UserAccountService.method.hasPermission,
    {
      accountId: account?.id ?? '',
      resourceId: connection?.id ?? '',
      permission: HasPermissionRequest_Permission.CREATE,
    },
    { enabled: !!account?.id && !!connection?.id }
  );

  if (!connection?.id) {
    return null;
  }

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
          <ButtonText text="Edit" />
        </Button>
      )}
    </div>
  );
}
