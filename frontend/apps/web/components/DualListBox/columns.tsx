import { ColumnDef } from '@tanstack/react-table';
import IndeterminateCheckbox from '../jobs/JobMappingTable/IndeterminateCheckbox';
import ColumnHeader from './ColumnHeader';

export type Mode = 'single' | 'many';

export interface Row {
  value: string;
}

interface ListBoxColumnProps {
  title: string;
  mode?: Mode;
}

export function getListBoxColumns(props: ListBoxColumnProps): ColumnDef<Row>[] {
  const { title, mode = 'many' } = props;
  return [
    {
      accessorKey: 'isSelected',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
            className: mode === 'single' ? 'flex hidden' : 'flex w-4 h-4',
          }}
        />
      ),
      cell: ({ row }) => (
        <div>
          <IndeterminateCheckbox
            {...{
              checked: row.getIsSelected(),
              indeterminate: row.getIsSomeSelected(),
              onChange: row.getToggleSelectedHandler(),
              id: row.getValue('value'),
            }}
          />
        </div>
      ),
      size: 30,
    },
    {
      accessorKey: 'value',
      header: ({ column }) => <ColumnHeader column={column} title={title} />,
      cell: ({ getValue }) => {
        return (
          <label className="font-medium cursor-pointer break-all whitespace-normal">
            {getValue<string>()}
          </label>
        );
      },
    },
  ];
}
