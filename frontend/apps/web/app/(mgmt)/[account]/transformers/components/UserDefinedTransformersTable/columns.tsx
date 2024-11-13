'use client';

import TruncatedText from '@/components/TruncatedText';
import { Badge } from '@/components/ui/badge';
import {
  formatDateTime,
  getTransformerDataTypesString,
  getTransformerSourceString,
} from '@/util/util';
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
                  <TruncatedText text={row.original.name} align="start" />
                </NextLink>
              </div>
            </span>
          </div>
        );
      },
    },
    {
      id: 'types',
      accessorKey: 'types',
      accessorFn: (row) => getTransformerDataTypesString(row.dataTypes),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Data Types" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex space-x-2">
            <Badge variant="outline">{getValue<string>()}</Badge>
          </div>
        );
      },
    },
    {
      accessorKey: 'source',
      accessorFn: (row) => getTransformerSourceString(row.source),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Source" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {getValue<string>()}
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
                formatDateTime(row.getValue<Timestamp>('createdAt')?.toDate())}
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
                formatDateTime(row.getValue<Timestamp>('updatedAt')?.toDate())}
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
