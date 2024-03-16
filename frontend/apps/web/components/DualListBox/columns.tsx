import { ColumnDef } from '@tanstack/react-table';
import { HTMLProps, useEffect, useRef } from 'react';
import ColumnHeader from './ColumnHeader';

export interface Row {
  value: string;
}

interface ListBoxColumnProps {
  title: string;
}

export function getListBoxColumns(props: ListBoxColumnProps): ColumnDef<Row>[] {
  const { title } = props;
  return [
    {
      accessorKey: 'isSelected',
      header: ({ table }) => (
        <IndeterminateCheckbox
          {...{
            checked: table.getIsAllRowsSelected(),
            indeterminate: table.getIsSomeRowsSelected(),
            onChange: table.getToggleAllRowsSelectedHandler(),
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
      cell: ({ row }) => {
        return (
          <label
            htmlFor={row.getValue('value')}
            className="max-w-[500px] truncate font-medium cursor-pointer"
          >
            {row.getValue('value')}
          </label>
        );
      },
    },
  ];
}

function IndeterminateCheckbox({
  indeterminate,
  className = 'w-4 h-4',
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
