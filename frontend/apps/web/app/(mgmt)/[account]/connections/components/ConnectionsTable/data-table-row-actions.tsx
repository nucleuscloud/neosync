'use client';

import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { Connection, ConnectionConfig } from '@neosync/sdk';
import { deleteConnection } from '@neosync/sdk/connectquery';
import { useRouter } from 'next/navigation';
import { getConnectionType } from '../../util';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
}: DataTableRowActionsProps<TData>) {
  const connection = row.original as Connection;
  const router = useRouter();
  const { account } = useAccount();
  const { toast } = useToast();
  const { mutateAsync } = useMutation(deleteConnection);

  async function onDelete(): Promise<void> {
    try {
      await mutateAsync({ id: connection.id });
      toast({
        title: 'Connection removed successfully!',
        variant: 'success',
      });
      onDeleted();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove connection',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
        >
          <DotsHorizontalIcon className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[160px]">
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() =>
            router.push(`/${account?.name}/connections/${connection.id}`)
          }
        >
          View
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() =>
            router.push(
              `/${account?.name}/new/connection/${getConnectionType(connection.connectionConfig ?? new ConnectionConfig())}?sourceId=${connection.id}`
            )
          }
        >
          Clone
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DeleteConfirmationDialog
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()}
            >
              Delete
            </DropdownMenuItem>
          }
          headerText="Are you sure you want to delete this connection?"
          description="Deleting this connection is irreversable!"
          onConfirm={async () => onDelete()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
