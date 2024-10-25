'use client';

import { Row, Table } from '@tanstack/react-table';

import { SingleTableSchemaFormValues } from '@/app/(mgmt)/[account]/new/job/job-form-validations';
import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import ButtonText from '@/components/ButtonText';
import ConfirmationDialog from '@/components/ConfirmationDialog';
import FormErrorMessage from '@/components/FormErrorMessage';
import SwitchCard from '@/components/switches/SwitchCard';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import { cn } from '@/libs/utils';
import { isSystemTransformer, Transformer } from '@/shared/transformers';
import {
  getTransformerFromField,
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import {
  convertJobMappingTransformerToForm,
  DefaultTransformerFormValues,
  JobMappingTransformerForm,
  SchemaFormValues,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GenerateDefault,
  JobMappingTransformer,
  Passthrough,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { CheckIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { useForm, useFormContext } from 'react-hook-form';
import { fromRowDataToColKey, getTransformerFilter } from './SchemaColumns';
import { Row as RowData } from './SchemaPageTable';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';
import { JobType, SchemaConstraintHandler } from './schema-constraint-handler';
import { TransformerHandler } from './transformer-handler';

export type TableViewType = 'table' | 'json';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  transformerHandler: TransformerHandler;
  constraintHandler: SchemaConstraintHandler;
  jobType: JobType;
  currentView?: TableViewType;
  onViewChange(view: TableViewType): void;
}

export function SchemaTableToolbar<TData>({
  table,
  transformerHandler,
  constraintHandler,
  jobType,
  currentView = 'table',
  onViewChange,
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

  const defaultTransformerForm = useForm<DefaultTransformerFormValues>({
    resolver: yupResolver(DefaultTransformerFormValues),
    defaultValues: {
      overrideTransformers: false,
    },
  });

  const handleAlertDescriptionBody = (): JSX.Element => {
    return (
      <div>
        <Form {...form}>
          <form className="space-y-8">
            <FormField
              control={defaultTransformerForm.control}
              name="overrideTransformers"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <SwitchCard
                      isChecked={field.value}
                      onCheckedChange={field.onChange}
                      title="Override Mapped Transformers"
                      description="Do you want to overwrite the Transformers you have already mapped."
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </form>
        </Form>
      </div>
    );
  };

  return (
    <div className="flex flex-col items-start w-full gap-2">
      <div className="flex flex-col md:flex-row justify-between pb-2 md:items-center w-full gap-3">
        <div className="flex flex-col md:flex-row gap-3">
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
            buttonClassName="md:max-w-[275px]"
            notFoundText="No transformers found for the given selection."
          />
          <EditTransformerOptions
            transformer={transformer}
            value={bulkTransformer}
            onSubmit={setBulkTransformer}
            disabled={!hasSelectedRows || isInvalidTransformer(transformer)}
          />
          <Button
            disabled={isBulkApplyDisabled}
            type="button"
            variant="outline"
            className={cn(isBulkApplyDisabled ? undefined : 'border-blue-600')}
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
                convertJobMappingTransformerToForm(new JobMappingTransformer())
              );
              form.trigger('mappings'); // trigger validation after bulk updating the selected form options
              table.resetRowSelection(true);
            }}
          >
            <CheckIcon />
          </Button>
          <div className="flex items-center">
            {isBulkApplyDisabled &&
              hasSelectedRows &&
              !isTransformerAllowed(allowedTransformers, transformer) && (
                <FormErrorMessage
                  message={`Can't apply bulk Transformer. The selected rows don't
                        have any overlapping Transformers.`}
                />
              )}
          </div>
        </div>
        <div className="flex flex-col md:flex-row md:items-center gap-2">
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
              description="This setting will apply the 'Passthrough' Transformer to every column that is not Generated, while applying the 'Use Column Default' Transformer to all Generated (non-Identity)columns."
              body={handleAlertDescriptionBody()}
              containerClassName="max-w-xl"
              onConfirm={() => {
                const formMappings = form.getValues('mappings');
                const defaultTransformerValues =
                  defaultTransformerForm.getValues();
                formMappings.forEach((fm, idx) => {
                  // skips setting the default transformer if the user has already set the transformer
                  if (
                    fm.transformer.source != 0 &&
                    !defaultTransformerValues.overrideTransformers
                  ) {
                    return;
                  } else {
                    const colkey = {
                      schema: fm.schema,
                      table: fm.table,
                      column: fm.column,
                    };
                    const isGenerated =
                      constraintHandler.getIsGenerated(colkey);
                    const identityType =
                      constraintHandler.getIdentityType(colkey);
                    const newJm =
                      isGenerated && !identityType
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
                  }
                });
                form.trigger('mappings'); // trigger validation after bulk updating the selected form options
              }}
            />
          )}
          <SchemaTableViewOptions table={table} />
          <div>
            <Button type="button" onClick={() => onViewChange('table')}>
              Table View
            </Button>
            <Button type="button" onClick={() => onViewChange('json')}>
              JSON View
            </Button>
          </div>
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
