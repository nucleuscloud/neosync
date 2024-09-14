'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatDateTime } from '@/util/util';
import { useQuery } from '@connectrpc/connect-query';
import { JobStatus } from '@neosync/sdk';
import { getJobNextRuns } from '@neosync/sdk/connectquery';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
}

export default function JobNextRuns({ jobId, status }: Props): ReactElement {
  const { data, isLoading, error } = useQuery(
    getJobNextRuns,
    { jobId },
    { enabled: !!jobId }
  );

  if (isLoading) {
    return <Skeleton className="w-full h-full" />;
  }

  return (
    <div>
      {!data?.nextRuns || error ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve recent runs`}</AlertTitle>
        </Alert>
      ) : (
        <div>
          <Table>
            <TableHeader className="bg-gray-100 dark:bg-gray-800">
              <TableRow>
                <TableHead className="pl-4">Upcoming Runs</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(status && status === JobStatus.PAUSED) ||
              data?.nextRuns?.nextRunTimes.length === 0 ? (
                <TableRow className="hover:bg-background">
                  <TableCell>
                    <span className="font-medium justify-center flex pt-20">
                      No upcoming runs
                    </span>
                  </TableCell>
                </TableRow>
              ) : (
                data?.nextRuns?.nextRunTimes.slice(0, 4).map((r, index) => {
                  return (
                    <TableRow key={`${r}-${index}`}>
                      <TableCell className="py-3">
                        <span className="font-medium">
                          {formatDateTime(r?.toDate())}
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
    </div>
  );
}
