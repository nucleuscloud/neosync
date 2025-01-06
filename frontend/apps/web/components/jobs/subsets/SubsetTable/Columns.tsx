import TruncatedText from '@/components/TruncatedText';
import { ColumnDef, createColumnHelper } from '@tanstack/react-table';
import { SchemaColumnHeader } from '../../SchemaTable/SchemaColumnHeader';

interface SubsetTableRow {
  schema: string;
  table: string;
  where?: string;
  isRootTable?: boolean;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function getColumns(): ColumnDef<SubsetTableRow, any>[] {
  const columnHelper = createColumnHelper<SubsetTableRow>();

  const schemaColumn = columnHelper.accessor('schema', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Schema" />;
    },
    cell({ getValue }) {
      return <TruncatedText text={getValue()} />;
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

  const isRootTableColumn = columnHelper.accessor('isRootTable', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Root Table" />;
    },
    cell({ getValue }) {
      // todo
      return <div>{getValue() ? 'Yes' : 'No'}</div>;
    },
  });

  const whereColumn = columnHelper.accessor('where', {
    header({ column }) {
      return <SchemaColumnHeader column={column} title="Subset Filters" />;
    },
    cell({ getValue }) {
      <div className="flex justify-center">
        <span className="truncate font-medium max-w-[200px] inline-block">
          {getValue<boolean>() && (
            <pre className="bg-gray-100 rounded border border-gray-300 text-xs px-2 dark:bg-transparent dark:border dark:border-gray-700 whitespace-nowrap overflow-hidden text-ellipsis max-w-[100%]">
              {getValue<boolean>()}
            </pre>
          )}
        </span>
      </div>;
    },
    size: 250,
  });

  const actionsColumn = columnHelper.display({
    id: 'actions',
    enableSorting: false,
    enableColumnFilter: false,
    header: ({ column }) => (
      <SchemaColumnHeader column={column} title="Actions" />
    ),
    cell: ({}) => {
      return <div>Actions</div>;
      // return (
      //   <div className="flex gap-2">
      //     <EditAction onClick={() => onEdit(schema, table)} />
      //     <ResetAction
      //       onClick={() => onReset(schema, table)}
      //       isDisabled={!hasLocalChange(schema, table)}
      //     />
      //   </div>
      // );
    },
  });

  return [
    schemaColumn,
    tableColumn,
    isRootTableColumn,
    whereColumn,
    actionsColumn,
  ];
}

export const SUBSET_TABLE_COLUMNS = getColumns();
