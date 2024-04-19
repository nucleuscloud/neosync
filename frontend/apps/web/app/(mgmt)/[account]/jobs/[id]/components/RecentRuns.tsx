'use client';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardTitle } from '@/components/ui/card';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { toast } from '@/components/ui/use-toast';
import { useGetJobRecentRuns } from '@/libs/hooks/useGetJobRecentRuns';
import { useGetJobRunsByJob } from '@/libs/hooks/useGetJobRunsByJob';
import { formatDateTime, getErrorMessage } from '@/util/util';
import { JobRunStatus as JobRunStatusEnum } from '@neosync/sdk';
import {
  Cross2Icon,
  DotsHorizontalIcon,
  ReloadIcon,
} from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import JobRunStatus from '../../../runs/components/JobRunStatus';
import {
  cancelJobRun,
  removeJobRun,
  terminateJobRun,
} from '../../../runs/components/JobRunsTable/data-table-row-actions';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { account } = useAccount();
  const {
    data: recentRunsData,
    isLoading,
    error: recentRunsError,
    mutate: recentRunsMutate,
    isValidating,
  } = useGetJobRecentRuns(account?.id ?? '', jobId);
  const {
    data: jobRunsData,
    isLoading: jobRunsLoading,
    mutate: jobsRunsMutate,
    isValidating: jobRunsValidating,
  } = useGetJobRunsByJob(account?.id ?? '', jobId);

  const recentRuns = (recentRunsData?.recentRuns ?? []).toReversed();
  const jobRunsIdMap = new Map(
    (jobRunsData?.jobRuns ?? []).map((jr) => [jr.id, jr])
  );
  const router = useRouter();

  if (isLoading || jobRunsLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  async function onRefreshClick(): Promise<void> {
    await Promise.all([recentRunsMutate(), jobsRunsMutate()]);
  }

  async function onDelete(runId: string): Promise<void> {
    try {
      await removeJobRun(runId, account?.id ?? '');
      onRefreshClick();
      toast({
        title: 'Removing Job Run. This may take a minute to delete!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to remove job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onCancel(runId: string): Promise<void> {
    try {
      await cancelJobRun(runId, account?.id ?? '');
      toast({
        title: 'Canceling Job Run. This may take a minute to cancel!',
        variant: 'success',
      });
      onRefreshClick();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to cancel job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  async function onTerminate(runId: string): Promise<void> {
    try {
      await terminateJobRun(runId, account?.id ?? '');
      toast({
        title: 'Terminating Job Run. This may take a minute to terminate!',
        variant: 'success',
      });
      onRefreshClick();
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to terminate job run',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Card className="overflow-hidden">
      {recentRunsError ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve recent runs`}</AlertTitle>
        </Alert>
      ) : (
        <div>
          <div className="flex flex-row items-center px-2">
            <CardTitle className="py-6 pl-4">Recent Job Runs</CardTitle>
            <Button
              className={
                isValidating || jobRunsValidating ? 'animate-spin' : ''
              }
              disabled={isValidating || jobRunsValidating}
              variant="ghost"
              size="icon"
              onClick={() => onRefreshClick()}
            >
              <ReloadIcon className="h-4 w-4" />
            </Button>
          </div>
          <Table className="pt-5">
            <TableHeader className="bg-gray-100 dark:bg-gray-800 ">
              <TableRow>
                <TableHead className="pl-6">Run Id</TableHead>
                <TableHead>Start At</TableHead>
                <TableHead>Completed At</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {recentRuns.map((r) => {
                const jobRun = jobRunsIdMap.get(r.jobRunId);
                return (
                  <TableRow key={r.jobRunId}>
                    <TableCell className="pl-6">
                      {jobRun ? (
                        <Link
                          className="hover:underline"
                          href={`/${account?.name}/runs/${r.jobRunId}`}
                        >
                          <span className="font-medium">{r.jobRunId}</span>
                        </Link>
                      ) : (
                        <span className="font-medium">{r.jobRunId}</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="font-medium">
                        {formatDateTime(r.startTime?.toDate())}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="font-medium">
                        {jobRun && formatDateTime(jobRun.completedAt?.toDate())}
                      </span>
                    </TableCell>
                    <TableCell>
                      <span className="font-medium">
                        <JobRunStatus status={jobRun?.status} />
                      </span>
                    </TableCell>
                    <TableCell>
                      {jobRun && (
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
                          <DropdownMenuContent
                            align="end"
                            className="w-[160px]"
                          >
                            <DropdownMenuItem
                              className="cursor-pointer"
                              onClick={() =>
                                router.push(
                                  `/${account?.name}/runs/${jobRun.id}`
                                )
                              }
                            >
                              View
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            {(jobRun.status === JobRunStatusEnum.RUNNING ||
                              jobRun.status === JobRunStatusEnum.PENDING) && (
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
                                  headerText="Are you sure you want to cancel this job run?"
                                  description=""
                                  onConfirm={async () => onCancel(jobRun.id)}
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
                                  headerText="Are you sure you want to terminate this job run?"
                                  description=""
                                  onConfirm={async () => onTerminate(jobRun.id)}
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
                              headerText="Are you sure you want to delete this job run?"
                              description=""
                              onConfirm={async () => onDelete(jobRun.id)}
                            />
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </div>
      )}
    </Card>
  );
}
