'use client';

import { Button } from '@/components/ui/button';
import { TrashIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { VirtualForeignKeysColumnHeader } from './VirtualFkColumnHeaders';
import { Row as RowData } from './VirtualFkPageTable';

interface Props {
  removeVirtualForeignKey?: (index: number) => void;
}

export function getVirtualForeignKeysColumns(
  props: Props
): ColumnDef<RowData>[] {
  const { removeVirtualForeignKey } = props;
  return [
    {
      accessorFn: (row) => `${row.foreignKey.schema}.${row.foreignKey.table}`,
      id: 'sourceTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <VirtualForeignKeysColumnHeader column={column} title="Source Table" />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue<string>()}
          </span>
        );
      },
      maxSize: 500,
      size: 300,
    },
    {
      accessorFn: (row) => row.foreignKey.columns,
      id: 'sourceColumns',
      header: ({ column }) => (
        <VirtualForeignKeysColumnHeader
          column={column}
          title="Source Columns"
        />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue<string[]>().toString()}
          </span>
        );
      },
      maxSize: 500,
      size: 200,
    },
    {
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'targetTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <VirtualForeignKeysColumnHeader column={column} title="Target Table" />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue<string>()}
          </span>
        );
      },
      maxSize: 500,
      size: 300,
    },
    {
      accessorFn: (row) => row.columns,
      id: 'targetColumns',
      header: ({ column }) => (
        <VirtualForeignKeysColumnHeader
          column={column}
          title="Source Columns"
        />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue<string[]>().toString()}
          </span>
        );
      },
      maxSize: 500,
      size: 200,
    },
    {
      id: 'actions',
      header: ({ column }) => (
        <VirtualForeignKeysColumnHeader column={column} title="Actions" />
      ),
      cell: ({ row }) => {
        return (
          <div>
            {removeVirtualForeignKey && (
              <Button
                variant="destructive"
                size="sm"
                type="button"
                key="remove-vfk"
                onClick={() => {
                  removeVirtualForeignKey(row.index);
                }}
              >
                <TrashIcon />
              </Button>
            )}
          </div>
        );
      },
      maxSize: 500,
      size: 200,
    },
  ];
}
