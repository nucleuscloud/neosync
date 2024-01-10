'use client';
import JobRunStatus from '@/app/(mgmt)/[account]/runs/components/JobRunStatus';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
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
import { JobRun } from '@neosync/sdk';
import { ArrowRightIcon, ReloadIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement, useEffect } from 'react';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { account } = useAccount();
  const { data, isLoading, error, mutate, isValidating } = useGetJobRecentRuns(
    account?.id ?? '',
    jobId
  );

  const {
    data: jobRuns,
    isLoading: jobRunsLoading,
    mutate: jobsRunsMutate,
    isValidating: jobRunsValidating,
  } = useGetJobRunsByJob(account?.id ?? '', jobId);

  function onRefreshClick(): void {
    mutate();
    jobsRunsMutate();
  }

  useEffect(() => {
    // Set a timeout to refresh once after 3 second
    // used to show new runs after trigger
    const timeoutId = setTimeout(() => {
      onRefreshClick();
    }, 3000);
    return () => clearTimeout(timeoutId);
  }, []);

  const jobRunsIdMap =
    jobRuns?.jobRuns.reduce(
      (prev, curr) => {
        return { ...prev, [curr.id]: curr };
      },
      {} as Record<string, JobRun>
    ) || {};

  if (isLoading || jobRunsLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  return (
    <Card>
      {!data?.recentRuns || error ? (
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
              {data?.recentRuns?.reverse().map((r) => {
                const jobRun = jobRunsIdMap[r.jobRunId];
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
                        {jobRun ? (
                          <JobRunStatus status={jobRun.status} />
                        ) : (
                          <Badge className="bg-gray-600">Archived</Badge>
                        )}
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
