import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import {
  StickyHeaderTable,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { cn } from '@/libs/utils';
import {
  ArrowDownIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ArrowUpIcon,
  CaretSortIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
  EyeNoneIcon,
} from '@radix-ui/react-icons';
import {
  Column,
  ColumnDef,
  OnChangeFn,
  RowSelectionState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { useVirtualizer } from '@tanstack/react-virtual';
import {
  HTMLProps,
  ReactElement,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';

interface Option {
  value: string;
  label: string;
}
export type Action = 'add' | 'add-all' | 'remove' | 'remove-all';
interface Props {
  options: Option[];
  selected: Set<string>;
  onChange(value: Set<string>, action: Action): void;
}

export default function DualListBox(props: Props): ReactElement {
  const { options, selected, onChange } = props;

  const [leftSelected, setLeftSelected] = useState<RowSelectionState>({});
  const [rightSelected, setRightSelected] = useState<RowSelectionState>({});

  const cols = useMemo(() => getListBoxColumns({}), []);
  const leftData = options
    .filter((value) => !selected.has(value.value))
    .map((value): ListBoxRow => ({ table: value.value }));
  const rightData = options
    .filter((value) => selected.has(value.value))
    .map((value): ListBoxRow => ({ table: value.value }));

  return (
    <div className="flex gap-3 flex-row">
      <div className="flex">
        <ListBox
          columns={cols}
          data={leftData}
          onRowSelectionChange={setLeftSelected}
          rowSelection={leftSelected}
        />
      </div>
      <div className="flex flex-col justify-center gap-2">
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(
                new Set(options.map((option) => option.value)),
                'add-all'
              );
              setLeftSelected({});
            }}
          >
            <DoubleArrowRightIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              const newSet = new Set(selected);
              Object.entries(leftSelected).forEach(([key, isSelected]) => {
                if (isSelected) {
                  newSet.add(leftData[parseInt(key, 10)].table);
                }
              });
              onChange(newSet, 'add');
              setLeftSelected({});
            }}
          >
            <ArrowRightIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              const newSet = new Set(selected);
              Object.entries(rightSelected).forEach(([key, isSelected]) => {
                if (isSelected) {
                  newSet.delete(rightData[parseInt(key, 10)].table);
                }
              });
              onChange(newSet, 'remove');
              setRightSelected({});
            }}
          >
            <ArrowLeftIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(new Set(), 'remove-all');
              setRightSelected({});
            }}
          >
            <DoubleArrowLeftIcon />
          </Button>
        </div>
      </div>
      <div className="flex">
        <ListBox
          columns={cols}
          data={rightData}
          onRowSelectionChange={setRightSelected}
          rowSelection={rightSelected}
        />
      </div>
    </div>
  );
}

interface ListBoxRow {
  table: string;
}

interface ListBoxColumnProps {}

function getListBoxColumns(props: ListBoxColumnProps): ColumnDef<ListBoxRow>[] {
  const {} = props;
  return [
    {
      accessorKey: 'isSelected',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
          }}
        />
      ),
      cell: ({ row }) => (
        <div>
          <IndeterminateCheckbox
            {...{
              checked: row.getIsSelected(),
              indeterminate: row.getIsSomeSelected(),
              onChange: row.getToggleSelectedHandler(),
              id: row.getValue('table'),
            }}
          />
        </div>
      ),
      size: 30,
    },
    {
      accessorKey: 'table',
      header: ({ column }) => <ColumnHeader column={column} title="Table" />,
      cell: ({ row }) => {
        return (
          <label
            htmlFor={row.getValue('table')}
            className="max-w-[500px] truncate font-medium cursor-pointer"
          >
            {row.getValue('table')}
          </label>
        );
      },
    },
  ];
}

function IndeterminateCheckbox({
  indeterminate,
  className = 'w-4 h-4',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = useRef<HTMLInputElement>(null!);

  useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate, rest.checked]);

  return (
    <input
      type="checkbox"
      ref={ref}
      className={className + ' cursor-pointer '}
      {...rest}
    />
  );
}

interface DataTableColumnHeaderProps<TData, TValue>
  extends React.HTMLAttributes<HTMLDivElement> {
  column: Column<TData, TValue>;
  title: string;
}

function ColumnHeader<TData, TValue>({
  column,
  title,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  if (!column.getCanSort()) {
    return <span className={cn(className, 'text-xs')}>{title}</span>;
  }
  return (
    <div className={cn('flex items-center space-x-2', className)}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="-ml-3 data-[state=open]:bg-accent hover:border hover:border-gray-400 text-nowrap"
          >
            <span>{title}</span>
            {column.getIsSorted() === 'desc' ? (
              <ArrowDownIcon className="ml-2 h-4 w-4" />
            ) : column.getIsSorted() === 'asc' ? (
              <ArrowUpIcon className="ml-2 h-4 w-4" />
            ) : (
              <CaretSortIcon className="ml-2 h-4 w-4" />
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start">
          <Input
            type="text"
            value={(column.getFilterValue() ?? '') as string}
            onChange={(e) => column.setFilterValue(e.target.value)}
            placeholder={`Search...`}
            className="w-36 border rounded"
          />
          <DropdownMenuItem onClick={() => column.toggleSorting(false)}>
            <ArrowUpIcon className="mr-2 h-3.5 w-3.5 text-muted-foreground/70" />
            Asc
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => column.toggleSorting(true)}>
            <ArrowDownIcon className="mr-2 h-3.5 w-3.5 text-muted-foreground/70" />
            Desc
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => column.toggleVisibility(false)}>
            <EyeNoneIcon className="mr-2 h-3.5 w-3.5 text-muted-foreground/70" />
            Hide
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

interface ListBoxProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  rowSelection: RowSelectionState;
  onRowSelectionChange: OnChangeFn<RowSelectionState>;
}

function ListBox<TData, TValue>(
  props: ListBoxProps<TData, TValue>
): ReactElement {
  const { columns, data, rowSelection, onRowSelectionChange } = props;
  const table = useReactTable({
    data,
    columns,
    state: {
      rowSelection: rowSelection,
    },
    enableRowSelection: true,
    onRowSelectionChange: onRowSelectionChange,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    // getFacetedUniqueValues: getFacetedUniqueValues(),
    // getFacetedMinMaxValues: getFacetedMinMaxValues(),
    // enableMultiRowSelection: true,
  });
  const { rows } = table.getRowModel();
  const tableContainerRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33,
    getScrollElement: () => tableContainerRef.current,
    measureElement:
      typeof window !== 'undefined' &&
      navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined,
    overscan: 5,
  });

  return (
    <div className="w-full" ref={tableContainerRef}>
      <StickyHeaderTable>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow
              key={headerGroup.id}
              className="flex items-center flex-row w-full"
            >
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    className="flex items-center"
                    key={header.id}
                    style={{ minWidth: `${header.column.getSize()}px` }}
                    colSpan={header.colSpan}
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                );
              })}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`, // tells scrollbar how big the table is
          }}
          className="relative grid"
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const row = rows[virtualRow.index];
            return (
              <TableRow
                data-index={virtualRow.index} // needed for dynamic row height measurement
                ref={(node) => rowVirtualizer.measureElement(node)} // measure dynamic row height
                key={row.id}
                style={{
                  transform: `translateY(${virtualRow.start}px)`,
                }}
                className="items-center flex absolute w-full"
              >
                {row.getVisibleCells().map((cell) => {
                  return (
                    <TableCell
                      className="px-0"
                      key={cell.id}
                      style={{
                        minWidth: cell.column.getSize(),
                      }}
                    >
                      <div className="truncate">
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </div>
                    </TableCell>
                  );
                })}
              </TableRow>
            );
          })}
        </TableBody>
      </StickyHeaderTable>
    </div>
  );
}
