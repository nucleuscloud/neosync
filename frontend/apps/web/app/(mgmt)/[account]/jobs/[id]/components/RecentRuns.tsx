'use client';
import JobRunStatus from '@/app/(mgmt)/[account]/runs/components/JobRunStatus';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useGetJobRecentRuns } from '@/libs/hooks/useGetJobRecentRuns';
import { useGetJobRunsByJob } from '@/libs/hooks/useGetJobRunsByJob';
import { formatDateTime } from '@/util/util';
import { ArrowRightIcon, ReloadIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

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

  const recentRuns = recentRunsData?.recentRuns ?? [];

  function onRefreshClick(): void {
    recentRunsMutate();
    jobsRunsMutate();
  }

  const jobRunsIdMap = new Map(jobRunsData?.jobRuns.map((jr) => [jr.id, jr]));

  if (isLoading || jobRunsLoading) {
    return <Skeleton className="w-full h-full" />;
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
              {recentRuns.reverse().map((r) => {
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
                        <Link href={`/${account?.name}/runs/${jobRun.id}`}>
                          <Button variant="ghost" size="icon">
                            <ArrowRightIcon />
                          </Button>
                        </Link>
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
