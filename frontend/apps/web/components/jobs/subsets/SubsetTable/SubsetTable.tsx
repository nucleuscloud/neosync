import ButtonText from '@/components/ButtonText';
import FastTable from '@/components/FastTable/FastTable';
import { Button } from '@/components/ui/button';
import { Cross2Icon } from '@radix-ui/react-icons';
import {
  ColumnDef,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  RowData,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement } from 'react';

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface TableMeta<TData extends RowData> {
    subsetTable?: {
      onEdit(rowIndex: number, schema: string, table: string): void;
      onReset(rowIndex: number, schema: string, table: string): void;
      hasLocalChange(rowIndex: number, schema: string, table: string): boolean;
    };
  }
}

interface Props<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
  onEdit(rowIndex: number, schema: string, table: string): void;
  onReset(rowIndex: number, schema: string, table: string): void;
  hasLocalChange(rowIndex: number, schema: string, table: string): boolean;
}

export default function SubsetTable<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const { data, columns, onEdit, onReset, hasLocalChange } = props;

  const table = useReactTable({
    data,
    columns,
    enableRowSelection: false,
    initialState: {},
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    meta: {
      subsetTable: {
        onEdit,
        onReset,
        hasLocalChange,
      },
    },
  });

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-row items-center gap-2 justify-end">
        <div className="flex flex-row items-center gap-2">
          <Button
            disabled={table.getState().columnFilters.length === 0}
            type="button"
            variant="outline"
            className="px-2 lg:px-3"
            onClick={() => table.resetColumnFilters()}
          >
            <ButtonText
              leftIcon={<Cross2Icon className="h-3 w-3" />}
              text="Clear filters"
            />
          </Button>
        </div>
      </div>
      <FastTable table={table} estimateRowSize={() => 33} rowOverscan={5} />
    </div>
  );
}
