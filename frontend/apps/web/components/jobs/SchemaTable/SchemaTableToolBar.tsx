'use client';

import { Row, Table } from '@tanstack/react-table';

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
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { CheckIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useFormContext } from 'react-hook-form';
import { fromRowDataToColKey, getTransformerFilter } from './SchemaColumns';
import { Row as RowData } from './SchemaPageTable';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import { TransformerHandler } from './transformer-handler';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  transformerHandler: TransformerHandler;
  constraintHandler: SchemaConstraintHandler;
  jobType: JobType;
}

export function SchemaTableToolbar<TData>({
  table,
  transformerHandler,
  constraintHandler,
  jobType,
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
    bulkTransformer.source === TransformerSource.USER_DEFINED &&
    bulkTransformer.config.case === 'userDefinedTransformerConfig'
  ) {
    transformer = transformerHandler.getUserDefinedTransformerById(
      bulkTransformer.config.value.id
    );
  } else {
    transformer = transformerHandler.getSystemTransformerBySource(
      bulkTransformer.source
    );
  }
  const buttonText = transformer ? transformer.name : 'Bulk set transformers';

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-row justify-between pb-2 items-center w-full">
        <div className="flex flex-col md:flex-row gap-3 w-[250px]">
          <TransformerSelect
            getTransformers={() => {
              const systemArrays: SystemTransformer[][] = [];
              const userDefinedArrays: UserDefinedTransformer[][] = [];
              table.getSelectedRowModel().rows.forEach((row) => {
                const { system, userDefined } =
                  transformerHandler.getFilteredTransformers(
                    getTransformerFilter(
                      constraintHandler,
                      fromRowDataToColKey(row as unknown as Row<RowData>),
                      jobType
                    )
                  );
                systemArrays.push(system);
                userDefinedArrays.push(userDefined);
              });
              const uniqueSystemSources = findCommonSystem(systemArrays);
              const uniqueSystem = uniqueSystemSources
                .map((source) =>
                  transformerHandler.getSystemTransformerBySource(source)
                )
                .filter((x): x is SystemTransformer => !!x);
              const uniqueIds = findCommonUserDefined(userDefinedArrays);
              const uniqueUserDef = uniqueIds
                .map((id) =>
                  transformerHandler.getUserDefinedTransformerById(id)
                )
                .filter((x): x is UserDefinedTransformer => !!x);
              return {
                system: uniqueSystem,
                userDefined: uniqueUserDef,
              };
            }}
            value={bulkTransformer}
            side={'bottom'}
            onSelect={(value) => {
              setBulkTransformer(value);
            }}
            buttonText={buttonText}
            disabled={!hasSelectedRows}
            buttonClassName="w-[275px]"
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
                      shouldValidate: false, // this is really expensive, see the trigger call below
                    }
                  );
                }
              });
              setBulkTransformer(
                convertJobMappingTransformerToForm(new JobMappingTransformer())
              );
              form.trigger('mappings'); // trigger validation after bulk updating the selected form options
              table.resetRowSelection(true);
            }}
          >
            <CheckIcon />
          </Button>
          <EditTransformerOptions
            transformer={transformer ?? new SystemTransformer({})}
            value={bulkTransformer}
            onSubmit={setBulkTransformer}
            disabled={!transformer}
          />
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
  if (!isForeignKey || transformer.source === TransformerSource.UNSPECIFIED) {
    return true;
  }
  const valid = new Set<TransformerSource>([TransformerSource.PASSTHROUGH]);
  if (isNullable) {
    valid.add(TransformerSource.GENERATE_NULL);
  }

  return valid.has(transformer.source);
}

function findCommonSystem(arrays: SystemTransformer[][]): TransformerSource[] {
  const elementCount: Record<TransformerSource, number> = {} as any;
  const subArrayCount = arrays.length;
  const commonElements: TransformerSource[] = [];

  arrays.forEach((subArray) => {
    // Use a Set to ensure each element in a sub-array is counted only once
    new Set(subArray).forEach((element) => {
      if (!elementCount[element.source]) {
        elementCount[element.source] = 1;
      } else {
        elementCount[element.source]++;
      }
    });
  });

  for (const [element, count] of Object.entries(elementCount)) {
    if (count === subArrayCount) {
      commonElements.push(+element as TransformerSource);
    }
  }

  return commonElements;
}

function findCommonUserDefined(arrays: UserDefinedTransformer[][]): string[] {
  const elementCount: Record<string, number> = {};
  const subArrayCount = arrays.length;
  const commonElements: string[] = [];

  arrays.forEach((subArray) => {
    // Use a Set to ensure each element in a sub-array is counted only once
    new Set(subArray).forEach((element) => {
      if (!elementCount[element.source]) {
        elementCount[element.source] = 1;
      } else {
        elementCount[element.source]++;
      }
    });
  });

  for (const [element, count] of Object.entries(elementCount)) {
    if (count === subArrayCount) {
      commonElements.push(element);
    }
  }

  return commonElements;
}
