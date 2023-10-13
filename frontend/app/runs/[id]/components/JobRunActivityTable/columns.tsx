'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  JobRunEvent,
  JobRunEventMetadata,
  JobRunEventTaskError,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTimeMilliseconds } from '@/util/util';
import { Timestamp } from '@bufbuild/protobuf';
import {
  ChevronDownIcon,
  ChevronRightIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {}

export function getColumns(props: GetColumnsProps): ColumnDef<JobRunEvent>[] {
  const {} = props;
  return [
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Id" />
      ),
      cell: ({ row }) => <div>{row.getValue<number>('id').toString()}</div>,
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'startTime',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Time" />
      ),
      cell: ({ row }) => {
        const startTime = row.getValue<Timestamp>('startTime');
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {startTime && formatDateTimeMilliseconds(startTime.toDate())}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Type" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('type')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'metadata',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Metadata" />
      ),
      cell: ({ row }) => {
        const metadata = row.getValue<JobRunEventMetadata>('metadata');

        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              <pre>
                {JSON.stringify(metadata?.metadata?.value, undefined, 2)}
              </pre>
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'error',
      accessorFn: (originalRow, _) =>
        originalRow.tasks.find((t) => t.error)?.error,
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Error" />
      ),
      cell: ({ row }) => {
        const err = row.getValue<JobRunEventTaskError>('error');
        return (
          <div className={`flex space-x-2`}>
            <span className="max-w-[500px] truncate font-medium">
              {err && (
                <Alert variant="destructive">
                  <AlertTitle className="flex flex-row space-x-2 justify-center">
                    <ExclamationTriangleIcon />
                    <p>Error</p>
                  </AlertTitle>
                  <AlertDescription>
                    <pre>{JSON.stringify(err, undefined, 2)}</pre>
                  </AlertDescription>
                </Alert>
              )}
            </span>
          </div>
        );
      },
    },
    {
      id: 'expander',
      header: () => null,
      cell: ({ row }) => {
        return (
          <div>
            {row.getIsExpanded() ? (
              <ChevronDownIcon className="w-4 h-4" />
            ) : (
              <ChevronRightIcon className="w-4 h-4" />
            )}
          </div>
        );
      },
    },
  ];
}
