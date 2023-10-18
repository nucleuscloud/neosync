'use client';

import { Button } from '@/components/ui/button';
import { Pencil1Icon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { ReactElement } from 'react';
import { DataTableColumnHeader } from './data-table-column-header';

export interface TableRow {
  schema: string;
  table: string;
  where?: string;
}

interface GetColumnsProps {
  onEdit(schema: string, table: string): void;
}

export function getColumns(props: GetColumnsProps): ColumnDef<TableRow>[] {
  const { onEdit } = props;
  return [
    {
      id: 'select',
      // header: ({ table }) => (
      //   <Checkbox
      //     checked={table.getIsAllPageRowsSelected()}
      //     onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
      //     aria-label="Select all"
      //     className="translate-y-[2px]"
      //   />
      // ),
      cell: ({ row }) => (
        <div />
        // <Checkbox
        //   checked={row.getIsSelected()}
        //   onCheckedChange={(value) => row.toggleSelected(!!value)}
        //   aria-label="Select row"
        //   className="translate-y-[2px] "
        // />
      ),
      enableSorting: false,
      enableHiding: false,
      enableColumnFilter: false,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Schema" />
      ),
      cell: ({ row }) => <div>{row.getValue('schema')}</div>,
      enableSorting: true,
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Table" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('table')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'where',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Where" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('where')}
            </span>
          </div>
        );
      },
      enableSorting: false,
      enableColumnFilter: false,
      // filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'edit',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Edit" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <EditAction
              onClick={() =>
                onEdit(row.getValue('schema'), row.getValue('table'))
              }
            />
          </div>
        );
      },
      enableSorting: false,
      enableColumnFilter: false,
      // filterFn: 'arrIncludesSome',
    },
  ];
}

interface EditActionProps {
  onClick(): void;
}

function EditAction(props: EditActionProps): ReactElement {
  const { onClick } = props;
  return (
    <Button variant="outline" size="icon" onClick={() => onClick()}>
      <Pencil1Icon />
    </Button>
  );
}
