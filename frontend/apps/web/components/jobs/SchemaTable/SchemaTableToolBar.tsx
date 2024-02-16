'use client';

import { Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import { MultiSelect, Option } from '@/components/MultiSelect';
import { Button } from '@/components/ui/button';
import { FormLabel } from '@/components/ui/form';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { Row } from './SchemaPageTable';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  data: TData[];
  transformers: Transformer[];
}

export function SchemaTableToolbar<TData>({
  table,
  data,
  transformers,
}: DataTableToolbarProps<TData>) {
  const [selectedSchemaOptions, setSelectedSchemaOptions] = useState<Option[]>(
    []
  );

  const [selectedTableOptions, setSelectedTableOptions] = useState<Option[]>(
    []
  );

  const isFiltered = table.getState().columnFilters.length > 0;

  const dataRow = data as Row[];

  const schemaSet = new Set(dataRow.map((obj) => obj.schema));

  const tableSet = new Set(dataRow.map((obj) => obj.table));

  const schemaValues: Option[] = Array.from(schemaSet).map((item) => ({
    value: item,
    label: item,
  }));

  const tableValues: Option[] = Array.from(tableSet).map((item) => ({
    value: item,
    label: item,
  }));

  const handleMultiSelectSchemaChange = (selectedOptions: Option[]) => {
    setSelectedSchemaOptions(selectedOptions);
    const filteredValues = selectedOptions.map((option) => option.value);
    // handles the user removing items from the multi-select
    if (filteredValues.length > 0) {
      table.getColumn('schema')?.setFilterValue(filteredValues);
    } else {
      table.getColumn('schema')?.setFilterValue(undefined);
    }
  };

  const handleMultiSelectTableChange = (selectedOptions: Option[]) => {
    setSelectedSchemaOptions(selectedOptions);
    const filteredValues = selectedOptions.map((option) => option.value);
    // handles the user removing items from the multi-select
    if (filteredValues.length > 0) {
      table.getColumn('schema')?.setFilterValue(filteredValues);
    } else {
      table.getColumn('schema')?.setFilterValue(undefined);
    }
  };

  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>({
      source: '',
      config: { case: '', value: {} },
    });

  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-col items-center gap-2">
        <div className="w-[275px] lg:w-[650px] z-50 flex flex-col gap-2">
          <FormLabel>Filter Schema(s)</FormLabel>
          <MultiSelect
            defaultOptions={schemaValues}
            placeholder="Filter by schema(s)..."
            emptyIndicator={
              <p className="text-center text-sm leading-10 text-gray-600 dark:text-gray-400">
                No schemas available to filter
              </p>
            }
            value={selectedSchemaOptions}
            hidePlaceholderWhenSelected={true}
            onChange={(selectedOptions) =>
              handleMultiSelectSchemaChange(selectedOptions)
            }
          />
        </div>
        <div className="w-[275px] lg:w-[650px] z-40 flex flex-col gap-2">
          <FormLabel>Filter Table(s)</FormLabel>
          <MultiSelect
            defaultOptions={tableValues}
            placeholder="Filter by table(s)..."
            emptyIndicator={
              <p className="text-center text-sm leading-10 text-gray-600 dark:text-gray-400">
                No tables available to filter
              </p>
            }
            value={selectedTableOptions}
            hidePlaceholderWhenSelected={true}
            onChange={(selectedOptions) =>
              handleMultiSelectTableChange(selectedOptions)
            }
          />
        </div>
      </div>
      <div className="flex flex-row justify-between pb-2 items-center w-full">
        <div className="w-[250px]">
          <TransformerSelect
            transformers={transformers}
            value={bulkTransformer}
            side={'bottom'}
            onSelect={(value) => {
              table.getSelectedRowModel().rows.forEach((r) => {
                form.setValue(`mappings.${r.index}.transformer`, value, {
                  shouldDirty: true,
                });
              });
              setBulkTransformer({
                source: '',
                config: { case: '', value: {} },
              });
            }}
            placeholder="Bulk update Transformers"
          />
        </div>
        <div className="flex flex-row items-center gap-2">
          {isFiltered && (
            <Button
              variant="outline"
              type="button"
              onClick={() => {
                setSelectedSchemaOptions([]);
                setSelectedTableOptions([]);
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
