'use client';

import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { AccountApiKey } from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { useRouter } from 'next/navigation';
import RemoveAccountApiKeyButton from '../../[id]/components/RemoveAccountApiKeyButton';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
}: DataTableRowActionsProps<TData>) {
  const apikey = row.original as AccountApiKey;
  const router = useRouter();

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
          onClick={() => router.push(`/settings/account-api-keys/${apikey.id}`)}
        >
          View
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <RemoveAccountApiKeyButton
          id={apikey.id}
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()}
            >
              Delete
            </DropdownMenuItem>
          }
          onDeleted={() => onDeleted()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

async function removeApiKey(keyId: string): Promise<void> {
  const res = await fetch(`/api/api-keys/account/${keyId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
