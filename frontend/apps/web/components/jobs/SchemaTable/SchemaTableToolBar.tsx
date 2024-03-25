'use client';

import { Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Button } from '@/components/ui/button';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import {
  JobMappingTransformer,
  SystemTransformer,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { CheckIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { SchemaConstraintHandler } from './SchemaColumns';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  userDefinedTransformers: UserDefinedTransformer[];
  systemTransformers: SystemTransformer[];

  userDefinedTransformerMap: Map<string, UserDefinedTransformer>;
  systemTransformerMap: Map<string, SystemTransformer>;
  constraintHandler: SchemaConstraintHandler;
}

export function SchemaTableToolbar<TData>({
  table,
  userDefinedTransformerMap,
  userDefinedTransformers,
  systemTransformerMap,
  systemTransformers,
  constraintHandler,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0;
  const hasSelectedRows = Object.values(table.getState().rowSelection).some(
    (value) => value
  );

  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>(
      convertJobMappingTransformerToForm(new JobMappingTransformer())
    );

  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();

  let transformer: Transformer | undefined;
  if (
    bulkTransformer.source === 'custom' &&
    bulkTransformer.config.case === 'userDefinedTransformerConfig'
  ) {
    transformer = userDefinedTransformerMap.get(
      bulkTransformer.config.value.id
    );
  } else {
    transformer = systemTransformerMap.get(bulkTransformer.source);
  }

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-row justify-between pb-2 items-center w-full">
        <div className="flex flex-col md:flex-row gap-3 w-[250px]">
          <TransformerSelect
            systemTransformerMap={systemTransformerMap}
            systemTransformers={systemTransformers}
            userDefinedTransformerMap={userDefinedTransformerMap}
            userDefinedTransformers={userDefinedTransformers}
            value={bulkTransformer}
            side={'bottom'}
            onSelect={(value) => {
              setBulkTransformer(value);
            }}
            placeholder="Bulk set transformers"
            disabled={!hasSelectedRows}
          />
          <Button
            disabled={!bulkTransformer || !hasSelectedRows}
            type="button"
            variant="outline"
            onClick={() => {
              table.getSelectedRowModel().rows.forEach((r) => {
                const colkey = {
                  schema: r.getValue<string>('schema'),
                  table: r.getValue<string>('table'),
                  column: r.getValue<string>('column'),
                };
                const [isForeignKey] =
                  constraintHandler.getIsForeignKey(colkey);
                const isNullable = constraintHandler.getIsNullable(colkey);
                if (
                  isBulkUpdateable(bulkTransformer, isForeignKey, isNullable)
                ) {
                  form.setValue(
                    `mappings.${r.index}.transformer`,
                    bulkTransformer,
                    {
                      shouldDirty: true,
                      shouldTouch: true,
                      shouldValidate: false, // this is really expensive
                    }
                  );
                }
              });
              setBulkTransformer(
                convertJobMappingTransformerToForm(new JobMappingTransformer())
              );
              form.trigger('mappings');
              table.resetRowSelection(true);
            }}
          >
            <CheckIcon />
          </Button>
          {transformer && (
            <EditTransformerOptions
              transformer={transformer}
              value={bulkTransformer}
              onSubmit={setBulkTransformer}
              disabled={false}
            />
          )}
        </div>
        <div className="flex flex-row items-center gap-2">
          {isFiltered && (
            <Button
              variant="outline"
              type="button"
              onClick={() => {
                table.resetColumnFilters();
              }}
              className="h-8 px-2 lg:px-3"
            >
              <Cross2Icon className="mr-2 h-3 w-3" />
              Clear Filters
            </Button>
          )}
          <SchemaTableViewOptions table={table} />
        </div>
      </div>
    </div>
  );
}

// see method in SchemaColumns.tsx transformer cell
// should update to use similar logic
function isBulkUpdateable(
  transformer: JobMappingTransformerForm,
  isForeignKey: boolean,
  isNullable: boolean
): boolean {
  console.log('transformer', transformer);
  if (!isForeignKey || transformer.source === '') {
    return true;
  }
  const valid = new Set(['passthrough']);
  if (isNullable) {
    valid.add('null');
  }

  return valid.has(transformer.source);
}
