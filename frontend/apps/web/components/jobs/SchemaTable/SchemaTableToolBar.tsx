'use client';

import { Row, Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { isSystemTransformer, Transformer } from '@/shared/transformers';
import {
  getTransformerFromField,
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import {
  convertJobMappingTransformerToForm,
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import {
  GenerateDefault,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';
import {
  CheckIcon,
  Cross2Icon,
  ExclamationTriangleIcon,
} from '@radix-ui/react-icons';
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

  const transformer = getTransformerFromField(
    transformerHandler,
    bulkTransformer
  );
  // conditionally computed the allowed transformers only if there are selected rows
  const allowedTransformers = hasSelectedRows
    ? getFilteredTransformersForBulkSet(
        table.getSelectedRowModel().rows,
        transformerHandler,
        constraintHandler,
        jobType
      )
    : { system: [], userDefined: [] };
  const isBulkApplyDisabled =
    !bulkTransformer ||
    !hasSelectedRows ||
    !isTransformerAllowed(allowedTransformers, transformer);

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-row justify-between pb-2 items-center w-full">
        <div className="flex flex-col md:flex-row gap-3 w-[250px]">
          <TransformerSelect
            getTransformers={() => allowedTransformers}
            value={bulkTransformer}
            side={'bottom'}
            onSelect={(value) => {
              setBulkTransformer(value);
            }}
            buttonText={getTransformerSelectButtonText(
              transformer,
              'Bulk set transformers'
            )}
            disabled={!hasSelectedRows}
            buttonClassName="w-[275px]"
            notFoundText="No transformers found for the given selection."
          />
          <EditTransformerOptions
            transformer={transformer}
            value={bulkTransformer}
            onSubmit={setBulkTransformer}
            disabled={!hasSelectedRows || isInvalidTransformer(transformer)}
          />
          <div className="flex flex-row gap-1">
            <Button
              disabled={isBulkApplyDisabled}
              type="button"
              variant="outline"
              className={cn(
                isBulkApplyDisabled ? undefined : 'border-blue-600'
              )}
              onClick={() => {
                table.getSelectedRowModel().rows.forEach((r) => {
                  form.setValue(
                    `mappings.${r.index}.transformer`,
                    bulkTransformer,
                    {
                      shouldDirty: true,
                      shouldTouch: true,
                      shouldValidate: false, // this is really expensive, see the trigger call below
                    }
                  );
                });
                setBulkTransformer(
                  convertJobMappingTransformerToForm(
                    new JobMappingTransformer()
                  )
                );
                form.trigger('mappings'); // trigger validation after bulk updating the selected form options
                table.resetRowSelection(true);
              }}
            >
              <CheckIcon />
            </Button>
            {isBulkApplyDisabled &&
              hasSelectedRows &&
              !isTransformerAllowed(allowedTransformers, transformer) && (
                <TooltipProvider>
                  <Tooltip delayDuration={0}>
                    <TooltipTrigger asChild>
                      <div>
                        <ExclamationTriangleIcon className="h-4 w-4  dark:text-yellow-400 text-yellow-600" />
                      </div>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>
                        Can't apply bulk Transformer. The selected rows don't
                        have any overlapping Transformers.
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              )}
          </div>
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
              <ButtonText
                leftIcon={<Cross2Icon className="h-3 w-3" />}
                text="Clear filters"
              />
            </Button>
          )}
          {jobType === 'sync' && (
            <ConfirmationDialog
              trigger={
                <Button
                  variant="outline"
                  type="button"
                  disabled={form.getValues('mappings').length == 0}
                >
                  <ButtonText text="Apply Default Transformers" />
                </Button>
              }
              headerText="Apply Default Transformers?"
              description="This setting will apply the 'Passthrough' Transformer to every column that is not Generated, while applying the 'Use Column Default' Transformer to all Generated columns."
              buttonText="Apply"
              onConfirm={() => {
                const formMappings = form.getValues('mappings');
                formMappings.forEach((fm, idx) => {
                  const colkey = {
                    schema: fm.schema,
                    table: fm.table,
                    column: fm.column,
                  };
                  const isGenerated = constraintHandler.getIsGenerated(colkey);
                  const identityType =
                    constraintHandler.getIdentityType(colkey);
                  const newJm =
                    isGenerated || identityType
                      ? new JobMappingTransformer({
                          source: TransformerSource.GENERATE_DEFAULT,
                          config: new TransformerConfig({
                            config: {
                              case: 'generateDefaultConfig',
                              value: new GenerateDefault(),
                            },
                          }),
                        })
                      : new JobMappingTransformer({
                          source: TransformerSource.PASSTHROUGH,
                          config: new TransformerConfig({
                            config: {
                              case: 'passthroughConfig',
                              value: new Passthrough(),
                            },
                          }),
                        });

                  form.setValue(
                    `mappings.${idx}.transformer`,
                    convertJobMappingTransformerToForm(newJm),
                    {
                      shouldDirty: true,
                      shouldTouch: true,
                      shouldValidate: false,
                    }
                  );
                });
                form.trigger('mappings'); // trigger validation after bulk updating the selected form options
              }}
            />
          )}
          <SchemaTableViewOptions table={table} />
        </div>
      </div>
    </div>
  );
}

function isTransformerAllowed(
  {
    system,
    userDefined,
  }: {
    system: SystemTransformer[];
    userDefined: UserDefinedTransformer[];
  },
  selected: Transformer
): boolean {
  if (isInvalidTransformer(selected)) {
    return true; // allows folks to unset transformers. We should eventually make this a discrete button somewhere
  }
  if (isSystemTransformer(selected)) {
    return system.some((t) => t.source === selected.source);
  } else {
    return userDefined.some((t) => t.id === selected.id);
  }
}

function getFilteredTransformersForBulkSet<TData>(
  rows: Row<TData>[],
  transformerHandler: TransformerHandler,
  constraintHandler: SchemaConstraintHandler,
  jobType: JobType
): {
  system: SystemTransformer[];
  userDefined: UserDefinedTransformer[];
} {
  const systemArrays: SystemTransformer[][] = [];
  const userDefinedArrays: UserDefinedTransformer[][] = [];

  rows.forEach((row) => {
    const { system, userDefined } = transformerHandler.getFilteredTransformers(
      getTransformerFilter(
        constraintHandler,
        fromRowDataToColKey(row as unknown as Row<RowData>), // this will bite us at some point
        jobType
      )
    );
    systemArrays.push(system);
    userDefinedArrays.push(userDefined);
  });

  const uniqueSystemSources = findCommonSystemSources(systemArrays);
  const uniqueSystem = uniqueSystemSources
    .map((source) => transformerHandler.getSystemTransformerBySource(source))
    .filter((x): x is SystemTransformer => !!x);

  const uniqueIds = findCommonUserDefinedIds(userDefinedArrays);
  const uniqueUserDef = uniqueIds
    .map((id) => transformerHandler.getUserDefinedTransformerById(id))
    .filter((x): x is UserDefinedTransformer => !!x);

  return {
    system: uniqueSystem,
    userDefined: uniqueUserDef,
  };
}

function findCommonSystemSources(
  arrays: SystemTransformer[][]
): TransformerSource[] {
  const elementCount: Record<TransformerSource, number> = {} as Record<
    TransformerSource,
    number
  >;
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

function findCommonUserDefinedIds(
  arrays: UserDefinedTransformer[][]
): string[] {
  const elementCount: Record<string, number> = {};
  const subArrayCount = arrays.length;
  const commonElements: string[] = [];

  arrays.forEach((subArray) => {
    // Use a Set to ensure each element in a sub-array is counted only once
    new Set(subArray).forEach((element) => {
      if (!elementCount[element.id]) {
        elementCount[element.id] = 1;
      } else {
        elementCount[element.id]++;
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
