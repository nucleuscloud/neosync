'use client';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
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
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  DatabaseColumn,
  ForeignConstraintTables,
  ForeignKey,
  PrimaryConstraint,
  SystemTransformer,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef, Row } from '@tanstack/react-table';
import { HTMLProps, ReactElement, useEffect, useRef } from 'react';
import { useFormContext } from 'react-hook-form';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { Row as RowData } from './SchemaPageTable';
import TransformerSelect from './TransformerSelect';

interface ColumnKey {
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
function fromRowDataToColKey(row: Row<RowData>): ColumnKey {
  return {
    schema: row.getValue('schema'),
    table: row.getValue('table'),
    column: row.getValue('column'),
  };
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
    return <div className="hidden" />;
  }

  return (
    <TooltipProvider delayDuration={100}>
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
  systemTransformers: SystemTransformer[];
  userDefinedTransformers: UserDefinedTransformer[];
  systemMap: Map<string, SystemTransformer>;
  userDefinedMap: Map<string, UserDefinedTransformer>;
  constraintHandler: SchemaConstraintHandler;
}

export function getSchemaColumns(props: Props): ColumnDef<RowData>[] {
  const {
    systemTransformers,
    userDefinedTransformers,
    systemMap,
    userDefinedMap,
    constraintHandler,
  } = props;

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
      enableHiding: false,
      size: 30,
    },
    {
      accessorKey: 'schema',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'table',
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorFn: (row) => `${row.schema}.${row.table}`,
      id: 'schemaTable',
      footer: (props) => props.column.id,
      header: ({ column }) => (
        <SchemaColumnHeader column={column} title="Table" />
      ),
      cell: ({ row, getValue }) => {
        return (
          <div className="flex flex-row gap-2 items-center">
            <RowAlert row={row} handler={constraintHandler} />
            <span className="max-w-[500px] truncate font-medium">
              {getValue() as string}
            </span>
          </div>
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
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        const isPrimaryKey = constraintHandler.getIsPrimaryKey(key);
        const [isForeignKey, fkCols] = constraintHandler.getIsForeignKey(key);

        const pieces: string[] = [];
        if (isPrimaryKey) {
          pieces.push('Primary Key');
        }
        if (isForeignKey) {
          fkCols.forEach((col) => pieces.push(`Foreign Key: ${col}`));
        }
        return pieces.join('\n');
      },
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
                  <TooltipProvider delayDuration={200}>
                    <Tooltip>
                      <TooltipTrigger>
                        <Badge
                          variant="outline"
                          className="text-xs bg-orange-100 text-gray-800 cursor-default dark:dark:text-gray-900 dark:bg-orange-300"
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
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        return constraintHandler.getIsNullable(key) ? 'Yes' : 'No';
      },
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
      accessorFn: (row) => {
        const key = toColKey(row.schema, row.table, row.column);
        return handleDataTypeBadge(constraintHandler.getDataType(key));
      },
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
      cell: (info) => {
        const colkey = fromRowDataToColKey(info.row);
        const [isForeignKey] = constraintHandler.getIsForeignKey(colkey);

        const fkSystemTransformers: SystemTransformer[] = [];
        if (isForeignKey) {
          const passthrough = systemMap.get('passthrough');
          if (passthrough) {
            fkSystemTransformers.push(passthrough);
          }
          if (constraintHandler.getIsNullable(colkey)) {
            const nullableTf = systemMap.get('null');
            if (nullableTf) {
              fkSystemTransformers.push(nullableTf);
            }
          }
        }

        const fkSystemTransformersMap = new Map(
          fkSystemTransformers.map((t) => [t.source, t])
        );

        const fctx = useFormContext<
          SchemaFormValues | SingleTableSchemaFormValues
        >();
        return (
          <div>
            <FormField<SchemaFormValues | SingleTableSchemaFormValues>
              name={`mappings.${info.row.index}.transformer`}
              control={fctx.control}
              render={({ field, fieldState, formState }) => {
                const fv = field.value as JobMappingTransformerForm;
                let transformer: Transformer | undefined;
                if (
                  fv.source === 'custom' &&
                  fv.config.case === 'userDefinedTransformerConfig'
                ) {
                  transformer = userDefinedMap.get(fv.config.value.id);
                } else {
                  transformer = systemMap.get(fv.source);
                }
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
                            userDefinedTransformers={
                              isForeignKey ? [] : userDefinedTransformers
                            }
                            userDefinedTransformerMap={
                              isForeignKey ? new Map() : userDefinedMap
                            }
                            systemTransformers={
                              isForeignKey
                                ? fkSystemTransformers
                                : systemTransformers
                            }
                            systemTransformerMap={
                              isForeignKey ? fkSystemTransformersMap : systemMap
                            }
                            value={fv}
                            onSelect={field.onChange}
                            placeholder="Select Transformer..."
                            side={'left'}
                            disabled={false}
                          />
                        </div>
                        {transformer && (
                          <EditTransformerOptions
                            transformer={transformer}
                            value={fv}
                            onSubmit={(newvalue) => {
                              field.onChange(newvalue);
                            }}
                            disabled={false}
                          />
                        )}
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
  className = 'w-4 h-4 flex',
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
      return 'timestamp(tz)';
    case 'timestamp without time zone':
      return 'timestamp';
    default:
      return dataType;
  }
}
