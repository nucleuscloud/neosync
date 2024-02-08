'use client';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
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
import { ExclamationTriangleIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { HTMLProps, useEffect, useRef } from 'react';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import { JobMapRow } from './makeData';

interface Props {
  transformers: Transformer[];
}

export function getSchemaColumns(props: Props): ColumnDef<JobMapRow>[] {
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
      size: 260,
    },
    {
      accessorKey: 'schema',
      // header: ({ column }) => (
      //   <SchemaTableColumnHeader column={column} title="Schema" />
      // ),
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
      // header: ({ column }) => (
      //   <SchemaTableColumnHeader column={column} title="Table" />
      // ),
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
      // header: ({ column }) => (
      //    <SchemaTableColumnHeader column={column} title="Column" />
      // ),
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
      // header: ({ column }) => (
      //   <SchemaTableColumnHeader column={column} title="Data Type" />
      // ),
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
      // header: ({ column }) => (
      //   <SchemaTableColumnHeader column={column} title="Transformer" />
      // ),
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
  ];
}

function IndeterminateCheckbox({
  indeterminate,
  className = '',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = useRef<HTMLInputElement>(null!);

  useEffect(() => {
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
