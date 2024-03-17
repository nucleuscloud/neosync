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
import { ConnectionSchemaMap } from '@/libs/hooks/useGetConnectionSchemaMap';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  DatabaseColumn,
  ForeignConstraint,
  ForeignConstraintTables,
  ForeignKey,
  PrimaryConstraint,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef, FilterFn, Row, SortingFn } from '@tanstack/react-table';
import { HTMLProps, ReactElement, useEffect, useRef } from 'react';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { Row as RowData } from './SchemaPageTable';
import TransformerSelect from './TransformerSelect';
// import {
//   RankingInfo,
//   rankItem,
//   compareItems,
// } from '@tanstack/match-sorter-utils';

export interface ColumnKey {
  schema: string;
  table: string;
  column: string;
}

function fromColKey(key: ColumnKey): string {
  return `${key.schema}.${key.table}.${key.column}`;
}
function fromDbCol(dbcol: PlainMessage<DatabaseColumn>): string {
  return `${dbcol.schema}.${dbcol.table}.${dbcol.column}`;
}
function toColKey(schema: string, table: string, column: string): ColumnKey {
  return {
    schema,
    table,
    column,
  };
}

export interface SchemaConstraintHandler {
  getIsPrimaryKey(key: ColumnKey): boolean;
  getIsForeignKey(key: ColumnKey): [boolean, string[]];
  getIsNullable(key: ColumnKey): boolean;
  getDataType(key: ColumnKey): string;
  getIsInSchema(key: ColumnKey): boolean;
}

interface ColDetails {
  isPrimaryKey: boolean;
  fk: [boolean, string[]];
  isNullable: boolean;
  dataType: string;
}

export function getSchemaConstraintHandler(
  schema: ConnectionSchemaMap,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>
): SchemaConstraintHandler {
  const colmap = buildColDetailsMap(
    schema,
    primaryConstraints,
    foreignConstraints
  );
  return {
    getDataType(key) {
      return colmap[fromColKey(key)]?.dataType ?? '';
    },
    getIsForeignKey(key) {
      return colmap[fromColKey(key)]?.fk ?? [false, []];
    },
    getIsNullable(key) {
      return colmap[fromColKey(key)]?.isNullable ?? false;
    },
    getIsPrimaryKey(key) {
      return colmap[fromColKey(key)]?.isPrimaryKey ?? false;
    },
    getIsInSchema(key) {
      return !!colmap[fromColKey(key)];
    },
  };
}

function buildColDetailsMap(
  schema: ConnectionSchemaMap,
  primaryConstraints: Record<string, PrimaryConstraint>,
  foreignConstraints: Record<string, ForeignConstraintTables>
): Record<string, ColDetails> {
  const colmap: Record<string, ColDetails> = {};
  Object.entries(schema).forEach(([key, dbcols]) => {
    const tablePkeys = primaryConstraints[key] ?? new PrimaryConstraint();
    const primaryCols = new Set(tablePkeys.columns);
    const foreignFkeys =
      foreignConstraints[key] ?? new ForeignConstraintTables();
    const fkConstraints = foreignFkeys.constraints;
    const fkconstraintsMap: Record<string, ForeignKey> = {};
    fkConstraints.forEach((constraint) => {
      fkconstraintsMap[constraint.column] =
        constraint.foreignKey ?? new ForeignKey();
    });

    dbcols.forEach((dbcol) => {
      const fk: ForeignKey | undefined = fkconstraintsMap[dbcol.column];
      colmap[fromDbCol(dbcol)] = {
        isNullable: dbcol.isNullable === 'YES',
        dataType: dbcol.dataType,
        fk: [
          fk !== undefined,
          fk !== undefined ? [`${fk.table}.${fk.column}`] : [],
        ],
        isPrimaryKey: primaryCols.has(dbcol.column),
        isInSchema: true,
      };
    });
  });
  return colmap;
}

interface RowAlertProps {
  row: Row<RowData>;
  handler: SchemaConstraintHandler;
}

function RowAlert(props: RowAlertProps): ReactElement {
  const { row, handler } = props;
  const key: ColumnKey = {
    schema: row.getValue('schema'),
    table: row.getValue('table'),
    column: row.getValue('column'),
  };
  const isInSchema = handler.getIsInSchema(key);

  const messages: string[] = [];

  if (!isInSchema) {
    messages.push('This column was not found in the backing source schema');
  }

  if (messages.length === 0) {
    return <div />;
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div className="cursor-default">
            <ExclamationTriangleIcon className="text-yellow-600 dark:text-yellow-300" />
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{messages.join('\n')}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}

interface Props {
  transformers: Transformer[];
  constraintHandler: SchemaConstraintHandler;
}

export function getSchemaColumns(props: Props): ColumnDef<RowData>[] {
  const { transformers, constraintHandler } = props;

  // const fc = useFormContext();

  // const columnHelper = createColumnHelper<RowData>();
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
      id: 'alert',
      size: 1,
      cell: ({ row }) => <RowAlert row={row} handler={constraintHandler} />,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Schema" />
      ),
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
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'schemaTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Table" />
      ),
      cell: ({ getValue }) => {
        return (
          <span className="max-w-[500px] truncate font-medium">
            {getValue() as string}
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
      // filterFn: filterConstraints(columnMetadata),
      // sortingFn: sortConstraints(columnMetadata),
      // meta: columnMetadata,
      cell: ({ row }) => {
        const key: ColumnKey = {
          schema: row.getValue('schema'),
          table: row.getValue('table'),
          column: row.getValue('column'),
        };
        const isPrimaryKey = constraintHandler.getIsPrimaryKey(key);
        const [isForeignKey, fkCols] = constraintHandler.getIsForeignKey(key);
        return (
          <span className="max-w-[500px] truncate font-medium">
            <div className="flex flex-col lg:flex-row items-start gap-1">
              <div>
                {isPrimaryKey && (
                  <Badge
                    variant="outline"
                    className="text-xs bg-blue-100 text-gray-800 cursor-default dark:bg-blue-200 dark:text-gray-900"
                  >
                    Primary Key
                  </Badge>
                )}
              </div>
              <div>
                {isForeignKey && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge
                          variant="outline"
                          className="text-xs bg-orange-100 text-gray-800 dark:dark:text-gray-900 dark:bg-orange-300"
                        >
                          Foreign Key
                        </Badge>
                      </TooltipTrigger>
                      <TooltipContent>
                        {fkCols.map((col) => `Primary Key: ${col}`).join('\n')}
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
        const key: ColumnKey = {
          schema: row.getValue('schema'),
          table: row.getValue('table'),
          column: row.getValue('column'),
        };
        const isNullable = constraintHandler.getIsNullable(key);
        const text = isNullable ? 'Yes' : 'No';
        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">{text}</Badge>
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
        const key: ColumnKey = {
          schema: row.getValue('schema'),
          table: row.getValue('table'),
          column: row.getValue('column'),
        };
        const datatype = constraintHandler.getDataType(key);
        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">{handleDataTypeBadge(datatype)}</Badge>
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
      // meta: columnMetadata,
      cell: (info) => {
        // const rowKey = `${info.row.getValue('schema')}.${info.row.getValue('table')}`;

        // const isForeignKeyConstraint =
        //   columnMetadata?.fk &&
        //   columnMetadata?.fk[rowKey]?.constraints.filter(
        //     (item: ForeignConstraint) =>
        //       item.column == info.row.getValue('column')
        //   ).length > 0;

        // let disableTransformer = false;

        // const foreignKeyConstraint = {
        //   table: columnMetadata?.fk[rowKey]?.constraints.find(
        //     (item) => item.column == info.row.getValue('column')
        //   )?.foreignKey?.table,
        //   column: columnMetadata?.fk[rowKey]?.constraints.find(
        //     (item) => item.column == info.row.getValue('column')
        //   )?.foreignKey?.column,
        //   value: 'Foreign Key',
        // };

        // if the current row is a foreignKey constraint, then check that it's primary key transformer
        // if (isForeignKeyConstraint) {
        //   disableTransformer =
        //     fc
        //       .getValues()
        //       .mappings.find(
        //         (item: JobMappingFormValues) =>
        //           item.schema + '.' + item.table ==
        //             foreignKeyConstraint.table &&
        //           item.column == foreignKeyConstraint.column
        //       ).transformer?.source !== 'passthrough';
        // }

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
                            disabled={false}
                            // disabled={disableTransformer} // todo
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
                          // disabled={disableTransformer}
                          disabled={false} // todo
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
      if (splitDt[1] == undefined) {
        return 'varchar(' + splitDt[1] + ')';
      } else {
        return 'varchar(' + splitDt[1];
      }
    case 'timestamp with time zone':
      return 'timestamp(timezone)';
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
