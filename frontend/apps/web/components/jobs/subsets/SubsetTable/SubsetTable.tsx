import FastTable from '@/components/FastTable/FastTable';
import {
  ColumnDef,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  RowData,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement } from 'react';
import { SubsetTableToolbar } from './SubsetTableToolbar';

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
  onBulkEdit(data: TData[], onClearSelection: () => void): void;
}

export default function SubsetTable<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement<any> {
  const { data, columns, onEdit, onReset, hasLocalChange, onBulkEdit } = props;

  const table = useReactTable({
    data,
    columns,
    enableRowSelection: true,
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
      <SubsetTableToolbar
        isFilterButtonDisabled={table.getState().columnFilters.length === 0}
        onClearFilters={() => table.resetColumnFilters()}
        isBulkEditButtonDisabled={
          Object.keys(table.getState().rowSelection).length <= 1
        }
        onBulkEditClick={() => {
          const selectedRows = table
            .getSelectedRowModel()
            .rows.map((row) => row.original);
          onBulkEdit(selectedRows, () => table.resetRowSelection());
        }}
      />
      <FastTable table={table} estimateRowSize={() => 33} rowOverscan={50} />
    </div>
  );
}
