import { ColumnDef } from '@tanstack/react-table';
import { HTMLProps, useEffect, useRef } from 'react';
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

function IndeterminateCheckbox({
  indeterminate,
  className = 'flex w-4 h-4',
  ...rest
}: { indeterminate?: boolean } & HTMLProps<HTMLInputElement>) {
  const ref = useRef<HTMLInputElement>(null!);

  useEffect(() => {
    if (typeof indeterminate === 'boolean') {
      ref.current.indeterminate = !rest.checked && indeterminate;
    }
  }, [ref, indeterminate, rest.checked]);

  return (
    <input
      type="checkbox"
      ref={ref}
      className={className + ' cursor-pointer '}
      {...rest}
    />
  );
}
