'use client';

import { ColumnDef } from '@tanstack/react-table';

import NextLink from 'next/link';

import TruncatedText from '@/components/TruncatedText';
import { Badge, BadgeProps } from '@/components/ui/badge';
import { formatDateTime } from '@/util/util';
import { Timestamp, timestampDate } from '@bufbuild/protobuf/wkt';
import { AccountApiKey } from '@neosync/sdk';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface GetColumnsProps {
  onDeleted(id: string): void;
  accountName: string;
}

export function getColumns(props: GetColumnsProps): ColumnDef<AccountApiKey>[] {
  const { onDeleted, accountName } = props;
  return [
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Id" />
      ),
      cell: ({ row }) => (
        <div>
          <NextLink
            className="hover:underline"
            href={`/${accountName}/settings/api-keys/${row.getValue('id')}`}
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
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              <TruncatedText text={row.getValue('name')} align="start" />
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Status" />
      ),
      cell: ({ row }) => {
        const expiresAt = row.getValue<Timestamp>('expiresAt')
          ? timestampDate(row.getValue<Timestamp>('expiresAt')).getTime()
          : undefined;
        const text =
          !!expiresAt && expiresAt > Date.now() ? 'active' : 'expired';
        const badgeVariant: BadgeProps['variant'] =
          text === 'active' ? 'success' : 'destructive';

        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              <Badge variant={badgeVariant}>{text}</Badge>
            </span>
          </div>
        );
      },
    },
    {
      accessorKey: 'expiresAt',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Expires At" />
      ),
      cell: ({ row }) => {
        const expiresAt = row.getValue<Timestamp>('expiresAt');
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(
                expiresAt ? timestampDate(expiresAt) : new Date()
              )}
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
        const createdAt = row.getValue<Timestamp>('createdAt');
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(
                createdAt ? timestampDate(createdAt) : new Date()
              )}
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
        const updatedAt = row.getValue<Timestamp>('updatedAt');
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {formatDateTime(
                updatedAt ? timestampDate(updatedAt) : new Date()
              )}
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
