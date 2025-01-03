'use client';

import { Badge } from '@/components/ui/badge';
import {
  getTransformerDataTypesString,
  getTransformerJobTypesString,
  getTransformerSourceString,
} from '@/util/util';
import { SystemTransformer } from '@neosync/sdk';
import { ColumnDef } from '@tanstack/react-table';
import NextLink from 'next/link';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

interface GetSystemTransformercolumnsProps {
  accountName: string;
}

export function getSystemTransformerColumns(
  props: GetSystemTransformercolumnsProps
): ColumnDef<SystemTransformer>[] {
  const { accountName } = props;
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
                  href={`/${accountName}/transformers/systemTransformers/${row.original.source}`}
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
      id: 'jobTypes',
      accessorKey: 'jobTypes',
      accessorFn: (row) => getTransformerJobTypesString(row.supportedJobTypes),
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Job Types" />
      ),
      cell: ({ getValue }) => {
        return (
          <div className="flex space-x-2">
            {getValue<string[]>().map((item) => (
              <Badge variant="outline" key={item}>
                {item}
              </Badge>
            ))}
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
      id: 'description',
      accessorKey: 'description',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Description" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {row.original.description}
            </span>
          </div>
        );
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => <DataTableRowActions row={row} />,
    },
  ];
}
