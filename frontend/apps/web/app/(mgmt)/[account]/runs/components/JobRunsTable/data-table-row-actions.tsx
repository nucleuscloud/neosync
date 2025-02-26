'use client';

import { Cross2Icon, DotsHorizontalIcon } from '@radix-ui/react-icons';
import { Row } from '@tanstack/react-table';

import ConfirmationDialog from '@/components/ConfirmationDialog';
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
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import {
  JobRun,
  JobRunStatus as JobRunStatusEnum,
  JobService,
} from '@neosync/sdk';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';

interface DataTableRowActionsProps<TData> {
  row: Row<TData>;
  onDeleted(): void;
}

export function DataTableRowActions<TData>({
  row,
  onDeleted,
}: DataTableRowActionsProps<TData>) {
  const run = row.original as JobRun;
  const router = useRouter();
  const { account } = useAccount();
  const { mutateAsync: removeJobRunAsync } = useMutation(
    JobService.method.deleteJobRun
  );
  const { mutateAsync: cancelJobRunAsync } = useMutation(
    JobService.method.cancelJobRun
  );
  const { mutateAsync: terminateJobRunAsync } = useMutation(
    JobService.method.terminateJobRun
  );

  async function onDelete(): Promise<void> {
    try {
      await removeJobRunAsync({ accountId: account?.id, jobRunId: run.id });
      toast.success('Removing Job Run. This may take a minute to delete!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onCancel(): Promise<void> {
    try {
      await cancelJobRunAsync({ accountId: account?.id, jobRunId: run.id });
      toast.success('Canceling Job Run. This may take a minute to cancel!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to cancel job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onTerminate(): Promise<void> {
    try {
      await terminateJobRunAsync({ accountId: account?.id, jobRunId: run.id });
      toast.success(
        'Terminating Job Run. This may take a minute to terminate!'
      );
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to terminate job run', {
        description: getErrorMessage(err),
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
          onClick={() => router.push(`/${account?.name}/runs/${run.id}`)}
        >
          View
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        {(run.status === JobRunStatusEnum.RUNNING ||
          run.status === JobRunStatusEnum.PENDING) && (
          <div>
            <ConfirmationDialog
              trigger={
                <DropdownMenuItem
                  className="cursor-pointer"
                  onSelect={(e) => e.preventDefault()}
                >
                  Cancel
                </DropdownMenuItem>
              }
              headerText="Cancel Job Run?"
              description="Are you sure you want to cancel this job run?"
              onConfirm={async () => onCancel()}
              buttonText="Cancel"
              buttonVariant="default"
              buttonIcon={<Cross2Icon />}
            />
            <DropdownMenuSeparator />
            <ConfirmationDialog
              trigger={
                <DropdownMenuItem
                  className="cursor-pointer"
                  onSelect={(e) => e.preventDefault()}
                >
                  Terminate
                </DropdownMenuItem>
              }
              headerText="Terminate Job Run?"
              description="Are you sure you want to terminate this job run?"
              onConfirm={async () => onTerminate()}
              buttonText="Terminate"
              buttonVariant="default"
              buttonIcon={<Cross2Icon />}
            />
            <DropdownMenuSeparator />
          </div>
        )}

        <DeleteConfirmationDialog
          trigger={
            <DropdownMenuItem
              className="cursor-pointer"
              onSelect={(e) => e.preventDefault()}
            >
              Delete
            </DropdownMenuItem>
          }
          headerText="Delete Job Run?"
          description="Are you sure you want to delete this job run?"
          onConfirm={async () => {
            await onDelete();
          }}
        />
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
