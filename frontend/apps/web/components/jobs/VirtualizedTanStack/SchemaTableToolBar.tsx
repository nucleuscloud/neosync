'use client';

import { Table } from '@tanstack/react-table';

import { Button } from '@/components/ui/button';
import { Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { MultiSelect, Option } from './SchemaMultiSelect';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import { Row } from './main';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  data: TData[];
}

export function SchemaTableToolbar<TData>({
  table,
  data,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0;

  const [selectedOptions, setSelectedOptions] = useState<Option[]>([]);

  const dataRow = data as Row[];

  const schemaSet = new Set(dataRow.map((obj) => obj.schema));

  const schemaValues: Option[] = Array.from(schemaSet).map((item) => ({
    value: item,
    label: item,
  }));

  const schemaFilter = table.getColumn('schema')?.getIsFiltered();

  const handleMultiSelectChange = (selectedOptions: Option[]) => {
    setSelectedOptions(selectedOptions);
    const filteredValues = selectedOptions.map((option) => option.value);
    table.getColumn('schema')?.setFilterValue(filteredValues);
  };

  const resetFilters = () => {
    setSelectedOptions([]);
    table.resetColumnFilters();
  };

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-1 items-center space-x-2">
        <div className="w-[150px] lg:w-[650px] z-50 flex flex-col">
          <div className="flex flex-row items-center gap-2">
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
            {schemaFilter && (
              <Button
                variant="outline"
                size="sm"
                onClick={resetFilters}
                className="h-8 px-2 mb-2 lg:px-3"
              >
                <Cross2Icon className="mr-2 h-3 w-3" />
                Reset
              </Button>
            )}
          </div>
          <div className="text-xs text-gray-600 dark:text-300">
            Total rows: ({new Intl.NumberFormat('en-US').format(data.length)})
            Rows visible: (
            {new Intl.NumberFormat('en-US').format(
              table.getRowModel().rows.length
            )}
            )
          </div>
        </div>
      </div>
      <div className="flex flex-row gap-2 pb-2">
        {isFiltered && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.resetColumnFilters()}
            className="h-8 px-2 lg:px-3"
          >
            <Cross2Icon className="mr-2 h-3 w-3" />
            Reset Column Filters
          </Button>
        )}
        <SchemaTableViewOptions table={table} />
      </div>
    </div>
  );
}
