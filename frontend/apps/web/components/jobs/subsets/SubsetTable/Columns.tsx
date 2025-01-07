import TruncatedText from '@/components/TruncatedText';
import { ColumnDef, createColumnHelper } from '@tanstack/react-table';
import IndeterminateCheckbox from '../../JobMappingTable/IndeterminateCheckbox';
import { SchemaColumnHeader } from '../../SchemaTable/SchemaColumnHeader';
import ActionsCell from './ActionsCell';
import RootTableCell from './RootTableCell';

export interface SubsetTableRow {
  schema: string;
  table: string;
  where?: string;
  isRootTable: boolean;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type ColumnTValue = any;

function getColumns(): ColumnDef<SubsetTableRow, ColumnTValue>[] {
  const columnHelper = createColumnHelper<SubsetTableRow>();

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
        <IndeterminateCheckbox
          {...{
            checked: row.getIsSelected(),
            indeterminate: row.getIsSomeSelected(),
            onChange: row.getToggleSelectedHandler(),
          }}
        />
      );
    },
    maxSize: 20,
  });

  const schemaColumn = columnHelper.accessor('schema', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Schema" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} maxWidth={150} />;
    },
  });

  const tableColumn = columnHelper.accessor('table', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Table" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} maxWidth={150} />;
    },
  });

  const isRootTableColumn = columnHelper.accessor(
    (row) => (row.isRootTable ? 'Root' : ''),
    {
      id: 'isRootTable',
      header({ column }) {
        return <SchemaColumnHeader column={column} title="Root Table" />;
      },
      cell({ getValue }) {
        return <RootTableCell isRootTable={!!getValue()} />;
      },
    }
  );

  const whereColumn = columnHelper.accessor((row) => row.where ?? '', {
    id: 'where',
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Subset Filters" />;
    },
    cell({ getValue }) {
      const whereValue = getValue();
      return (
        <div className="flex justify-center">
          <span className="truncate font-medium max-w-[200px] inline-block">
            {!!whereValue && (
              <pre className="bg-gray-100 rounded border border-gray-300 text-xs px-2 dark:bg-transparent dark:border dark:border-gray-700 whitespace-nowrap overflow-hidden text-ellipsis max-w-[100%]">
                {whereValue}
              </pre>
            )}
          </span>
        </div>
      );
    },
  });

  const actionsColumn = columnHelper.display({
    id: 'actions',
    enableSorting: false,
    enableColumnFilter: false,
    header: ({ column }) => (
      <SchemaColumnHeader column={column} title="Actions" />
    ),
    cell: ({ row, table }) => {
      const isResetDisabled = !table.options.meta?.subsetTable?.hasLocalChange(
        row.index,
        row.original.schema,
        row.original.table
      );

      return (
        <ActionsCell
          onEdit={() =>
            table.options.meta?.subsetTable?.onEdit(
              row.index,
              row.original.schema,
              row.original.table
            )
          }
          onReset={() =>
            table.options.meta?.subsetTable?.onReset(
              row.index,
              row.original.schema,
              row.original.table
            )
          }
          isResetDisabled={isResetDisabled}
        />
      );
    },
  });

  return [
    checkboxColumn,
    schemaColumn,
    tableColumn,
    isRootTableColumn,
    whereColumn,
    actionsColumn,
  ];
}

export const SUBSET_TABLE_COLUMNS = getColumns();
