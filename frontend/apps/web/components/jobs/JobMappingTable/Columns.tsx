import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import TruncatedText from '@/components/TruncatedText';
import {
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import { SystemTransformer } from '@neosync/sdk';
import { ColumnDef, createColumnHelper } from '@tanstack/react-table';
import { DataTableRowActions } from '../NosqlTable/data-table-row-actions';
import EditCollection from '../NosqlTable/EditCollection';
import EditDocumentKey from '../NosqlTable/EditDocumentKey';
import { SchemaColumnHeader } from '../SchemaTable/SchemaColumnHeader';
import TransformerSelect from '../SchemaTable/TransformerSelect';
import IndeterminateCheckbox from './IndeterminateCheckbox';

export interface JobMappingRow {
  schema: string;
  table: string;
  column: string;
  constraints: string;
  dataType: string;
  isNullable: string;
  attributes: string;
  transformer: JobMappingTransformerForm;
}

export interface NosqlJobMappingRow {
  collection: string; // combined schema.table
  column: string;
  transformer: JobMappingTransformerForm;
}

export function getJobMappingColumns(): ColumnDef<JobMappingRow, any>[] {
  const columnHelper = createColumnHelper<JobMappingRow>();

  const checkboxColumn = columnHelper.display({
    id: 'isSelected',
    header({ table }) {
      return (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
          }}
        />
      );
    },
    cell({ row }) {
      return (
        <div>
          <IndeterminateCheckbox
            {...{
              checked: row.getIsSelected(),
              indeterminate: row.getIsSomeSelected(),
              onChange: row.getToggleSelectedHandler(),
            }}
          />
        </div>
      );
    },
  });

  const schemaColumn = columnHelper.accessor('schema', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Schema" />;
    },
  });

  const tableColumn = columnHelper.accessor('table', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Table" />;
    },
  });

  const columnColumn = columnHelper.accessor('column', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Column" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} />;
    },
  });

  const dataTypeColumn = columnHelper.accessor('dataType', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Data Type" />;
    },
    cell({ getValue }) {
      return <span>{getValue()}</span>;
    },
  });

  const isNullableColumn = columnHelper.accessor('isNullable', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Nullable" />;
    },
    cell({ getValue }) {
      return <span>{getValue()}</span>;
    },
  });

  const constraintColumn = columnHelper.accessor('constraints', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Constraints" />;
    },
    cell({ getValue }) {
      return <span>{getValue()}</span>;
    },
  });

  const attributeColumn = columnHelper.accessor('attributes', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Attributes" />;
    },
    cell({ getValue }) {
      return <span>{getValue()}</span>;
    },
  });

  const transformerColumn = columnHelper.accessor('transformer', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Transformer" />;
    },
    cell({ getValue, table, row }) {
      const transformer =
        table.options.meta?.getTransformerFromField(row.index) ??
        new SystemTransformer();
      return (
        <div className="flex flex-row gap-2">
          <div>
            <TransformerSelect
              getTransformers={() =>
                table.options.meta?.getAvailableTransformers(row.index) ?? {
                  system: [],
                  userDefined: [],
                }
              }
              buttonText={getTransformerSelectButtonText(transformer)}
              buttonClassName="w-[175px]"
              value={getValue()}
              onSelect={(updatedValue) =>
                table.options.meta?.onTransformerUpdate(row.index, updatedValue)
              }
              disabled={false}
            />
          </div>
          <div>
            <EditTransformerOptions
              transformer={transformer}
              value={getValue()}
              onSubmit={(updatedValue) => {
                table.options.meta?.onTransformerUpdate(
                  row.index,
                  updatedValue
                );
              }}
              disabled={isInvalidTransformer(transformer)}
            />
          </div>
        </div>
      );
    },
  });

  return [
    checkboxColumn,
    schemaColumn,
    tableColumn,
    columnColumn,
    dataTypeColumn,
    isNullableColumn,
    constraintColumn,
    attributeColumn,
    transformerColumn,
  ];
}

export function getNosqlJobMappingColumns(): ColumnDef<
  NosqlJobMappingRow,
  any
>[] {
  const columnHelper = createColumnHelper<NosqlJobMappingRow>();

  const checkboxColumn = columnHelper.display({
    id: 'isSelected',
    header({ table }) {
      return (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
          }}
        />
      );
    },
    cell({ row }) {
      return (
        <div>
          <IndeterminateCheckbox
            {...{
              checked: row.getIsSelected(),
              indeterminate: row.getIsSomeSelected(),
              onChange: row.getToggleSelectedHandler(),
            }}
          />
        </div>
      );
    },
  });

  const collectionColumn = columnHelper.accessor('collection', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Collection" />;
    },
    cell({ getValue, table, row }) {
      return (
        <EditCollection
          text={getValue()}
          collections={
            table.options.meta?.getAvailableCollectionsByRow(row.index) ?? []
          }
          onEdit={(updatedValue) => {
            if (table.options.meta?.onRowUpdate) {
              table.options.meta.onRowUpdate(row.index, {
                ...row.original,
                collection: updatedValue.collection,
              });
            }
          }}
        />
      );
    },
  });

  const columnColumn = columnHelper.accessor('column', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Document Key" />;
    },
    cell({ getValue, table, row }) {
      return (
        <EditDocumentKey
          text={getValue()}
          isDuplicate={(newValue, currValue) => {
            return (
              newValue !== currValue &&
              (table.options.meta?.canRenameColumn(row.index, newValue) ??
                false)
            );
          }}
          onEdit={(updatedValue) => {
            if (table.options.meta?.onRowUpdate) {
              table.options.meta.onRowUpdate(row.index, {
                ...row.original,
                column: updatedValue.column,
              });
            }
          }}
        />
      );
    },
  });

  const transformerColumn = columnHelper.accessor('transformer', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Transformer" />;
    },
    cell({ getValue, table, row }) {
      const transformer =
        table.options.meta?.getTransformerFromField(row.index) ??
        new SystemTransformer();
      return (
        <div className="flex flex-row gap-2">
          <div>
            <TransformerSelect
              getTransformers={() =>
                table.options.meta?.getAvailableTransformers(row.index) ?? {
                  system: [],
                  userDefined: [],
                }
              }
              buttonText={getTransformerSelectButtonText(transformer)}
              buttonClassName="w-[175px]"
              value={getValue()}
              onSelect={(updatedValue) =>
                table.options.meta?.onTransformerUpdate(row.index, updatedValue)
              }
              disabled={false}
            />
          </div>
          <div>
            <EditTransformerOptions
              transformer={transformer}
              value={getValue()}
              onSubmit={(updatedValue) => {
                table.options.meta?.onTransformerUpdate(
                  row.index,
                  updatedValue
                );
              }}
              disabled={isInvalidTransformer(transformer)}
            />
          </div>
        </div>
      );
    },
  });

  const actionsColumn = columnHelper.display({
    id: 'actions',
    header({}) {
      return <p>Actions</p>;
    },
    cell({ row, table }) {
      return (
        <DataTableRowActions
          row={row}
          onDuplicate={() => table.options.meta?.onDuplicateRow(row.index)}
          onDelete={() => table.options.meta?.onDeleteRow(row.index)}
        />
      );
    },
  });

  return [
    checkboxColumn,
    collectionColumn,
    columnColumn,
    transformerColumn,
    actionsColumn,
  ];
}

export const SQL_COLUMNS = getJobMappingColumns();

export const NOSQL_COLUMNS = getNosqlJobMappingColumns();
