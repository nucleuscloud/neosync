'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
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
import { useGetJobRecentRuns } from '@/libs/hooks/useGetJobRecentRuns';
import { formatDateTime } from '@/util/util';
import { ReactElement } from 'react';

interface Props {
  jobId: string;
}

export default function JobRecentRuns({ jobId }: Props): ReactElement {
  const { data, isLoading, error } = useGetJobRecentRuns(jobId);

  if (isLoading) {
    return <Skeleton className="w-full h-full" />;
  }
  return (
    <Card className="p-2">
      {!data?.recentRuns || error ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve recent runs`}</AlertTitle>
        </Alert>
      ) : (
        <Table>
          <TableCaption>Recent job runs</TableCaption>

          <TableHeader>
            <TableRow>
              <TableHead>Start Time</TableHead>
              <TableHead>Run Id</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.recentRuns?.runs.map((r) => {
              return (
                <TableRow key={r.jobRunId}>
                  <TableCell>
                    <span className="font-medium">
                      {formatDateTime(r.startTime?.toDate())}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="font-medium">{r.jobRunId}</span>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      )}
    </Card>
  );
}
