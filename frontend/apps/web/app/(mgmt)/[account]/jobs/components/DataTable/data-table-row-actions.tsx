'use client';

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Job, JobService } from '@neosync/sdk';
import { Row } from '@tanstack/react-table';
import { useRouter } from 'next/navigation';

import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { DotsHorizontalIcon } from '@radix-ui/react-icons';
import { nanoid } from 'nanoid';
import { toast } from 'sonner';
import { getJobCloneUrlFromJob } from '../../[id]/components/JobCloneButton';
import { setDefaultNewJobFormValues } from '../../util';

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
  const { account } = useAccount();
  const { mutateAsync: removeJob } = useMutation(JobService.method.deleteJob);

  async function onDelete(): Promise<void> {
    try {
      await removeJob({ id: job.id });
      toast.success('Job removed successfully!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job', {
        description: getErrorMessage(err),
      });
    }
  }

  function onCloneClick(): void {
    if (!account) {
      return;
    }
    const sessionToken = nanoid();
    setDefaultNewJobFormValues(window.sessionStorage, job, sessionToken);
    router.push(getJobCloneUrlFromJob(account, job, sessionToken));
  }

  return (
    <DropdownMenu
      modal={false} // needed because otherwise this breaks after a single use in conjunction with the delete dialog
    >
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="flex h-8 w-8 p-0 data-[state=open]:bg-muted"
        >
          <DotsHorizontalIcon className="h-4 w-4" />
          <span className="sr-only">Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() => router.push(`/${account?.name}/jobs/${job.id}`)}
        >
          View
        </DropdownMenuItem>
        <DropdownMenuItem
          className="cursor-pointer"
          onClick={() => onCloneClick()}
        >
          Clone
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
