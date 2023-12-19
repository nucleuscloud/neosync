'use client';
import { useAccount } from '@/components/providers/account-provider';
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
import { useGetJobNextRuns } from '@/libs/hooks/useGetJobNextRuns';
import { formatDateTime } from '@/util/util';
import { JobStatus } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
  status?: JobStatus;
}

export default function JobNextRuns({ jobId, status }: Props): ReactElement {
  const { account } = useAccount();
  const { data, isLoading, error } = useGetJobNextRuns(
    account?.id ?? '',
    jobId
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
          <Table className="pt-5">
            <TableHeader className="bg-gray-100 dark:bg-gray-800">
              <TableRow>
                <TableHead>Upcoming Runs</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(status && status == JobStatus.PAUSED) ||
              data?.nextRuns?.nextRunTimes.length == 0 ? (
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
    </div>
  );
}
