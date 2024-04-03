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
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { SystemTransformer, TransformerSource } from '@neosync/sdk';
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef, Row } from '@tanstack/react-table';
import { HTMLProps, ReactElement, useEffect, useRef } from 'react';
import { useFormContext } from 'react-hook-form';
import { SchemaColumnHeader } from './SchemaColumnHeader';
import { Row as RowData } from './SchemaPageTable';
import TransformerSelect from './TransformerSelect';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import {
  TransformerFilters,
  TransformerHandler,
  toSupportedJobtype,
} from './transformer-handler';

interface ColumnKey {
  schema: string;
  table: string;
  column: string;
}

export function fromRowDataToColKey(row: Row<RowData>): ColumnKey {
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
  transformerHandler: TransformerHandler;
  constraintHandler: SchemaConstraintHandler;
  jobType: JobType;
}

export function getSchemaColumns(props: Props): ColumnDef<RowData>[] {
  const { transformerHandler, constraintHandler, jobType } = props;

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
      maxSize: 500,
      size: 300,
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
      maxSize: 500,
      size: 200,
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
        const isUniqueConstraint = constraintHandler.getIsUniqueConstraint(key);

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
        const isUniqueConstraint = constraintHandler.getIsUniqueConstraint(key);
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
              <div>
                {isUniqueConstraint && (
                  <Badge
                    variant="outline"
                    className="text-xs bg-purple-100 text-gray-800 cursor-default dark:bg-purple-300 dark:text-gray-900"
                  >
                    Unique
                  </Badge>
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
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div>
                  <Badge variant="outline" className="max-w-[100px]">
                    <span className="truncate block overflow-hidden">
                      {handleDataTypeBadge(datatype)}
                    </span>
                  </Badge>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                <p>{datatype}</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
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
                const colkey = fromRowDataToColKey(info.row);

                const filtered = transformerHandler.getFilteredTransformers(
                  getTransformerFilter(constraintHandler, colkey, jobType)
                );

                const filteredTransformerHandler = new TransformerHandler(
                  filtered.system,
                  filtered.userDefined
                );

                let transformer: Transformer | undefined;
                if (
                  fv.source === TransformerSource.USER_DEFINED &&
                  fv.config.case === 'userDefinedTransformerConfig'
                ) {
                  transformer =
                    filteredTransformerHandler.getUserDefinedTransformerById(
                      fv.config.value.id
                    );
                } else {
                  transformer =
                    filteredTransformerHandler.getSystemTransformerBySource(
                      fv.source
                    );
                }
                const buttonText = transformer
                  ? transformer.name
                  : 'Select Transformer';
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
                            getTransformers={() => filtered}
                            buttonText={buttonText}
                            value={fv}
                            onSelect={field.onChange}
                            side={'left'}
                            disabled={false}
                            buttonClassName="w-[175px]"
                          />
                        </div>
                        <EditTransformerOptions
                          transformer={transformer ?? new SystemTransformer()}
                          value={fv}
                          onSubmit={(newvalue) => {
                            field.onChange(newvalue);
                          }}
                          disabled={!transformer}
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

// cleans up the data type values since some are too long , can add on more here
function handleDataTypeBadge(dataType: string): string {
  // Check for "timezone" and replace accordingly without entering the switch
  if (dataType.includes('timezone')) {
    return dataType
      .replace('timestamp with time zone', 'timestamp(tz)')
      .replace('timestamp without time zone', 'timestamp');
  }

  const splitDt = dataType.split('(');
  switch (splitDt[0]) {
    case 'character varying':
      // The condition inside the if statement seemed reversed. It should return 'varchar' directly if splitDt[1] is undefined.
      return splitDt[1] !== undefined ? `varchar(${splitDt[1]}` : 'varchar';
    default:
      return dataType;
  }
}

export function getTransformerFilter(
  constraintHandler: SchemaConstraintHandler,
  colkey: ColumnKey,
  jobType: JobType
): TransformerFilters {
  const [isForeignKey] = constraintHandler.getIsForeignKey(colkey);
  const isNullable = constraintHandler.getIsNullable(colkey);
  const convertedDataType = constraintHandler.getConvertedDataType(colkey);
  const hasDefault = constraintHandler.getHasDefault(colkey);
  return {
    dataType: convertedDataType,
    hasDefault,
    isForeignKey,
    isNullable,
    jobType: toSupportedJobtype(jobType),
  };
}
