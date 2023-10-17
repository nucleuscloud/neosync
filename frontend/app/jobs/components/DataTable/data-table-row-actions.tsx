'use client';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Job } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { Row } from '@tanstack/react-table';
import { useRouter } from 'next/navigation';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
}: DataTableRowActionsProps<TData>) {
  const job = row.original as Job;
  const router = useRouter();

  const { toast } = useToast();

  async function onDelete(): Promise<void> {
    try {
      await removeJob(job.id);
      toast({
        title: 'Job removed successfully!',
      });
      onDeleted();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove job',
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DropdownMenu
      modal={false} // needed because otherwise this breaks after a single use in conjunction with the delete dialog
    >
      <DropdownMenuTrigger className="hover:bg-gray-100 py-1 px-2 rounded-lg">
        <DotsHorizontalIcon className="h-4 w-4" />
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() => router.push(`/jobs/${job.id}`)}
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
          headerText="Are you sure you want to delete this job?"
          description="Deleting this job will also delete all job runs."
          onConfirm={() => onDelete()}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

async function removeJob(jobId: string): Promise<void> {
  const res = await fetch(`/api/jobs/${jobId}`, {
    method: 'DELETE',
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
