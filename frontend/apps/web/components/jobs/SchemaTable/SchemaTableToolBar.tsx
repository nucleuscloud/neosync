'use client';

import { Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/schema';
import { Button } from '@/components/ui/button';
import { Transformer } from '@/shared/transformers';
import {
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  // data: TData[];
  transformers: Transformer[];
  // jobType: string;
}

export function SchemaTableToolbar<TData>({
  table,
  // data,
  transformers,
}: DataTableToolbarProps<TData>) {
  // const [selectedSchemaOptions, setSelectedSchemaOptions] = useState<Option[]>(
  //   []
  // );

  // const [selectedTableOptions, setSelectedTableOptions] = useState<Option[]>(
  //   []
  // );

  const isFiltered = table.getState().columnFilters.length > 0;

  // const dataRow = data as Row[];

  // const schemaSet = new Set(dataRow.map((obj) => obj.schema));

  // const defaultSchemaValues: Option[] = Array.from(schemaSet).map((item) => ({
  //   value: item,
  //   label: item,
  // }));

  // const [schemaOptions, setSchemaOptions] =
  //   useState<Option[]>(defaultSchemaValues);

  // /* table names might not be unique across schemas so create a unique value to get unique combination of table and schema to show in drop down. We have to use a character that someone won't put in their schema or table name. So _ and - are out of the picture. It's highly unlikely that someone would put *** in their schema or table name. There is probably a better way to do this but going with this for right now.
  //  */
  // const tableSchemaSet = new Set(
  //   dataRow.map((obj) => obj.schema + '***' + obj.table)
  // );

  // const defaultTableValues: Option[] = Array.from(tableSchemaSet).map(
  //   (item) => ({
  //     value: item,
  //     label: item.split('***')[1],
  //     schema: item.split('***')[0],
  //   })
  // );

  // const [tableOptions, setTableOptions] =
  //   useState<Option[]>(defaultTableValues);

  // const schemaTableMap: Record<string, string[]> = {};
  // const tableSchemaMap: Record<string, string> = {};

  /* creates the schemaTableMap <schema:[tables]> and tableSchemaMap <schema***table:schema>
   this is used to correctly filter for values both ways. If you select a schema you should only see the
   tables in that schema and if you select a table, you should only see the schema that table belongs to
   */

  // dataRow.forEach((row) => {
  //   const { schema, table } = row;
  //   // update schemaTableMap
  //   if (!schemaTableMap[schema]) {
  //     schemaTableMap[schema] = [table];
  //   } else if (!schemaTableMap[schema].includes(table)) {
  //     schemaTableMap[schema].push(table);
  //   }

  //   // update tableSchemaMap
  //   if (!tableSchemaMap[row.schema + '***' + row.table]) {
  //     tableSchemaMap[row.schema + '***' + row.table] = schema;
  //   } else !tableSchemaMap[row.schema + '***' + row.table].includes(schema);
  // });

  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>({
      source: '',
      config: { case: '', value: {} },
    });

  const form = useFormContext<SingleTableSchemaFormValues | SchemaFormValues>();

  return (
    <div className="flex flex-col items-start w-full gap-2">
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
            disabled={false}
          />
        </div>
        <div className="flex flex-row items-center gap-2">
          {isFiltered && (
            <Button
              variant="outline"
              type="button"
              onClick={() => {
                // setSelectedSchemaOptions([]);
                // setSelectedTableOptions([]);
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
