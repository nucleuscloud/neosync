'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';

import { Badge } from '@/components/ui/badge';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { handleTransformerMetadata } from '../../EditTransformerOptions';
import { DataTableColumnHeader } from './data-table-column-header';
import { DataTableRowActions } from './data-table-row-actions';

// interface GetTransformerProps {
//   onTransformerDeleted(id: string): void;
// }

export function getColumns(): ColumnDef<PlainMessage<Transformer>>[] {
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
      accessorKey: 'category',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Category" />
      ),
      cell: () => {
        // const t = handleTransformerMetadata(row.original.value);

        return (
          <div className="flex space-x-2">
            <Badge variant="outline" className="bg-blue-100">
              system
            </Badge>
          </div>
        );
      },
    },
    {
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

//define the two tranformer types: system-generated and custom

// function getCategory(cc?: PlainMessage<TransformerConfig>): string {
//   if (!cc) {
//     return '-';
//   }
//   switch (cc.config.case) {
//     case 'pgConfig':
//       return 'Database';
//     case 'awsS3Config':
//       return 'File Storage';
//     default:
//       return '-';
//   }
// }
