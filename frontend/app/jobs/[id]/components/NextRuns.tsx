'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useGetJobNextRuns } from '@/libs/hooks/useGetJobNextRuns';
import { JobStatus } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
}

export default function JobNextRuns({ jobId, status }: Props): ReactElement {
  const { data, isLoading, error } = useGetJobNextRuns(jobId);

  if (isLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  return (
    <Card className="p-2">
      {!data?.nextRuns || error ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve recent runs`}</AlertTitle>
        </Alert>
      ) : (
        <div>
          <Table className="pt-5">
            <TableHeader className="bg-gray-100">
              <TableRow>
                <TableHead>Upcoming Runs</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {status && status == JobStatus.PAUSED ? (
                <TableRow>
                  <TableCell>
                    <span className="font-medium">No upcoming runs</span>
                  </TableCell>
                </TableRow>
              ) : (
                data?.nextRuns?.nextRunTimes.map((r, index) => {
                  return (
                    <TableRow key={`${r}-${index}`}>
                      <TableCell>
                        <span className="font-medium">
                          {formatDateTime(r.toDate())}
                        </span>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </div>
      )}
    </Card>
  );
}
