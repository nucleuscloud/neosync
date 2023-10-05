'use client';

import {
  Column,
  ColumnDef,
  ColumnFiltersState,
  FilterFn,
  SortingState,
  Table as TableType,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import * as React from 'react';

import SkeletonTable from '@/components/skeleton/SkeletonTable';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Tree, TreeDataItem } from '@/components/ui/tree';
import { cn } from '@/libs/utils';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { DataTableToolbar } from './data-table-toolbar';

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  transformers?: Transformer[];
  schemaMap: Record<string, Record<string, string>>;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  transformers,
  schemaMap,
}: DataTableProps<TData, TValue>) {
  const [rowSelection, setRowSelection] = React.useState({});
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [sorting, setSorting] = React.useState<SortingState>([]);

  const multiFilter: FilterFn<TData> = (row, columnId, value, addMeta) => {
    console.log(JSON.stringify(row));
    console.log('columnId', columnId);
    console.log('value', value);

    return true;
  };

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
    },
    filterFns: {
      multi: multiFilter,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  });

  const treedata = Object.keys(schemaMap).map((schema) => {
    return {
      id: schema,
      name: schema,
      isSelected: true,
      children: Object.keys(schemaMap[schema]).map((table) => {
        return {
          id: table,
          name: table,
          isSelected: true,
        };
      }),
    };
  });

  const other = [
    { id: '1', name: '1', isSelected: false },
    { id: '2', name: '2', isSelected: false },
    {
      id: '3',
      name: '3',
      isSelected: false,
      children: [
        { id: 'c1', name: 'c1', isSelected: false },
        { id: 'c2', name: 'c2', isSelected: false },
        { id: 'c3', name: 'c3', isSelected: false },
      ],
    },
    {
      id: '4',
      name: '4',
      isSelected: false,
      children: [
        {
          id: 'd1',
          name: 'd1',
          isSelected: false,
          children: [
            { id: 'd11', name: 'd11', isSelected: false },
            { id: 'd12', name: 'd12', isSelected: false },
            { id: 'd13', name: 'd13', isSelected: false },
          ],
        },
        { id: 'd2', name: 'd2', isSelected: false },
        { id: 'd3', name: 'd3', isSelected: false },
      ],
    },
    {
      id: '5',
      name: '5',
      isSelected: false,
      children: [
        {
          id: 'e1',
          name: 'e1',
          isSelected: false,
          children: [
            { id: 'e11', name: 'e11', isSelected: false },
            { id: 'e12', name: 'e12', isSelected: false },
            { id: 'e13', name: 'e13', isSelected: false },
          ],
        },
        { id: 'e2', name: 'e2', isSelected: false },
        { id: 'e3', name: 'e3', isSelected: false },
      ],
    },
  ];

  function handlefilter(items: TreeDataItem[]) {
    // table.setColumnFilters()
    setColumnFilters([{ id: 'schema', value: ['public', 'sales'] }]);
  }

  if (!data) {
    return <SkeletonTable />;
  }

  return (
    <div>
      <div className="flex flex-row">
        <div className="w-[230px] mb-10 "></div>
        <div className="w-full  ">
          <DataTableToolbar table={table} transformers={transformers} />
        </div>
      </div>

      <div className="flex flex-row">
        <div className="basis-1/3 mb-10">
          <Tree
            data={treedata}
            className="h-full w-[200px] border rounded-md border-r-0"
            onSelectChange={handlefilter}
          />
        </div>
        <div className="basis-2/3">
          <div className="rounded-md border">
            <ScrollArea className="h-[700px]">
              <Table>
                <TableHeader className="sticky top-0">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => {
                        return (
                          <TableHead key={header.id}>
                            {header.isPlaceholder
                              ? null
                              : flexRender(
                                  header.column.columnDef.header,
                                  header.getContext()
                                )}
                            {header.column.getCanFilter() ? (
                              <div>
                                <FilterSelect
                                  column={header.column}
                                  table={table}
                                  transformers={transformers || []}
                                />
                              </div>
                            ) : null}
                          </TableHead>
                        );
                      })}
                    </TableRow>
                  ))}
                </TableHeader>
                <TableBody>
                  {table.getRowModel().rows?.length ? (
                    table.getRowModel().rows.map((row) => (
                      <TableRow
                        key={row.id}
                        data-state={row.getIsSelected() && 'selected'}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell key={cell.id}>
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell
                        colSpan={columns.length}
                        className="h-24 text-center"
                      >
                        No results.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </ScrollArea>
          </div>
          <div className="flex-1 text-sm text-muted-foreground">
            {table.getFilteredSelectedRowModel().rows.length} of{' '}
            {table.getFilteredRowModel().rows.length} row(s) selected.
          </div>
        </div>
      </div>
    </div>
  );
}

interface FilterSelectProps<TData, TValue> {
  column: Column<TData, TValue>;
  table: TableType<TData>;
  transformers: Transformer[];
}

function FilterSelect<TData, TValue>(props: FilterSelectProps<TData, TValue>) {
  const { column, table, transformers } = props;
  const [open, setOpen] = React.useState(false);
  const firstValue = table
    .getPreFilteredRowModel()
    .flatRows[0]?.getValue(column.id || '');

  const columnFilterValue = (column.getFilterValue() as string[]) || [];

  const sortedUniqueValues = React.useMemo(
    () =>
      typeof firstValue === 'number'
        ? []
        : Array.from(column.getFacetedUniqueValues().keys()).sort(),
    [column.getFacetedUniqueValues()]
  );

  function getLabel(columnId: string, filter: string | boolean): string {
    if (columnId == 'exclude') {
      return filter ? 'Exclude' : 'Include';
    }
    if (columnId == 'transformer') {
      const t = transformers.find((t) => t.value == filter);
      return t?.title || '';
    }
    if (typeof filter === 'string') {
      return filter;
    }
    return '';
  }
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-[200px] justify-between"
        >
          <p className="truncate ...">
            {columnFilterValue && columnFilterValue.length
              ? columnFilterValue.join(', ')
              : 'Filter...'}
          </p>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search filters..." />
          <CommandEmpty>No filters found.</CommandEmpty>
          <CommandGroup>
            {sortedUniqueValues.map((i) => (
              <CommandItem
                key={i}
                onSelect={(currentValue) => {
                  if (columnFilterValue.includes(currentValue)) {
                    column.setFilterValue(
                      columnFilterValue.filter((v) => v != currentValue)
                    );
                  } else {
                    column.setFilterValue([...columnFilterValue, currentValue]);
                  }

                  setOpen(false);
                }}
                value={i}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    columnFilterValue.includes(i) ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {getLabel(column.id, i)}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
