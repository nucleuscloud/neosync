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

  const defaultSchemaValues: Option[] = Array.from(schemaSet).map((item) => ({
    value: item,
    label: item,
  }));

  const [schemaOptions, setSchemaOptions] =
    useState<Option[]>(defaultSchemaValues);

  /* table names might not be unique across schemas so create a unique value to get unique combination of table and schema to show in drop down. We have to use a character that someone won't put in their schema or table name. So _ and - are out of the picture. It's highly unlikely that someone would put *** in their schema or table name. There is probably a better way to do this but going with this for right now.
   */
  const tableSchemaSet = new Set(
    dataRow.map((obj) => obj.schema + '***' + obj.table)
  );

  const defaultTableValues: Option[] = Array.from(tableSchemaSet).map(
    (item) => ({
      value: item,
      label: item.split('***')[1],
      schema: item.split('***')[0],
    })
  );

  const [tableOptions, setTableOptions] =
    useState<Option[]>(defaultTableValues);

  const schemaTableMap: Record<string, string[]> = {};
  const tableSchemaMap: Record<string, string> = {};

  /* creates the schemaTableMap <schema:[tables]> and tableSchemaMap <schema***table:schema>
   this is used to correctly filter for values both ways. If you select a schema you should only see the 
   tables in that schema and if you select a table, you should only see the schema that table belongs to 
   */

  dataRow.forEach((row) => {
    const { schema, table } = row;
    // update schemaTableMap
    if (!schemaTableMap[schema]) {
      schemaTableMap[schema] = [table];
    } else if (!schemaTableMap[schema].includes(table)) {
      schemaTableMap[schema].push(table);
    }

    // update tableSchemaMap
    if (!tableSchemaMap[row.schema + '***' + row.table]) {
      tableSchemaMap[row.schema + '***' + row.table] = schema;
    } else !tableSchemaMap[row.schema + '***' + row.table].includes(schema);
  });

  const handleMultiSelectSchemaChange = (selectedOptions: Option[]) => {
    setSelectedSchemaOptions(selectedOptions);
    const filteredSchemaValues = selectedOptions.map((option) => option.value);
    // handles the user removing items from the multi-select
    if (filteredSchemaValues.length > 0) {
      //create the table object [] to update the tables options
      const filteredTableOptions = filteredSchemaValues.flatMap(
        (schema) =>
          schemaTableMap[schema]?.map((table) => ({
            value: schema + '***' + table,
            label: table,
            schema: schema,
          })) || []
      );

      setTableOptions(filteredTableOptions);
      table.getColumn('schema')?.setFilterValue(filteredSchemaValues);
    } else {
      setSelectedTableOptions([]);
      table.getColumn('schema')?.setFilterValue(undefined);
      table.getColumn('table')?.setFilterValue(undefined);
    }
  };

  const handleMultiSelectTableChange = (selectedOptions: Option[]) => {
    setSelectedTableOptions(selectedOptions);
    // handles the user removing items from the multi-select
    if (selectedOptions.length > 0) {
      const uniqueSchemaOptions = new Set<string>();
      // iterate over selected table options and add them to the uniqueSchemaOptions to get list of unique schemas in case a user adds multiple tables from the same schema, we don't want to show the same schema more than once
      selectedOptions.forEach((table) => {
        // add a new schema to the set
        uniqueSchemaOptions.add(tableSchemaMap[table.value].split('***')[0]);
      });

      const selectedSchemaValues: Option[] = [];
      // iterate over the unique schema options and structure them as Option types so that we can render the selected schema badges in the multi-select
      Array.from(uniqueSchemaOptions.values()).map((item: string) => {
        selectedSchemaValues.push({
          value: item,
          label: item,
        });
      });
      // update the schema options that show up in the drop down, should be whatever schemas are remaining that aren't selected
      setSchemaOptions(
        defaultSchemaValues.filter(
          (item) => !uniqueSchemaOptions.has(item.value)
        )
      );
      // update the selected schemas that show up as badges in the multi select to only be the schemas of the selected tables
      setSelectedSchemaOptions(selectedSchemaValues);
      // filter the table column by only the tables that were selected
      table
        .getColumn('table')
        ?.setFilterValue(selectedOptions.map((option) => option.label));
      // filter the schema column by only the tables that were selected. This handles the edge case where two tables have the same name, so we need to differentiate by schema as well
      table
        .getColumn('schema')
        ?.setFilterValue(
          selectedOptions.map((option) => option.value.split('***')[0])
        );
    } else {
      // reset tbale and schema options
      setSelectedSchemaOptions([]);
      setSchemaOptions(defaultSchemaValues);
      setTableOptions(defaultTableValues);
      table.getColumn('table')?.setFilterValue(undefined);
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
        <div className="w-[275px] lg:w-[650px] z-40 flex flex-col gap-2">
          <FormLabel>Filter Schema(s)</FormLabel>
          <MultiSelect
            defaultOptions={defaultSchemaValues}
            options={schemaOptions}
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
        <div className="w-[275px] lg:w-[650px] z-30 flex flex-col gap-2">
          <FormLabel>Filter Table(s)</FormLabel>
          <MultiSelect
            defaultOptions={defaultTableValues}
            options={tableOptions}
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
