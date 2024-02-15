'use client';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Button } from '@/components/ui/button';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import {
  ArrowDownIcon,
  ArrowUpIcon,
  CaretSortIcon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
import { ColumnDef, FilterFn } from '@tanstack/react-table';
import { HTMLProps, useEffect, useRef } from 'react';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { Row } from './SchemaPageTable';
import TransformerSelect from './TransformerSelect';

interface Props {
  transformers: Transformer[];
}

export function getSchemaColumns(props: Props): ColumnDef<Row>[] {
  const { transformers } = props;

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
            }}
          />
        </div>
      ),
      enableSorting: false,
      size: 30,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="-ml-3 h-8 data-[state=open]:bg-accent hover:border hover:border-gray-400"
        >
          <span>{'Schema'}</span>
          {column.getIsSorted() === 'desc' ? (
            <ArrowDownIcon className="ml-2 h-4 w-4" />
          ) : column.getIsSorted() === 'asc' ? (
            <ArrowUpIcon className="ml-2 h-4 w-4" />
          ) : (
            <CaretSortIcon className="ml-2 h-4 w-4" />
          )}
        </Button>
      ),
      filterFn: exactMatchFilterFn, //handles the multi-select on the schema drop down
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('schema')}
          </span>
        );
      },
      size: 200,
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Table" />
      ),
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('table')}
          </span>
        );
      },
      size: 200,
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Column" />
      ),
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('column')}
          </span>
        );
      },
      size: 200,
      maxSize: 200,
    },
    {
      accessorKey: 'dataType',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('dataType')}
          </span>
        );
      },
      size: 200,
    },
    {
      accessorKey: 'transformer',
      id: 'transformer',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Transformer" />
      ),
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
                                <div>{fieldState.error.message}</div>
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
                            side={'left'}
                          />
                        </div>
                        <EditTransformerOptions
                          transformer={transformers.find((t) => {
                            if (!fv) {
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
      size: 300,
    },
  ];
}

function IndeterminateCheckbox({
  indeterminate,
  className = 'w-4 h-4 mt-2',
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

/* Custom filter function that does an exact match. The out of the box filter function -arrIncludeSome- matches unnecessary elements. If you filtered a schema by a value  - customer_1, it matches customer_1, customer_10, customer_11, etc. The underlying implementation is:

*******
const arrIncludesSome: FilterFn<any> = (
  row,
  columnId: string,
  filterValue: unknown[]
) => {
  return filterValue.some(
    val => row.getValue<unknown[]>(columnId)?.includes(val)
  )
}

arrIncludesSome.autoRemove = (val: any) => testFalsey(val) || !val?.length

*******

This filter function does an exact match to avoid unnecessary values. 
*/
// eslint-disable-next-line
const exactMatchFilterFn: FilterFn<any> = (
  row,
  columnId: string,
  filterValue: unknown[]
) => {
  // Ensure the filter value and row value are exactly the same
  const rowValue = row.getValue(columnId);
  return filterValue.includes(rowValue); // This checks for an exact match in the filterValue array
};
