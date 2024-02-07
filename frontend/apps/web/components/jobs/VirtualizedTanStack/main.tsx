import React, { HTMLProps, ReactElement } from 'react';

import './index.css';

import {
  Column,
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFacetedMinMaxValues,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  Table,
  useReactTable,
} from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import {
  isSystemTransformer,
  isUserDefinedTransformer,
  Transformer,
} from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { useVirtualizer } from '@tanstack/react-virtual';

import TransformerSelect from '../SchemaTable/TransformerSelect';
import { JobMapRow, makeData } from './makeData';

interface Props {
  transformers: Transformer[];
}

export default function SchemaTableTest({ transformers }: Props): ReactElement {
  const columns = React.useMemo<ColumnDef<JobMapRow>[]>(
    () => [
      // {
      //   accessorKey: 'schema',
      //   header: ({ table }) => (
      //     <>
      //       <IndeterminateCheckbox
      //         {...{
      //           checked: table.getIsAllRowsSelected(),
      //           indeterminate: table.getIsSomeRowsSelected(),
      //           onChange: table.getToggleAllRowsSelectedHandler(),
      //         }}
      //       />
      //       Schema
      //     </>
      //   ),
      //   cell: ({ row, getValue }) => (
      //     <div
      //     // style={{
      //     //   // Since rows are flattened by default,
      //     //   // we can use the row.depth property
      //     //   // and paddingLeft to visually indicate the depth
      //     //   // of the row
      //     //   paddingLeft: `${row.depth * 2}rem`,
      //     // }}
      //     >
      //       <IndeterminateCheckbox
      //         {...{
      //           checked: row.getIsSelected(),
      //           indeterminate: row.getIsSomeSelected(),
      //           onChange: row.getToggleSelectedHandler(),
      //         }}
      //       />
      //       <>{getValue()}</>
      //     </div>
      //   ),
      //   enableSorting: false,
      //   size: 60,
      // },
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
              }}
            />
          </div>
        ),
        enableSorting: false,
        size: 60,
      },
      {
        accessorKey: 'schema',
        header: 'Schema',
        // cell: (info) => info.getValue(),
        cell: ({ row }) => {
          return (
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('schema')}
            </span>
          );
        },
        size: 260,
      },
      {
        accessorKey: 'table',
        header: 'Table',
        // cell: (info) => info.getValue(),
        cell: ({ row }) => {
          return (
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('table')}
            </span>
          );
        },
        size: 260,
      },
      {
        accessorKey: 'column',
        header: 'Column',
        // cell: (info) => info.getValue(),
        cell: ({ row }) => {
          return (
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('column')}
            </span>
          );
        },
        size: 260,
      },
      {
        accessorKey: 'dataType',
        header: 'Data Type',
        // cell: (info) => info.getValue(),
        cell: ({ row }) => {
          return (
            <span className="max-w-[500px] truncate font-medium">
              {row.getValue('dataType')}
            </span>
          );
        },
        size: 160,
      },
      {
        accessorKey: 'transformer',
        header: 'Transformer',
        cell: (info) => {
          return (
            <div>
              <FormField<SchemaFormValues | SingleTableSchemaFormValues>
                name={`mappings.${info.row.original.formIdx}.transformer`}
                render={({ field, fieldState, formState }) => {
                  const fv = field.value as JobMappingTransformerForm;
                  return (
                    <FormItem>
                      <FormControl>
                        <div className="flex flex-row space-x-2">
                          {formState.errors.mappings && (
                            <div className="place-self-center">
                              {fieldState.error ? (
                                <div>
                                  <ExclamationTriangleIcon className="h-4 w-4 text-destructive" />
                                </div>
                              ) : (
                                <div className="w-4"></div>
                              )}
                            </div>
                          )}

                          <div>
                            <TransformerSelect
                              transformers={transformers}
                              value={fv}
                              onSelect={field.onChange}
                              placeholder="Select Transformer..."
                            />
                          </div>
                          <EditTransformerOptions
                            transformer={transformers.find((t) => {
                              if (!fv) {
                                console.log('mjrj');
                                return;
                              }
                              if (
                                fv.source === 'custom' &&
                                fv.config.case ===
                                  'userDefinedTransformerConfig' &&
                                isUserDefinedTransformer(t) &&
                                t.id === fv.config.value.id
                              ) {
                                return t;
                              }
                              return (
                                isSystemTransformer(t) && t.source === fv.source
                              );
                            })}
                            index={info.row.original.formIdx}
                          />
                        </div>
                      </FormControl>
                    </FormItem>
                  );
                }}
              />
            </div>
          );
        },
        size: 160,
      },
    ],
    []
  );

  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>(
    []
  );
  const [data, _setData] = React.useState(() => makeData(10000, 10, 100));

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
    },
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getFacetedMinMaxValues: getFacetedMinMaxValues(),
    debugTable: true,
  });

  const { rows } = table.getRowModel();

  //The virtualizer needs to know the scrollable container elementS
  const tableContainerRef = React.useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: rows.length,
    estimateSize: () => 33, //estimate row height for accurate scrollbar dragging
    getScrollElement: () => tableContainerRef.current,
    //measure dynamic row height, except in firefox because it measures table border height incorrectly
    measureElement:
      typeof window !== 'undefined' &&
      navigator.userAgent.indexOf('Firefox') === -1
        ? (element) => element?.getBoundingClientRect().height
        : undefined,
    overscan: 5,
  });

  //All important CSS styles are included as inline styles for this example. This is not recommended for your code.
  return (
    <div className="app">
      ({data.length} rows)
      <div
        className="container //rounded-md border"
        ref={tableContainerRef}
        style={{
          overflow: 'auto', //our scrollable table container
          position: 'relative', //needed for sticky header
          height: '500px', //should be a fixed height
          width: '100%',
        }}
      >
        {/* Even though we're still using sematic table tags, we must use CSS grid and flexbox for dynamic row heights */}
        <table style={{ display: 'grid' }}>
          <thead
            style={{
              display: 'grid',
              position: 'sticky',
              top: 0,
              zIndex: 1,
            }}
            className="bg-muted pb-4 rounded-md"
          >
            {table.getHeaderGroups().map((headerGroup) => (
              <tr
                key={headerGroup.id}
                style={{ display: 'flex', width: '100%' }}
              >
                {headerGroup.headers.map((header) => {
                  return (
                    <th
                      key={header.id}
                      style={{
                        display: 'flex',
                        width: header.getSize(),
                      }}
                    >
                      <div
                        {...{
                          className: header.column.getCanSort()
                            ? 'cursor-pointer select-none'
                            : '',
                          onClick: header.column.getToggleSortingHandler(),
                        }}
                      >
                        {flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                        {{
                          asc: ' 🔼',
                          desc: ' 🔽',
                        }[header.column.getIsSorted() as string] ?? null}
                        {header.column.getCanFilter() ? (
                          <div>
                            <Filter column={header.column} table={table} />
                          </div>
                        ) : null}
                      </div>
                    </th>
                  );
                })}
              </tr>
            ))}
          </thead>
          <tbody
            style={{
              display: 'grid',
              height: `${rowVirtualizer.getTotalSize()}px`, //tells scrollbar how big the table is
              position: 'relative', //needed for absolute positioning of rows
            }}
          >
            {rowVirtualizer.getVirtualItems().map((virtualRow) => {
              const row = rows[virtualRow.index] as Row<JobMapRow>;
              return (
                <tr
                  data-index={virtualRow.index} //needed for dynamic row height measurement
                  ref={(node) => rowVirtualizer.measureElement(node)} //measure dynamic row height
                  key={row.id}
                  style={{
                    display: 'flex',
                    position: 'absolute',
                    transform: `translateY(${virtualRow.start}px)`, //this should always be a `style` as it changes on scroll
                  }}
                >
                  {row.getVisibleCells().map((cell) => {
                    return (
                      <td
                        key={cell.id}
                        style={{
                          display: 'flex',
                          width: cell.column.getSize(),
                        }}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </td>
                    );
                  })}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function IndeterminateCheckbox({
  indeterminate,
  className = '',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = React.useRef<HTMLInputElement>(null!);

  React.useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate]);

  return (
    <input
      type="checkbox"
      ref={ref}
      className={className + ' cursor-pointer mr-4'}
      {...rest}
    />
  );
}

function Filter({
  column,
  table,
}: {
  //eslint-disable-next-line
  column: Column<any, unknown>;
  //eslint-disable-next-line
  table: Table<any>;
}) {
  const firstValue = table
    .getPreFilteredRowModel()
    .flatRows[0]?.getValue(column.id);

  const columnFilterValue = column.getFilterValue();

  const sortedUniqueValues = React.useMemo(
    () =>
      typeof firstValue === 'number'
        ? []
        : Array.from(column.getFacetedUniqueValues().keys()).sort(),
    [column.getFacetedUniqueValues()]
  );

  return typeof firstValue === 'number' ? (
    <div>
      {/* <div style={{display:'flexzclassName="flex space-x-2"> */}
      <div style={{ display: 'flex' }}>
        <DebouncedInput
          type="number"
          min={Number(column.getFacetedMinMaxValues()?.[0] ?? '')}
          max={Number(column.getFacetedMinMaxValues()?.[1] ?? '')}
          value={(columnFilterValue as [number, number])?.[0] ?? ''}
          onChange={(value) =>
            column.setFilterValue((old: [number, number]) => [value, old?.[1]])
          }
          placeholder={`Min ${
            column.getFacetedMinMaxValues()?.[0]
              ? `(${column.getFacetedMinMaxValues()?.[0]})`
              : ''
          }`}
          // className="w-24 border shadow rounded"
        />
        <DebouncedInput
          type="number"
          min={Number(column.getFacetedMinMaxValues()?.[0] ?? '')}
          max={Number(column.getFacetedMinMaxValues()?.[1] ?? '')}
          value={(columnFilterValue as [number, number])?.[1] ?? ''}
          onChange={(value) =>
            column.setFilterValue((old: [number, number]) => [old?.[0], value])
          }
          placeholder={`Max ${
            column.getFacetedMinMaxValues()?.[1]
              ? `(${column.getFacetedMinMaxValues()?.[1]})`
              : ''
          }`}
          // className="w-24 border shadow rounded"
          style={{
            width: '96px', // Equivalent to w-24
            border: '1px solid', // Default border, you might need to adjust color
            boxShadow: '0px 1px 2px rgba(0, 0, 0, 0.1)', // This is a general shadow, adjust as needed
            borderRadius: '0.25rem', // Default rounding, adjust as needed
          }}
        />
      </div>
      <div style={{ height: '4px' }} />
    </div>
  ) : (
    <>
      <datalist id={column.id + 'list' + table + firstValue}>
        {sortedUniqueValues.slice(0, 5000).map((value: any) => (
          <option value={value} key={value} />
        ))}
      </datalist>
      <DebouncedInput
        type="text"
        value={(columnFilterValue ?? '') as string}
        onChange={(value) => column.setFilterValue(value)}
        placeholder={`Search... (${column.getFacetedUniqueValues().size})`}
        // className="w-36 border shadow rounded"
        list={column.id + 'list'}
      />
      {/* <div className="h-1" /> */}
    </>
  );
}

// A debounced input react component
function DebouncedInput({
  value: initialValue,
  onChange,
  debounce = 500,
  ...props
}: {
  value: string | number;
  onChange: (value: string | number) => void;
  debounce?: number;
} & Omit<React.InputHTMLAttributes<HTMLInputElement>, 'onChange'>) {
  const [value, setValue] = React.useState(initialValue);

  React.useEffect(() => {
    setValue(initialValue);
  }, [initialValue]);

  React.useEffect(() => {
    const timeout = setTimeout(() => {
      onChange(value);
    }, debounce);

    return () => clearTimeout(timeout);
  }, [value]);

  return (
    <input
      {...props}
      value={value}
      onChange={(e) => setValue(e.target.value)}
    />
  );
}
