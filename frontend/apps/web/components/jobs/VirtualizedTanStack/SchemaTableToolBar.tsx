'use client';

import { Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import { MultiSelect, Option } from '@/components/MultiSelect';
import { Button } from '@/components/ui/button';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import { Row } from './main';

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
  const [selectedOptions, setSelectedOptions] = useState<Option[]>([]);

  const dataRow = data as Row[];

  const schemaSet = new Set(dataRow.map((obj) => obj.schema));

  const schemaValues: Option[] = Array.from(schemaSet).map((item) => ({
    value: item,
    label: item,
  }));

  const handleMultiSelectChange = (selectedOptions: Option[]) => {
    setSelectedOptions(selectedOptions);
    const filteredValues = selectedOptions.map((option) => option.value);
    // handles the user removing items from the multi-se
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
    <div className="flex flex-col items-start w-full">
      <div className="flex flex-1 items-center space-x-2">
        <div className="w-[150px] lg:w-[650px] z-50 flex flex-col">
          <div className="text-xs text-gray-600 dark:text-300">
            Total rows: ({new Intl.NumberFormat('en-US').format(data.length)})
            Rows visible: (
            {new Intl.NumberFormat('en-US').format(
              table.getRowModel().rows.length
            )}
            )
          </div>
          <MultiSelect
            defaultOptions={schemaValues}
            placeholder="Filter by schema(s)..."
            emptyIndicator={
              <p className="text-center text-lg leading-10 text-gray-600 dark:text-gray-400">
                No schemas available to filter
              </p>
            }
            value={selectedOptions}
            hidePlaceholderWhenSelected={true}
            onChange={(selectedOptions) =>
              handleMultiSelectChange(selectedOptions)
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
            placeholder="Bulk update Transformers..."
          />
        </div>
        <div className="flex flex-row items-center gap-2">
          <Button
            variant="outline"
            type="button"
            onClick={() => {
              setSelectedOptions([]);
              table.resetColumnFilters();
            }}
            className="h-8 px-2 lg:px-3"
          >
            <Cross2Icon className="mr-2 h-3 w-3" />
            Clear Filters
          </Button>
          <SchemaTableViewOptions table={table} />
        </div>
      </div>
    </div>
  );
}
