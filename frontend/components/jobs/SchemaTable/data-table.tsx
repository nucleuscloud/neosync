'use client';

import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { JobMappingFormValues } from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { DataTablePagination } from './data-table-pagination';
import { DataTableToolbar } from './data-table-toolbar';

interface DataTableProps {
  columns: ColumnDef<PlainMessage<DatabaseColumn>>[];
  data: JobMappingFormValues[];
  transformers?: Transformer[];
}

interface FilterItem {
  value: string;
  label: string;
}

export function DataTable({ columns, data, transformers }: DataTableProps) {
  const [rowSelection, setRowSelection] = React.useState({});
  const [columnVisibility, setColumnVisibility] =
    React.useState<VisibilityState>({});
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onColumnVisibilityChange: setColumnVisibility,
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  });

  function toFilterItems(items: Record<string, string>): FilterItem[] {
    return Object.keys(items)
      .sort()
      .map((d) => {
        return { value: d, label: d };
      });
  }

  function getFilterItems(
    colFilters: ColumnFiltersState
  ): Record<string, FilterItem[]> {
    console.log(JSON.stringify(colFilters));
    const setMap: Record<string, Record<string, string>> = {
      schema: {},
      table: {},
      column: {},
      dataType: {},
      transformer: {},
    };

    data.forEach((row) => {
      var shouldAdd = true;
      for (const [_, value] of Object.entries(colFilters)) {
        if (row[value.id as keyof JobMappingFormValues] != value.value) {
          shouldAdd = false;
          break;
        }
      }

      if (shouldAdd) {
        setMap.schema[row.schema] = row.schema;
        setMap.table[row.table] = row.table;
        setMap.column[row.column] = row.column;
        setMap.dataType[row.dataType] = row.dataType;
        setMap.transformer[row.transformer] = row.transformer;
      }
    });

    const uniqueTransformers = Object.keys(setMap.transformer);
    const filtersMap: Record<string, FilterItem[]> = {
      exclude: [
        { value: 'include', label: 'Include' },
        { value: 'exclude', label: 'Exclude' },
      ],
      transformer:
        transformers
          ?.filter((t) => uniqueTransformers.includes(t.value))
          .map((t) => {
            return { value: t.value, label: t.title };
          }) || [],
      schema: toFilterItems(setMap.schema),
      table: toFilterItems(setMap.table),
      column: toFilterItems(setMap.column),
      dataType: toFilterItems(setMap.dataType),
    };
    return filtersMap;
  }

  const [filterItems, setFilterItems] = React.useState<
    Record<string, FilterItem[]>
  >(getFilterItems([]));

  if (!data) {
    return <SkeletonTable />;
  }

  return (
    <div className="space-y-4">
      <DataTableToolbar table={table} transformers={transformers} />
      <div className="rounded-md border">
        <Table>
          <TableHeader>
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
                            setFilterValue={(value: string) => {
                              header.column.setFilterValue(value);
                              setFilterItems(
                                getFilterItems([
                                  ...columnFilters,
                                  { id: header.column.id, value },
                                ])
                              );
                              // if (value == '') {
                              //   setFilterItems(
                              //     getFilterItems([...columnFilters])
                              //   );
                              // } else {
                              //   setFilterItems(
                              //     getFilterItems([
                              //       ...columnFilters,
                              //       { id: header.column.id, value },
                              //     ])
                              //   );
                              // }
                            }}
                            items={
                              (filterItems && filterItems[header.column.id]) ||
                              []
                            }
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
      </div>
      <DataTablePagination table={table} />
    </div>
  );
}

interface FilterSelectProps {
  setFilterValue: (value: string) => void;
  items: FilterItem[];
}

function FilterSelect(props: FilterSelectProps) {
  const { items, setFilterValue } = props;
  const [open, setOpen] = React.useState(false);
  const [value, setValue] = React.useState('');
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-[200px] justify-between"
        >
          {value ? items.find((i) => i.value === value)?.label : 'Filter...'}
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder="Search framework..." />
          <CommandEmpty>No filters found.</CommandEmpty>
          <CommandGroup>
            {items.map((i) => (
              <CommandItem
                key={i.value}
                onSelect={(currentValue) => {
                  const newValue = currentValue === value ? '' : currentValue;
                  setValue(newValue);
                  setFilterValue(newValue);
                  setOpen(false);
                }}
                value={i.value}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    value === i.value ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {i.label}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
