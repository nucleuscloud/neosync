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
import { ArrowRightIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { data, isLoading, error } = useGetJobRecentRuns(jobId);

  const { data: jobRuns, isLoading: jobRunsLoading } =
    useGetJobRunsByJob(jobId);

  const router = useRouter();

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
          <CardTitle className="py-6 px-2">Recent Job Runs</CardTitle>
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
              {data?.recentRuns?.runs.map((r) => {
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
