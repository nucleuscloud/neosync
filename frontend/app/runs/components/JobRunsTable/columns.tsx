'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';

import { JobRun } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { JOB_RUN_STATUS } from '../status';
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
        const status = JOB_RUN_STATUS.find(
          (status) => status.value === row.getValue('status')
        );

        if (!status) {
          return null;
        }

        return (
          <div className="flex w-[100px] items-center">{status.badge}</div>
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
      cell: ({ row }) => <div>{row.getValue('id')}</div>,
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
          ? formatDateTime(row.getValue<Timestamp>('completedAt').toDate())
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
