'use client';

import { Badge } from '@/components/ui/badge';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { UserDefinedTransformer } from '@neosync/sdk';
import { ColumnDef } from '@tanstack/react-table';
import NextLink from 'next/link';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface getUserDefinedTransformerColumnsProps {
  onTransformerDeleted(id: string): void;
  accountName: string;
}

export function getUserDefinedTransformerColumns(
  props: getUserDefinedTransformerColumnsProps
): ColumnDef<PlainMessage<UserDefinedTransformer>>[] {
  const { onTransformerDeleted, accountName } = props;

  return [
    {
      id: 'name',
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
                  href={`/${accountName}/transformers/${row.original.id}`}
                >
                  {row.original.name}
                </NextLink>
              </div>
            </span>
          </div>
        );
      },
    },
    {
      id: 'type',
      accessorKey: 'type',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <Badge variant="outline">{row.original.dataType}</Badge>
          </div>
        );
      },
    },
    {
      accessorKey: 'source',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Source" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.original.source}
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
              {row.original.createdAt &&
                formatDateTime(row.getValue<Timestamp>('createdAt').toDate())}
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
              {row.original.updatedAt &&
                formatDateTime(row.getValue<Timestamp>('updatedAt').toDate())}
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
      accessorKey: 'actions',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Actions" />
      ),
      cell: ({ row }) => (
        <DataTableRowActions
          row={row}
          onDeleted={() => onTransformerDeleted(row.id)}
        />
      ),
    },
  ];
}
