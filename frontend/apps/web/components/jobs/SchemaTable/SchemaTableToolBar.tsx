'use client';

import { Row, Table } from '@tanstack/react-table';

import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import ButtonText from '@/components/ButtonText';
import FormErrorMessage from '@/components/FormErrorMessage';
import { Button } from '@/components/ui/button';
import { cn } from '@/libs/utils';
import { isSystemTransformer, Transformer } from '@/shared/transformers';
import {
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import {
  convertJobMappingTransformerToForm,
  JobMappingTransformerForm,
} from '@/yup-validations/jobs';
import {
  JobMapping,
  JobMappingTransformer,
  SystemTransformer,
  UserDefinedTransformer,
} from '@neosync/sdk';
import { CheckIcon, Cross2Icon } from '@radix-ui/react-icons';
import { useState } from 'react';
import ApplyDefaultTransformersButton from './ApplyDefaultTransformersButton';
import ExportJobMappingsButton from './ExportJobMappingsButton';
import ImportJobMappingsButton, {
  ImportMappingsConfig,
} from './ImportJobMappingsButton';
import { SchemaTableViewOptions } from './SchemaTableViewOptions';
import TransformerSelect from './TransformerSelect';
import { TransformerResult } from './transformer-handler';

interface DataTableToolbarProps<TData> {
  table: Table<TData>;
  getAllowedTransformers(rows: Row<TData>[]): TransformerResult;
  getTransformerFromField(selected: JobMappingTransformerForm): Transformer;
  onBulkUpdate(indices: number[], value: JobMappingTransformerForm): void;
  onExportMappingsClick(shouldFormat: boolean): void;
  onImportMappingsClick(
    jobmappings: JobMapping[],
    config: ImportMappingsConfig
  ): void;
  displayApplyDefaultTransformersButton: boolean;
  isApplyDefaultButtonDisabled: boolean;
  onApplyDefaultClick(override: boolean): void;
}

export function SchemaTableToolbar<TData>({
  table,
  onExportMappingsClick,
  onImportMappingsClick,
  getAllowedTransformers,
  getTransformerFromField,
  onBulkUpdate,
  displayApplyDefaultTransformersButton,
  isApplyDefaultButtonDisabled,
  onApplyDefaultClick,
}: DataTableToolbarProps<TData>) {
  const isFiltered = table.getState().columnFilters.length > 0;
  const hasSelectedRows = Object.values(table.getState().rowSelection).some(
    (value) => value
  );

  const [bulkTransformer, setBulkTransformer] =
    useState<JobMappingTransformerForm>(
      convertJobMappingTransformerToForm(new JobMappingTransformer())
    );

  const transformer = getTransformerFromField(bulkTransformer);
  const allowedTransformers = getAllowedTransformers(
    table.getSelectedRowModel().rows
  );
  const isBulkApplyDisabled =
    !bulkTransformer ||
    !hasSelectedRows ||
    !isTransformerAllowed(allowedTransformers, transformer);

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
              const rowIndices = table
                .getSelectedRowModel()
                .rows.map((r) => r.index);
              if (rowIndices.length === 0) {
                return;
              }
              onBulkUpdate(rowIndices, bulkTransformer);
              // table.getSelectedRowModel().rows.forEach((r) => {
              //   form.setValue(
              //     `mappings.${r.index}.transformer`,
              //     bulkTransformer,
              //     {
              //       shouldDirty: true,
              //       shouldTouch: true,
              //       shouldValidate: false, // this is really expensive, see the trigger call below
              //     }
              //   );
              // });
              setBulkTransformer(
                convertJobMappingTransformerToForm(new JobMappingTransformer())
              );
              // form.trigger('mappings'); // trigger validation after bulk updating the selected form options
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
          {displayApplyDefaultTransformersButton && (
            <ApplyDefaultTransformersButton
              // isDisabled={form.watch('mappings').length === 0}
              isDisabled={isApplyDefaultButtonDisabled}
              onClick={onApplyDefaultClick}
            />
          )}
          <ImportJobMappingsButton onImport={onImportMappingsClick} />
          <ExportJobMappingsButton
            onClick={onExportMappingsClick}
            count={table.getSelectedRowModel().rows.length}
          />
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
