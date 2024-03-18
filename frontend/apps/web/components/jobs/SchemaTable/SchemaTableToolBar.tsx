'use client';

import { Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Button } from '@/components/ui/button';
import {
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import { JobMappingTransformer } from '@neosync/sdk';
import { CheckIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  transformers: Transformer[];
}

export function SchemaTableToolbar<TData>({
  table,
  transformers,
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

  const transformer = transformers.find((t) => {
    if (
      isUserDefinedTransformer(t) &&
      bulkTransformer.source === 'custom' &&
      bulkTransformer.config.case === 'useefinedTransformerConfig'
    ) {
      return t;
    } else if (isSystemTransformer(t) && bulkTransformer.source === t.source) {
      return t;
    }
    return undefined;
  });

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-row justify-between pb-2 items-center w-full">
        <div className="flex flex-col md:flex-row gap-3 w-[250px]">
          <TransformerSelect
            transformers={transformers}
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
