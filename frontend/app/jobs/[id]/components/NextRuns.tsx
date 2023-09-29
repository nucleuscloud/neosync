'use client';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Card } from '@/components/ui/card';
import {
  Table,
  TableBody,
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

export default function JobNextRuns({ job }: Props): ReactElement {
  return (
    <Card>
      {!job.nextRuns ? (
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve upcoming runs`}</AlertTitle>
        </Alert>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px] text-center">
                Upcoming Runs
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {job.nextRuns?.nextRunTimes.map((r) => {
              return (
                <TableRow key={r.toDate().toString()}>
                  <TableCell className="text-center">
                    <span className="font-medium">
                      {formatDateTime(r.toDate())}
                    </span>
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
