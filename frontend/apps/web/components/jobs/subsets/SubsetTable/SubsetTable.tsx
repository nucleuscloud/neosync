import {
  ColumnDef,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement, useRef } from 'react';

interface Props<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
}

export default function SubsetTable<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const { data, columns } = props;

  const table = useReactTable({
    data,
    columns,
    enableRowSelection: false,
    initialState: {},
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    meta: {
      // subsetTable: {},
    },
  });

  const { rows } = table.getRowModel();
  const tableContainerRef = useRef<HTMLDivElement>(null);
}
