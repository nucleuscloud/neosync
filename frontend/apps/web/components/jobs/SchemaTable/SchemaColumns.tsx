'use client';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import { ColumnMetadata } from '@/app/(mgmt)/[account]/new/job/schema/page';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Badge } from '@/components/ui/badge';
import { FormControl, FormField, FormItem } from '@/components/ui/form';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { ForeignConstraint } from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef, FilterFn, Row, SortingFn } from '@tanstack/react-table';
import { HTMLProps, useEffect, useRef } from 'react';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { Row as RowData } from './SchemaPageTable';
import TransformerSelect from './TransformerSelect';

interface Props {
  transformers: Transformer[];
  columnMetadata: ColumnMetadata;
}

export function getSchemaColumns(props: Props): ColumnDef<RowData>[] {
  const { transformers, columnMetadata } = props;

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
      maxSize: 30,
      minSize: 30,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Schema" />
      ),
      filterFn: exactMatchFilterFn, //handles the multi-select on the schema drop down
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {row.getValue('schema')}
          </span>
        );
      },
    },
    {
      accessorKey: 'table',
      filterFn: exactMatchFilterFn,
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
    },
    {
      id: 'constraints',
      accessorKey: 'constraints',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Constraints" />
      ),
      filterFn: filterConstraints(columnMetadata),
      sortingFn: sortConstraints(columnMetadata),
      meta: columnMetadata,
      cell: ({ row }) => {
        const rowKey = `${row.getValue('schema')}.${row.getValue('table')}`;

        const hasForeignKeyConstraint =
          columnMetadata?.fk &&
          columnMetadata?.fk[rowKey]?.constraints.filter(
            (item: ForeignConstraint) => item.column == row.getValue('column')
          ).length > 0;

        const foreignKeyConstraint = {
          table:
            columnMetadata?.fk[rowKey]?.constraints[0].foreignKey?.table ?? '', // the foreignKey constraints object comes back from the API with two identical objects in an array, so just getting the first one. Need to investigate why it returns two.
          column:
            columnMetadata?.fk[rowKey]?.constraints[0].foreignKey?.column ?? '',
          value: 'Foreign Key',
        };

        return (
          <span className="max-w-[500px] truncate font-medium">
            <div className="flex flex-col lg:flex-row items-start gap-1">
              <div>
                {columnMetadata?.pk[rowKey]?.columns.includes(
                  row.getValue('column')
                ) && (
                  <Badge
                    variant="outline"
                    className="text-xs bg-blue-100 text-gray-800 cursor-default dark:bg-blue-200 dark:text-gray-900"
                  >
                    Primary Key
                  </Badge>
                )}
              </div>
              <div>
                {hasForeignKeyConstraint && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge
                          variant="outline"
                          className="text-xs bg-orange-100 text-gray-800 dark:dark:text-gray-900 dark:bg-orange-300"
                        >
                          {foreignKeyConstraint.value}
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        {`Primary Key: ${foreignKeyConstraint.table}.${foreignKeyConstraint.column}`}
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
              </div>
            </div>
          </span>
        );
      },
    },
    {
      accessorKey: 'isNullable',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Nullable" />
      ),
      cell: ({ row }) => {
        let isNullable = '';
        columnMetadata?.isNullable.map((item) => {
          if (
            item.schema == row.getValue('schema') &&
            item.table == row.getValue('table') &&
            item.column == row.getValue('column')
          ) {
            isNullable = item.isNullable;
          }
        });

        // fallback if it's empty for some reason then we should just default to a safe answer
        if (isNullable == '') {
          isNullable = 'No';
        }

        const toTitleCase = (s: string) => {
          if (!s) return '';
          const firstLetter = s[0].toUpperCase();
          const restOfString = s.slice(1).toLowerCase();
          return firstLetter + restOfString;
        };

        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">{toTitleCase(isNullable)}</Badge>
          </span>
        );
      },
    },
    {
      accessorKey: 'dataType',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">
              {handleDataTypeBadge(row.getValue('dataType'))}
            </Badge>
          </span>
        );
      },
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
      size: 250,
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
  // ensure the filter value and row value are exactly the same
  const rowValue = row.getValue(columnId);
  return filterValue.includes(rowValue); // this checks for an exact match in the filterValue array
};

// cleans up the data type values since some are too long , can add on more here as we
function handleDataTypeBadge(dataType: string): string {
  const splitDt = dataType.split('(');
  switch (splitDt[0]) {
    case 'character varying':
      return 'varchar(' + splitDt[1];
    default:
      return dataType;
  }
}

const buildRowKey = (row: Row<RowData>): string =>
  `${row.getValue('schema')}.${row.getValue('table')}`;

// SortinFn can only take in 3 params: rowA, rowB and id, so creating a closure to get access to the metadata
function sortConstraints(meta: ColumnMetadata): SortingFn<RowData> {
  return (rowA, rowB) => {
    // Check for primary key constraint presence
    const hasPrimaryKeyA =
      meta?.pk[buildRowKey(rowA)]?.columns.includes(rowA.getValue('column')) ??
      false;
    const hasPrimaryKeyB =
      meta?.pk[buildRowKey(rowB)]?.columns.includes(rowB.getValue('column')) ??
      false;

    // check for foreign key constraint presence
    const hasForeignKeyA =
      meta?.fk[buildRowKey(rowA)]?.constraints.filter(
        (item: ForeignConstraint) => item.column == rowA.getValue('column')
      ).length > 0;
    const hasForeignKeyB =
      meta?.fk[buildRowKey(rowB)]?.constraints.filter(
        (item: ForeignConstraint) => item.column == rowB.getValue('column')
      ).length > 0;

    // can't have primary key and foreign key so figure out which one exists on any given row
    const valueA = hasPrimaryKeyA
      ? 'Primary Key'
      : hasForeignKeyA
        ? 'Foreign Key'
        : '';
    const valueB = hasPrimaryKeyB
      ? 'Primary Key'
      : hasForeignKeyB
        ? 'Foreign Key'
        : '';

    // prioritize "Primary Key", then "Foreign Key", then empty strings in sorting
    if (valueA === 'Primary Key' || valueB === 'Primary Key') {
      return valueA === 'Primary Key' ? -1 : 1;
    } else if (valueA === 'Foreign Key' || valueB === 'Foreign Key') {
      return valueA === 'Foreign Key' ? -1 : 1;
    } else if (valueA === '' && valueB !== '') {
      return 1;
    } else if (valueA !== '' && valueB === '') {
      return -1;
    }

    return valueA.localeCompare(valueB);
  };
}

// custom filter for the constraints column
function filterConstraints(meta: ColumnMetadata): FilterFn<RowData> {
  return (row, columnId, filterValue) => {
    const filterValueStr = filterValue.toString().toLowerCase();

    const isPrimaryKey =
      meta?.pk[buildRowKey(row)]?.columns.includes(row.getValue('column')) ??
      false;
    const isForeignKey =
      Object.keys(meta?.fk).includes(buildRowKey(row)) &&
      meta?.fk[buildRowKey(row)].constraints.some(
        (fk) => fk.column === row.getValue('column')
      );

    const columnConstraintType = isPrimaryKey
      ? 'Primary Key'
      : isForeignKey
        ? 'Foreign Key'
        : '';

    return columnConstraintType.toLowerCase().includes(filterValueStr);
  };
}
