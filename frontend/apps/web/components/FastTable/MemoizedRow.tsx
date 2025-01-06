import { Cell, Row } from '@tanstack/react-table';
import { VirtualItem } from '@tanstack/react-virtual';
import { memo } from 'react';
import { TableRow } from '../ui/table';
import MemoizedCell from './MemoizedCell';

const MemoizedRow = memo(
  <TData,>({
    row,
    virtualRow,
  }: {
    row: Row<TData>;
    virtualRow: VirtualItem;
    selected: boolean;
  }) => {
    return (
      <TableRow
        key={row.id}
        style={{
          transform: `translateY(${virtualRow.start}px)`,
          height: `${virtualRow.size}px`,
        }}
        className="items-center flex absolute w-full justify-between px-2 gap-0 space-x-0"
      >
        {row.getVisibleCells().map((cell) => (
          <td
            key={cell.id}
            className="py-2"
            style={{
              minWidth: cell.column.getSize(),
              width:
                cell.column.columnDef.id != 'isSelected' ? '187px' : '20px',
            }}
          >
            <MemoizedCell cell={cell as Cell<unknown, unknown>} />
          </td>
        ))}
      </TableRow>
    );
  },
  (prev, next) => {
    // Compare virtualRow properties
    if (
      prev.virtualRow.start !== next.virtualRow.start ||
      prev.virtualRow.size !== next.virtualRow.size
    ) {
      return false;
    }

    // Compare row.id
    if (prev.row.id !== next.row.id) {
      return false;
    }

    // Check row selection state for "isSelected"
    if (prev.selected !== next.selected) {
      return false;
    }

    // Check if visible cells or their values have changed
    const prevCells = prev.row.getVisibleCells();
    const nextCells = next.row.getVisibleCells();

    if (prevCells.length !== nextCells.length) {
      return false;
    }

    for (let i = 0; i < prevCells.length; i++) {
      if (
        prevCells[i].id !== nextCells[i].id ||
        prevCells[i].getValue() !== nextCells[i].getValue()
      ) {
        return false;
      }
    }

    // If no differences are found, skip re-render
    return true;
  }
);
MemoizedRow.displayName = 'MemoizedRow';

export default MemoizedRow;
