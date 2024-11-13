import EditTransformerOptions from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import TruncatedText from '@/components/TruncatedText';
import {
  getTransformerSelectButtonText,
  isInvalidTransformer,
} from '@/util/util';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import { SystemTransformer } from '@neosync/sdk';
import { ColumnDef, createColumnHelper } from '@tanstack/react-table';
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

export const COLUMNS = getJobMappingColumns();
