'use client';

import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/components/ui/use-toast';
import { UserDefinedTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { useRouter } from 'next/navigation';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
}: DataTableRowActionsProps<TData>) {
  const transformer = row.original as UserDefinedTransformer;
  const router = useRouter();
  const { toast } = useToast();

  async function onDelete(): Promise<void> {
    try {
      await removeTransformer(transformer.id);
      toast({
        title: 'Transformer removed successfully!',
        variant: 'success',
      });
      onDeleted();
    } catch (err) {
      toast({
        title: 'Unable to remove tranformer',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <DropdownMenu>
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
          onClick={() => router.push(`/transformers/${transformer.id}`)}
        >
          View
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DeleteConfirmationDialog
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()} // needed for the delete modal to not automatically close
            >
              Delete
            </DropdownMenuItem>
          }
          headerText="Are you sure you want to delete this Transformer?"
          description="Deleting this Transformer may impact running Jobs. "
          onConfirm={() => onDelete()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

async function removeTransformer(transformerId: string): Promise<void> {
  const res = await fetch(
    `/api/transformers/user-defined?transformerId=${transformerId}`,
    {
      method: 'DELETE',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
