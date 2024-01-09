'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';
import NextLink from 'next/link';
import { Badge } from '@/components/ui/badge';
import { formatDateTime } from '@/util/util';
import { PlainMessage, Timestamp } from '@bufbuild/protobuf';
import { UserDefinedTransformer } from '@neosync/sdk';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface Props {
  onTransformerDeleted(id: string): void;
}

export function getUserDefinedTransformerColumns(
  props: Props
): ColumnDef<PlainMessage<UserDefinedTransformer>>[] {
  const { onTransformerDeleted } = props;

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
                    href={`/${account?.name}/transformers/${transformer.id}`}
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
            <NextLink
              className="hover:underline"
              href={`/${account?.name}/transformers/${row.original.id}`}
            >
              <Badge variant="outline">{row.original.dataType}</Badge>
            </NextLink>
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
            <NextLink
              className="hover:underline"
              href={`/${account?.name}/transformers/${row.original.id}`}
            >
              <span className="max-w-[500px] truncate font-medium">
                {row.original.source}
              </span>
            </NextLink>
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
