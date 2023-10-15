'use client';

import { Button } from '@/components/ui/button';
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

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
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
    <Dialog>
      <DropdownMenu modal={false}>
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
          <DropdownMenuItem>
            <DialogTrigger className="w-full text-left">Delete</DialogTrigger>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Job?</DialogTitle>
        </DialogHeader>
        <DialogDescription className="pt-8">
          Deleting this Job will also delete all of the Job Runs. Are you sure
          you want to delete this Job?
        </DialogDescription>
        <DialogFooter className="flex flex-row justify-between w-full">
          <Button variant="destructive" onClick={() => onDelete()}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
