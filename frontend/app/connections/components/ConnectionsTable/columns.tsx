'use client';

import { ColumnDef } from '@tanstack/react-table';

import NextLink from 'next/link';

import {
  Connection,
  ConnectionConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface GetColumnsProps {
  onConnectionDeleted(id: string): void;
}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<Connection>>[] {
  const { onConnectionDeleted } = props;
  return [
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Connection" />
      ),
      cell: ({ row }) => (
        <div>
          <NextLink
            className="hover:underline"
            href={`/connections/${row.getValue('id')}`}
          >
            <span>{row.getValue('id')}</span>
          </NextLink>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Name" />
      ),
      cell: ({ row }) => {
        // const label = labels.find((label) => label.value === row.original.label);

        return (
          <div className="flex space-x-2">
            {/* {label && <Badge variant="outline">{label.label}</Badge>} */}
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('name')}
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'category',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Category" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            {/* {label && <Badge variant="outline">{label.label}</Badge>} */}
            <span className="max-w-[500px] truncate font-medium">
              {getCategory(row.original.connectionConfig)}
            </span>
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
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
        <DataTableRowActions
          row={row}
          onDeleted={() => onConnectionDeleted(row.id)}
        />
      ),
    },
  ];
}

function getCategory(cc?: PlainMessage<ConnectionConfig>): string {
  if (!cc) {
    return '-';
  }
  switch (cc.config.case) {
    case 'pgConfig':
      return 'Postgres';
    case 'awsS3Config':
      return 'AWS S3';
    default:
      return '-';
  }
}
