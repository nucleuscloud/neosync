import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import TruncatedText from '@/components/TruncatedText';
import { Badge } from '@/components/ui/badge';
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
import AttributesCell from './AttributesCell';
import ConstraintsCell from './ConstraintsCell';
import DataTypeCell from './DataTypeCell';
import IndeterminateCheckbox from './IndeterminateCheckbox';

export interface JobMappingRow {
  schema: string;
  table: string;
  column: string;
  constraints: RowConstraint;
  dataType: string;
  isNullable: boolean;
  attributes: RowAttribute;
  transformer: JobMappingTransformerForm;
}

interface RowAttribute {
  value: string; // accessor fn value for search

  generatedType: string | undefined;
  identityType: string | undefined;
}
interface RowConstraint {
  value: string; // accessor fn value for search
  isPrimaryKey: boolean;
  foreignKey: [boolean, string[]];
  virtualForeignKey: [boolean, string[]];
  isUnique: boolean;
}

export interface NosqlJobMappingRow {
  collection: string; // combined schema.table
  column: string;
  transformer: JobMappingTransformerForm;
}

function getJobMappingColumns(): ColumnDef<JobMappingRow, string>[] {
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
    maxSize: 30,
  });

  const schemaColumn = columnHelper.accessor('schema', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Schema" />;
    },
  });

  const tableColumn = columnHelper.accessor((row) => row.table, {
    id: 'table',
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
      return <DataTypeCell value={getValue()} />;
    },
  });

  const isNullableColumn = columnHelper.accessor(
    (row) => (row.isNullable ? 'Yes' : 'No') as string,
    {
      id: 'isNullable',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Nullable" />;
      },
      cell({ getValue }) {
        return (
          <span className="max-w-[500px] truncate font-medium">
            <Badge variant="outline">{getValue()}</Badge>
          </span>
        );
      },
    }
  );

  const constraintColumn = columnHelper.accessor(
    (row) => row.constraints.value,
    {
      id: 'constraints',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Constraints" />;
      },
      cell({ row }) {
        const constraints = row.original.constraints;
        return (
          <ConstraintsCell
            isPrimaryKey={constraints.isPrimaryKey}
            foreignKey={constraints.foreignKey}
            virtualForeignKey={constraints.virtualForeignKey}
            isUnique={constraints.isUnique}
          />
        );
      },
    }
  );

  const attributeColumn = columnHelper.accessor((row) => row.attributes.value, {
    id: 'attributeValues',
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Attributes" />;
    },
    cell({ row }) {
      const val = row.original.attributes;
      return (
        <AttributesCell
          generatedType={val.generatedType}
          identityType={val.identityType}
          value={val.value}
        />
      );
    },
  });

  const transformerColumn = columnHelper.accessor(
    (row) => {
      if (row.transformer.config.case) {
        return row.transformer.config.case.toLowerCase();
      }
      return 'select transformer';
    },
    {
      id: 'transformer',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Transformer" />;
      },
      cell({ table, row }) {
        const transformer =
          table.options.meta?.getTransformerFromField(row.index) ??
          new SystemTransformer();
        const transformerForm = row.original.transformer;
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
                value={transformerForm}
                onSelect={(updatedValue) =>
                  table.options.meta?.onTransformerUpdate(
                    row.index,
                    updatedValue
                  )
                }
                disabled={false}
              />
            </div>
            <div>
              <EditTransformerOptions
                transformer={transformer}
                value={transformerForm}
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
    }
  );

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

function getNosqlJobMappingColumns(): ColumnDef<NosqlJobMappingRow, string>[] {
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

  const transformerColumn = columnHelper.accessor(
    (row) => {
      if (row.transformer.config.case) {
        return row.transformer.config.case.toLowerCase();
      }
      return 'select transformer';
    },
    {
      id: 'transformer',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Transformer" />;
      },
      cell({ getValue, table, row }) {
        const transformer =
          table.options.meta?.getTransformerFromField(row.index) ??
          new SystemTransformer();
        const transformerForm = row.original.transformer;
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
                value={transformerForm}
                onSelect={(updatedValue) =>
                  table.options.meta?.onTransformerUpdate(
                    row.index,
                    updatedValue
                  )
                }
                disabled={false}
              />
            </div>
            <div>
              <EditTransformerOptions
                transformer={transformer}
                value={transformerForm}
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
    }
  );

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
