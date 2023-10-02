'use client';

import { ColumnDef } from '@tanstack/react-table';

import { Checkbox } from '@/components/ui/checkbox';

import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { PlainMessage } from '@bufbuild/protobuf';
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
    // {
    //   accessorKey: 'id',
    //   header: ({ column }) => (
    //     <DataTableColumnHeader column={column} title="Transformer id" />
    //   ),
    //   cell: ({ row }) => <div className="w-[80px]">{row.getValue('id')}</div>,
    //   enableSorting: false,
    //   enableHiding: false,
    // },
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Title" />
      ),
      cell: ({ row }) => {
        // const label = labels.find((label) => label.value === row.original.label);

        return (
          <div className="flex space-x-2">
            {/* {label && <Badge variant="outline">{label.label}</Badge>} */}
            <span className="max-w-[500px] truncate font-medium">
              {/* {row.getValue('title')} */}
              {row.original.title}
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
        // const status = statuses.find(
        //   (status) => status.value === row.getValue('status')
        // );

        const status = { icon: 'fagit', label: 'status' };

        if (!status) {
          return null;
        }

        return (
          <div className="flex w-[100px] items-center">
            {status.icon && (
              // <status.icon className="mr-2 h-4 w-4 text-muted-foreground" />
              <div>{row.original.title}</div>
            )}
            {/* <span>{status.label}</span> */}
          </div>
        );
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id));
      },
    },
    // {
    //   accessorKey: 'category',
    //   header: ({ column }) => (
    //     <DataTableColumnHeader column={column} title="Category" />
    //   ),
    //   cell: ({ row }) => {
    //     return (
    //       <div className="flex space-x-2">
    //         {/* {label && <Badge variant="outline">{label.label}</Badge>} */}
    //         <span className="max-w-[500px] truncate font-medium">
    //           {/* {getCategory(row.original.connectionConfig)} */}
    //           {row.original.title}
    //         </span>
    //       </div>
    //     );
    //   },
    //   filterFn: (row, id, value) => {
    //     return value.includes(row.getValue(id));
    //   },
    // },
    // {
    //   accessorKey: 'createdAt',
    //   header: ({ column }) => (
    //     <DataTableColumnHeader column={column} title="Created At" />
    //   ),
    //   cell: ({ row }) => {
    //     return (
    //       <div className="flex space-x-2">
    //         <span className="max-w-[500px] truncate font-medium">
    //           {formatDateTime(row.getValue<Timestamp>('createdAt').toDate())}
    //         </span>
    //       </div>
    //     );
    //   },
    //   filterFn: (row, id, value) => {
    //     return value.includes(row.getValue(id));
    //   },
    // },
    // {
    //   accessorKey: 'updatedAt',
    //   header: ({ column }) => (
    //     <DataTableColumnHeader column={column} title="Updated At" />
    //   ),
    //   cell: ({ row }) => {
    //     return (
    //       <div className="flex space-x-2">
    //         <span className="max-w-[500px] truncate font-medium">
    //           {formatDateTime(row.getValue<Timestamp>('updatedAt').toDate())}
    //         </span>
    //       </div>
    //     );
    //   },
    //   filterFn: (row, id, value) => {
    //     return value.includes(row.getValue(id));
    //   },
    // },
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
