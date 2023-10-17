'use client';

import {
  Column,
  ColumnDef,
  ColumnFiltersState,
  SortingState,
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
import { CheckIcon } from '@radix-ui/react-icons';
import { AiOutlineFilter } from 'react-icons/ai';
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
  const [filtersUpdated, setFiltersUpdated] = React.useState(false);
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
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  });

  const [treeData, setTreeData] = React.useState<TreeDataItem[]>([]);

  function handlefilter(items: TreeDataItem[]) {
    const schemaFilters: string[] = [];
    const tableFilters: string[] = [];
    function walkTreeItems(items: TreeDataItem | TreeDataItem[]) {
      if (items instanceof Array) {
        // eslint-disable-next-line @typescript-eslint/prefer-for-of
        for (let i = 0; i < items.length; i++) {
          if (items[i].isSelected) {
            if (items[i].children) {
              schemaFilters.push(items[i]!.name);
            } else {
              tableFilters.push(items[i]!.name);
            }
          }
          if (walkTreeItems(items[i]!)) {
            return true;
          }
        }
      } else if (items.children) {
        return walkTreeItems(items.children);
      }
    }

    walkTreeItems(items);
    if (schemaFilters.length == 0 && tableFilters.length == 0) {
      return setColumnFilters([]);
    }
    setColumnFilters([
      { id: 'schema', value: schemaFilters },
      { id: 'table', value: tableFilters },
    ]);
  }

  function restoreTree(): void {
    const treedata = Object.keys(schemaMap).map((schema) => {
      const children = Object.keys(schemaMap[schema]).map((table) => {
        return {
          id: `${schema}.${table}`,
          name: table,
          isSelected: false,
        };
      });

      return {
        id: schema,
        name: schema,
        isSelected: false,
        children,
      };
    });
    setTreeData(treedata);
  }

  function isTableSelected(
    table: string,
    schema: string,
    tableFilters: string[],
    schemaFilters: string[]
  ): boolean {
    return (
      (schemaFilters.length === 0 || schemaFilters.includes(schema)) &&
      (tableFilters.length === 0 || tableFilters.includes(table))
    );
  }

  function getFiltersById(
    id: string,
    columnFilters: ColumnFiltersState
  ): string[] {
    return (columnFilters.find((f) => f.id == id)?.value as string[]) || [];
  }

  function updateTree(): void {
    const schemaFilters: string[] = getFiltersById('schema', columnFilters);
    const tableFilters: string[] = getFiltersById('table', columnFilters);

    const treedata = Object.keys(schemaMap).map((schema) => {
      const isSchemaSelected =
        columnFilters.length > 0 && schemaFilters.some((f) => f == schema);

      const tables = Object.keys(schemaMap[schema]).map((table) => {
        return {
          id: `${schema}-${table}`,
          name: table,
          isSelected: isTableSelected(
            table,
            schema,
            tableFilters,
            schemaFilters
          ),
        };
      });
      const isSomeTablesSelected = tables.some((t) => t.isSelected);

      return {
        id: schema,
        name: schema,
        isSelected: isSchemaSelected || isSomeTablesSelected,
        children: tables,
      };
    });
    setTreeData(treedata);
  }

  React.useEffect(() => {
    if (filtersUpdated) {
      if (columnFilters.length == 0) {
        restoreTree();
      } else {
        updateTree();
      }
    }
    setFiltersUpdated(false);
  }, [filtersUpdated]);

  React.useEffect(() => {
    restoreTree();
  }, [JSON.stringify(schemaMap)]);

  if (!data) {
    return <SkeletonTable />;
  }

  return (
    <div className="flex flex-row">
      <div className="basis-1/6 min-w-[170px] pt-[45px] ">
        <Tree
          data={treeData}
          className="border rounded-md"
          onSelectChange={handlefilter}
        />
      </div>
      <div className="basis-5/6 space-y-2 pl-8">
        <DataTableToolbar
          table={table}
          transformers={transformers}
          onClearFilters={() => {
            setFiltersUpdated(true);
          }}
        />
        <div className="rounded-md border overflow-hidden">
          <Table className="table-fixed">
            <TableHeader className="bg-gray-100">
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => {
                    return (
                      <TableHead
                        key={header.id}
                        className={
                          header.id == 'select'
                            ? 'w-[44px] pl-2 '
                            : 'w-[197px] px-0'
                        }
                      >
                        <div className="flex flex-row">
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
                                transformers={transformers || []}
                                onSelect={() => {
                                  setFiltersUpdated(true);
                                }}
                              />
                            </div>
                          ) : null}
                        </div>
                      </TableHead>
                    );
                  })}
                </TableRow>
              ))}
            </TableHeader>
          </Table>
          <ScrollArea className="max-h-[700px] overflow-y-auto">
            <Table className="table-fixed">
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow
                      key={row.id}
                      data-state={row.getIsSelected() && 'selected'}
                    >
                      {row.getVisibleCells().map((cell) => {
                        return (
                          <TableCell
                            className={
                              cell.column.id == 'select'
                                ? ' w-[40px]'
                                : 'w-[197px]'
                            }
                            key={cell.id}
                          >
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        );
                      })}
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
  );
}

interface FilterSelectProps<TData, TValue> {
  column: Column<TData, TValue>;
  transformers: Transformer[];
  onSelect: () => void;
}

function FilterSelect<TData, TValue>(props: FilterSelectProps<TData, TValue>) {
  const { column, transformers, onSelect } = props;
  const [open, setOpen] = React.useState(false);

  const columnFilterValue = (column.getFilterValue() as string[]) || [];

  const sortedUniqueValues = React.useMemo(
    () => Array.from(column.getFacetedUniqueValues().keys()).sort(),
    [column.getFacetedUniqueValues()]
  );

  function getLabel(columnId: string, filter: string | boolean): string {
    if (columnId == 'transformer') {
      const t = transformers.find((t) => t.value == filter);
      return t?.value || '';
    }
    if (typeof filter === 'string') {
      return filter;
    }
    return '';
  }

  function computeFilters(newValue: string, currentValues: string[]): string[] {
    if (currentValues.includes(newValue)) {
      return currentValues.filter((v) => v != newValue);
    }
    return [...currentValues, newValue];
  }
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          role="combobox"
          aria-expanded={open}
          className="hover:bg-gray-200 p-2"
        >
          <AiOutlineFilter />
          {columnFilterValue && columnFilterValue.length ? (
            <div
              id="notifbadge"
              className="bg-blue-500 w-[6px] h-[6px] text-white rounded-full text-[8px] relative top-[-8px] right-0"
            />
          ) : null}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="min-w-[175px] p-0">
        <Command>
          <CommandInput placeholder="Search filters..." />
          <CommandEmpty>No filters found.</CommandEmpty>
          <CommandGroup>
            {sortedUniqueValues.map((i, index) => (
              <CommandItem
                key={`${i}-${index}`}
                onSelect={(currentValue) => {
                  const newValues = computeFilters(
                    currentValue,
                    columnFilterValue
                  );
                  column.setFilterValue(newValues);
                  onSelect();
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
