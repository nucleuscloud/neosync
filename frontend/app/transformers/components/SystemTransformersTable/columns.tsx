'use client';

import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { ColumnDef } from '@tanstack/react-table';
import { handleTransformerMetadata } from '../../EditTransformerOptions';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

export function getSystemTransformerColumns(): ColumnDef<
  PlainMessage<Transformer>
>[] {
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
        const t = handleTransformerMetadata(row.original.value);

        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">{t.name}</span>
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
        const t = handleTransformerMetadata(row.original.value);

        return (
          <div className="flex space-x-2">
            <Badge variant="outline">{t.type}</Badge>
          </div>
        );
      },
    },
    {
      accessorKey: 'description',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Description" />
      ),
      cell: ({ row }) => {
        const t = handleTransformerMetadata(row.original.value);

        return (
          <div className="flex space-x-2">
            <span className="max-w-[500px] truncate font-medium">
              {t.description}
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
