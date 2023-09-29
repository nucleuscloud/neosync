'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Job } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { ReactElement } from 'react';

interface Props {
  job: Job;
}

export default function JobRecentRuns({ job }: Props): ReactElement {
  return (
    <Card>
      {!job.recentRuns ? (
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
            {job.recentRuns?.map((r) => {
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
