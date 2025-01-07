import FastTable from '@/components/FastTable/FastTable';
import { CardDescription, CardTitle } from '@/components/ui/card';
import { Transformer } from '@/shared/transformers';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import { JobMapping } from '@neosync/sdk';
import {
  ColumnDef,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  Row,
  RowData,
  useReactTable,
} from '@tanstack/react-table';
import { ReactElement } from 'react';
import { GoWorkflow } from 'react-icons/go';
import { ImportMappingsConfig } from '../SchemaTable/ImportJobMappingsButton';
import { SchemaTableToolbar } from '../SchemaTable/SchemaTableToolBar';
import { TransformerResult } from '../SchemaTable/transformer-handler';

interface Props<TData, TValue> {
  data: TData[];
  columns: ColumnDef<TData, TValue>[];
  onTransformerUpdate(index: number, config: JobMappingTransformerForm): void;
  getAvailableTransformers(index: number): TransformerResult;
  getTransformerFromField(index: number): Transformer;

  onTransformerBulkUpdate(
    indices: number[],
    config: JobMappingTransformerForm
  ): void;
  getAvalableTransformersForBulk(rows: Row<TData>[]): TransformerResult;
  getTransformerFromFieldValue(value: JobMappingTransformerForm): Transformer;

  isApplyDefaultTransformerButtonDisabled: boolean;
  onApplyDefaultClick(override: boolean): void;

  onExportMappingsClick(selected: Row<TData>[], shouldFormat: boolean): void;
  onImportMappingsClick(
    jobmappings: JobMapping[],
    config: ImportMappingsConfig
  ): void;

  onDuplicateRow(index: number): void;
  onDeleteRow(index: number): void;
  canRenameColumn(index: number, newColumn: string): boolean;
  onRowUpdate(index: number, newValue: TData): void;
  getAvailableCollectionsByRow(index: number): string[];
}

declare module '@tanstack/react-table' {
  interface TableMeta<TData extends RowData> {
    jmTable?: {
      onTransformerUpdate(
        rowIndex: number,
        transformer: JobMappingTransformerForm
      ): void;
      getAvailableTransformers(rowIndex: number): TransformerResult;
      getTransformerFromField(index: number): Transformer;

      onDuplicateRow(rowIndex: number): void;
      onDeleteRow(rowIndex: number): void;
      canRenameColumn(rowIndex: number, newColumn: string): boolean;
      onRowUpdate(rowIndex: number, newValue: TData): void;
      // Returns the available schema.table list
      getAvailableCollectionsByRow(rowIndex: number): string[];
    };
  }
}

export default function JobMappingTable<TData, TValue>(
  props: Props<TData, TValue>
): ReactElement {
  const {
    data,
    columns,
    onTransformerUpdate,
    getAvailableTransformers,
    getTransformerFromField,
    onExportMappingsClick,
    onImportMappingsClick,
    getAvalableTransformersForBulk,
    getTransformerFromFieldValue,
    isApplyDefaultTransformerButtonDisabled,
    onApplyDefaultClick,
    onTransformerBulkUpdate,
    onDeleteRow,
    onDuplicateRow,
    canRenameColumn,
    onRowUpdate,
    getAvailableCollectionsByRow,
  } = props;

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    meta: {
      jmTable: {
        onTransformerUpdate,
        getAvailableTransformers,
        getTransformerFromField,
        onDeleteRow,
        onDuplicateRow,
        canRenameColumn,
        onRowUpdate,
        getAvailableCollectionsByRow,
      },
    },
  });

  return (
    <div>
      <div className="flex flex-row items-center gap-2 pt-4 ">
        <div className="flex">
          <GoWorkflow className="h-4 w-4" />
        </div>
        <CardTitle>Transformer Mapping</CardTitle>
      </div>
      <CardDescription className="pt-2">
        Map Transformers to every column below.
      </CardDescription>
      <div className="z-50 pt-4">
        <SchemaTableToolbar<TData>
          table={table}
          displayApplyDefaultTransformersButton={true}
          isApplyDefaultButtonDisabled={isApplyDefaultTransformerButtonDisabled}
          getAllowedTransformers={getAvalableTransformersForBulk}
          getTransformerFromField={getTransformerFromFieldValue}
          onApplyDefaultClick={onApplyDefaultClick}
          onBulkUpdate={onTransformerBulkUpdate}
          onExportMappingsClick={(shouldFormat) =>
            onExportMappingsClick(
              table.getSelectedRowModel().rows,
              shouldFormat
            )
          }
          onImportMappingsClick={onImportMappingsClick}
        />
      </div>

      <FastTable table={table} estimateRowSize={() => 53} rowOverscan={50} />

      <div className="text-xs text-gray-600 dark:text-gray-400 pt-4">
        Total rows: ({getFormattedCount(data.length)}) Rows visible: (
        {getFormattedCount(table.getRowModel().rows.length)})
      </div>
    </div>
  );
}

const US_NUMBER_FORMAT = new Intl.NumberFormat('en-US');
function getFormattedCount(count: number): string {
  return US_NUMBER_FORMAT.format(count);
}
