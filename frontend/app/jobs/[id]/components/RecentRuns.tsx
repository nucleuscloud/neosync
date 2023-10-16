'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useGetJobNextRuns } from '@/libs/hooks/useGetJobNextRuns';
import { useGetJobRecentRuns } from '@/libs/hooks/useGetJobRecentRuns';
import { formatDateTime } from '@/util/util';
import { ArrowRightIcon, ClockIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { data, isLoading, error } = useGetJobRecentRuns(jobId);

  const { data: nextJob, isLoading: nextRunLoading } = useGetJobNextRuns(jobId);

  const router = useRouter();

  if (isLoading || nextRunLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  const nr = nextJob?.nextRuns?.nextRunTimes[0]?.toDate();

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
            <TableCaption>
              {nr && (
                <div className="flex flex-row space-x-2 items-center justify-center">
                  <ClockIcon />
                  <div>Next run scheduled for: </div>
                  {formatDateTime(nextJob?.nextRuns?.nextRunTimes[0]?.toDate())}
                </div>
              )}
            </TableCaption>
            <TableHeader>
              <TableRow>
                <TableHead>Run Id</TableHead>
                <TableHead>Start Time</TableHead>
                <TableHead>Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data?.recentRuns?.runs.map((r) => {
                return (
                  <TableRow key={r.jobRunId}>
                    <TableCell>
                      <span className="font-medium">{r.jobRunId}</span>
                    </TableCell>
                    <TableCell>
                      <span className="font-medium">
                        {formatDateTime(r.startTime?.toDate())}
                      </span>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        onClick={() => router.push(`/runs/${r.jobRunId}`)}
                      >
                        <ArrowRightIcon />
                      </Button>
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
