'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Job, JobStatus } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import NextLink from 'next/link';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

const JOB_STATUS = [
  {
    value: JobStatus.ENABLED,
    badge: <Badge className="bg-green-600">Enabled</Badge>,
  },
  {
    value: JobStatus.DISABLED,
    badge: <Badge variant="destructive">Error</Badge>,
  },
  {
    value: JobStatus.PAUSED,
    badge: <Badge className="bg-gray-600">Paused</Badge>,
  },
  {
    value: JobStatus.UNSPECIFIED,
    badge: <Badge variant="outline">Unknown</Badge>,
  },
];

interface GetJobsProps {
  onDeleted(id: string): void;
}

export function getColumns(
  props: GetJobsProps
): ColumnDef<PlainMessage<Job>>[] {
  const { onDeleted } = props;

  return [
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const status = JOB_STATUS.find(
          (status) => status.value === row.getValue('status')
        );

        if (!status) {
          return null;
        }

        return status.badge;
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job" />
      ),
      cell: ({ row }) => {
        return (
          <div>
            <NextLink
              className="hover:underline"
              href={`/jobs/${row.getValue('id')}`}
            >
              <span>{row.getValue('id')}</span>
            </NextLink>
          </div>
        );
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('name')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'createdAt',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Created At" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(row.getValue<Timestamp>('createdAt').toDate())}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'updatedAt',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Updated At" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(row.getValue<Timestamp>('updatedAt').toDate())}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <DataTableRowActions row={row} onDeleted={() => onDeleted(row.id)} />
      ),
    },
  ];
}
