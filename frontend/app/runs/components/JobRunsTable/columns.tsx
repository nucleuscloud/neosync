'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';

import {
  JobRun,
  JobRunStatus,
} from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface GetColumnsProps {
  onDeleted(id: string): void;
}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<JobRun>>[] {
  const { onDeleted } = props;
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-[2px]"
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const status = statuses.find(
          (status) => status.value === row.getValue('status')
        );

        if (!status) {
          return null;
        }

        return (
          <div className="flex w-[100px] items-center">
            {/* {status.icon && (
              <status.icon className="mr-2 h-4 w-4 text-muted-foreground bg-red-700" />
            )}
            <span>{status.label}</span> */}
            {status.badge}
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job Run" />
      ),
      cell: ({ row }) => <div className="w-[80px]">{row.getValue('id')}</div>,
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
      accessorKey: 'jobId',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job Id" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('jobId')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'startedAt',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Started At" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(row.getValue<Timestamp>('startedAt').toDate())}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    {
      accessorKey: 'completedAt',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Completed At" />
      ),
      cell: ({ row }) => {
        const completedAt = row.getValue('completedAt')
          ? formatDateTime(row.getValue<Timestamp>('closedAt').toDate())
          : undefined;
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {completedAt}
            </span>
          </div>
        );
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

export const statuses = [
  {
    value: JobRunStatus.ERROR,
    badge: <Badge variant="destructive">Error</Badge>,
  },
  {
    value: JobRunStatus.COMPLETE,
    badge: <Badge className="bg-green-600">Complete</Badge>,
  },
  {
    value: JobRunStatus.FAILED,
    badge: <Badge variant="destructive">Error</Badge>,
  },
  {
    value: JobRunStatus.RUNNING,
    badge: <Badge className="bg-blue-600">Running</Badge>,
  },
  {
    value: JobRunStatus.PENDING,
    badge: <Badge className="bg-purple-600">Running</Badge>,
  },
  {
    value: JobRunStatus.TERMINATED,
    badge: <Badge variant="destructive">Terminated</Badge>,
  },
  {
    value: JobRunStatus.CANCELED,
    badge: <Badge className="bg-yellow-600">Terminated</Badge>,
  },
  {
    value: JobRunStatus.UNSPECIFIED,
    badge: <Badge variant="outline">Unknown</Badge>,
  },
];
