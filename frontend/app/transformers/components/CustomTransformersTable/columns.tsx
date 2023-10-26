'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';

import { Badge } from '@/components/ui/badge';
import { CustomTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { handleTransformerMetadata } from '../../EditTransformerOptions';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

// interface GetTransformerProps {
//   onTransformerDeleted(id: string): void;
// }

export function getCustomTransformerColumns(): ColumnDef<
  PlainMessage<CustomTransformer>
>[] {
  // props: GetTransformerProps
  // const { onTransformerDeleted } = props;

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
        const t = handleTransformerMetadata(row.original.name);

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
        const t = handleTransformerMetadata(row.original.name);

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
        const t = handleTransformerMetadata(row.original.name);

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
      cell: ({ row }) => (
        <DataTableRowActions
          row={row}
          onDeleted={() => console.log('delete')}
          //onTransformerDeleted(row.id)}
        />
      ),
    },
  ];
}
