'use client';

import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { PlainMessage } from '@bufbuild/protobuf';
import { SystemTransformer } from '@neosync/sdk';
import { ColumnDef } from '@tanstack/react-table';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';
import NextLink from 'next/link';

interface getSystemTransformerProps {
  account: {
    name: string;
    // Add other properties specific to account
  };
  transformer: {
    source: string;
    // Add other properties specific to transformer
  };
}
export function getSystemTransformerColumns(props: getSystemTransformerProps): ColumnDef<
  PlainMessage<SystemTransformer>
>[] {
  const { account, transformer } = props;
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
      id: 'value',
      accessorKey: 'value',
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
                  href={`/${account?.name}/transformers/systemTransformers/${transformer.source}`}
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
        <DataTableColumnHeader column={column} title="Type" />
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
