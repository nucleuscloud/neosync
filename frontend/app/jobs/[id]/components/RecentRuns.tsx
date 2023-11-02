'use client';
import JobRunStatus from '@/app/runs/components/JobRunStatus';
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
import { JobRun } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { ArrowRightIcon, ReloadIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement, useEffect } from 'react';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { data, isLoading, error, mutate, isValidating } =
    useGetJobRecentRuns(jobId);

  const {
    data: jobRuns,
    isLoading: jobRunsLoading,
    mutate: jobsRunsMutate,
    isValidating: jobRunsValidating,
  } = useGetJobRunsByJob(jobId);

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
    <Card className="p-2">
      {!data?.recentRuns || error ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve recent runs`}</AlertTitle>
        </Alert>
      ) : (
        <div>
          <div className="flex flex-row items-center">
            <CardTitle className="py-6 px-2">Recent Job Runs</CardTitle>
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
            <TableHeader>
              <TableRow>
                <TableHead>Run Id</TableHead>
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
                    <TableCell>
                      <Link
                        className="hover:underline"
                        href={`/runs/${r.jobRunId}`}
                      >
                        <span className="font-medium">{r.jobRunId}</span>
                      </Link>
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
                        {jobRun && <JobRunStatus status={jobRun.status} />}
                      </span>
                    </TableCell>
                    <TableCell>
                      {jobRun && (
                        <Link href={`/runs/${jobRun.id}`}>
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
