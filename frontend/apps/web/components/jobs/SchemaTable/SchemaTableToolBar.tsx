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
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  userDefinedTransformers: UserDefinedTransformer[];
  systemTransformers: SystemTransformer[];

  userDefinedTransformerMap: Map<string, UserDefinedTransformer>;
  systemTransformerMap: Map<string, SystemTransformer>;
}

export function SchemaTableToolbar<TData>({
  table,
  userDefinedTransformerMap,
  userDefinedTransformers,
  systemTransformerMap,
  systemTransformers,
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
            placeholder="Bulk update Transformers"
            disabled={!hasSelectedRows}
          />
          <Button
            disabled={!bulkTransformer || !hasSelectedRows}
            type="button"
            variant="outline"
            onClick={() => {
              table.getSelectedRowModel().rows.forEach((r) => {
                form.setValue(
                  `mappings.${r.index}.transformer`,
                  bulkTransformer,
                  {
                    shouldDirty: true,
                  }
                );
              });
              setBulkTransformer(
                convertJobMappingTransformerToForm(new JobMappingTransformer())
              );
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
