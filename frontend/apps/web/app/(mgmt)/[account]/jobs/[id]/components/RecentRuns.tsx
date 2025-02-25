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
import { formatDateTime, getErrorMessage } from '@/util/util';
import { timestampDate } from '@bufbuild/protobuf/wkt';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { JobRunStatus as JobRunStatusEnum, JobService } from '@neosync/sdk';
import {
  Cross2Icon,
  DotsHorizontalIcon,
  ReloadIcon,
} from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { toast } from 'sonner';
import JobRunStatus from '../../../runs/components/JobRunStatus';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement<any> {
  const { account } = useAccount();
  const {
    data: recentRunsData,
    isLoading,
    error: recentRunsError,
    refetch: recentRunsMutate,
    isFetching: isValidating,
  } = useQuery(
    JobService.method.getJobRecentRuns,
    { jobId },
    { enabled: !!jobId }
  );
  const {
    data: jobRunsData,
    isLoading: jobRunsLoading,
    isFetching: jobRunsValidating,
    refetch: jobRunsMutate,
  } = useQuery(
    JobService.method.getJobRuns,
    { id: { case: 'jobId', value: jobId } },
    { enabled: !!jobId }
  );
  const { mutateAsync: removeJobRunAsync } = useMutation(
    JobService.method.deleteJobRun
  );
  const { mutateAsync: cancelJobRunAsync } = useMutation(
    JobService.method.cancelJobRun
  );
  const { mutateAsync: terminateJobRunAsync } = useMutation(
    JobService.method.terminateJobRun
  );

  const recentRuns = (recentRunsData?.recentRuns ?? []).toReversed();
  const jobRunsIdMap = new Map(
    (jobRunsData?.jobRuns ?? []).map((jr) => [jr.id, jr])
  );
  const router = useRouter();

  if (isLoading || jobRunsLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  function onRefreshClick(): void {
    recentRunsMutate();
    jobRunsMutate();
  }

  async function onDelete(runId: string): Promise<void> {
    try {
      await removeJobRunAsync({ accountId: account?.id, jobRunId: runId });
      onRefreshClick();
      toast.success('Removing Job Run. This may take a minute to delete!');
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onCancel(runId: string): Promise<void> {
    try {
      await cancelJobRunAsync({ accountId: account?.id, jobRunId: runId });
      toast.success('Canceling Job Run. This may take a minute to cancel!');
      onRefreshClick();
    } catch (err) {
      console.error(err);
      toast.error('Unable to cancel job run', {
        description: getErrorMessage(err),
      });
    }
  }

  async function onTerminate(runId: string): Promise<void> {
    try {
      await terminateJobRunAsync({ accountId: account?.id, jobRunId: runId });
      toast.success(
        'Terminating Job Run. This may take a minute to terminate!'
      );
      onRefreshClick();
    } catch (err) {
      console.error(err);
      toast.error('Unable to terminate job run', {
        description: getErrorMessage(err),
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
          <div className="flex flex-row items-center px-6">
            <CardTitle className="py-6">Recent Job Runs</CardTitle>
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
          <Table>
            <TableHeader className="bg-gray-100 dark:bg-gray-800 ">
              <TableRow>
                <TableHead className="px-6">Run Id</TableHead>
                <TableHead className="px-6">Start At</TableHead>
                <TableHead className="px-6">Completed At</TableHead>
                <TableHead className="px-6">Status</TableHead>
                <TableHead className="px-6">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {recentRuns.map((r) => {
                const jobRun = jobRunsIdMap.get(r.jobRunId);
                return (
                  <TableRow key={r.jobRunId}>
                    <TableCell className="px-6">
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
                    <TableCell className="px-6">
                      <span className="font-medium">
                        {formatDateTime(
                          r.startTime ? timestampDate(r.startTime) : undefined
                        )}
                      </span>
                    </TableCell>
                    <TableCell className="px-6">
                      <span className="font-medium">
                        {jobRun &&
                          formatDateTime(
                            jobRun.completedAt
                              ? timestampDate(jobRun.completedAt)
                              : undefined
                          )}
                      </span>
                    </TableCell>
                    <TableCell className="px-6">
                      <JobRunStatus status={jobRun?.status} />
                    </TableCell>
                    <TableCell className="px-6">
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
                                  headerText="Cancel Job Run?"
                                  description="Are you sure you want to cancel this job run?"
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
                                  headerText="Terminate Job Run?"
                                  description="Are you sure you want to terminate this job run?"
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
                              headerText="Delete Job Run?"
                              description="Are you sure you want to delete this job run?"
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
