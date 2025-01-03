'use client';

import { ColumnDef } from '@tanstack/react-table';

import TruncatedText from '@/components/TruncatedText';
import { Badge } from '@/components/ui/badge';
import { formatDateTime } from '@/util/util';
import { Timestamp, timestampDate } from '@bufbuild/protobuf/wkt';
import { JobStatus } from '@neosync/sdk';
import NextLink from 'next/link';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

const JOB_STATUS = [
  {
    value: JobStatus.ENABLED,
    badge: <Badge className="bg-blue-600">Enabled</Badge>,
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

interface JobColumn {
  id: string;
  name: string;
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  status: JobStatus;
}

interface GetJobsProps {
  accountName: string;
  onDeleted(id: string): void;
}

export function getColumns(props: GetJobsProps): ColumnDef<JobColumn>[] {
  const { onDeleted, accountName } = props;
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
        <DataTableColumnHeader column={column} title="Job Id" />
      ),
      cell: ({ row }) => {
        return (
          <div>
            <NextLink
              className="hover:underline"
              href={`/${accountName}/jobs/${row.getValue('id')}`}
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
              <div>
                <NextLink
                  className="hover:underline"
                  href={`/${accountName}/jobs/${row.getValue('id')}`}
                >
                  <TruncatedText text={row.getValue('name')} align="start" />
                </NextLink>
              </div>
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
              <Badge variant="outline">{row.getValue('type')}</Badge>
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
        const ts = row.getValue<Timestamp>('createdAt') ?? {
          nanos: 0,
          seconds: 0,
        };
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(timestampDate(ts))}
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
        const ts = row.getValue<Timestamp>('updatedAt') ?? {
          nanos: 0,
          seconds: 0,
        };

        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(timestampDate(ts))}
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
