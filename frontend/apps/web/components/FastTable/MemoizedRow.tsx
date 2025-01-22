import { cn } from '@/libs/utils';
import { Cell, Row } from '@tanstack/react-table';
import { VirtualItem } from '@tanstack/react-virtual';
import { memo, ReactNode } from 'react';
import { TableRow } from '../ui/table';
import MemoizedCell from './MemoizedCell';

interface Props<TData> {
  row: Row<TData>;
  virtualRow: VirtualItem;
  selected: boolean;
  tableRowClassName?: string;
  disableTdWidth?: boolean;
}

function InnerRow<TData>(props: Props<TData>): ReactNode {
  const { row, virtualRow, tableRowClassName, disableTdWidth } = props;
  return (
    <TableRow
      key={row.id}
      style={{
        transform: `translateY(${virtualRow.start}px)`,
        height: `${virtualRow.size}px`,
      }}
      className={cn(
        'items-center flex absolute w-full justify-between px-2 gap-0 space-x-0',
        tableRowClassName
      )}
    >
      {row.getVisibleCells().map((cell) => (
        <td
          key={cell.id}
          className="py-2"
          style={{
            minWidth: cell.column.getSize(),
            width: disableTdWidth
              ? undefined
              : cell.column.columnDef.id === 'isSelected'
                ? '20px'
                : '187px',
          }}
        >
          {/* For some reason TS can't figure out how to type the incoming cell dynamically as Cell<TData, unknown>
              so we have to cast it here */}
          <MemoizedCell cell={cell as Cell<unknown, unknown>} />
        </td>
      ))}
    </TableRow>
  );
}

function shouldReRender<TData>(
  prev: Props<TData>,
  next: Props<TData>
): boolean {
  if (
    prev.tableRowClassName !== next.tableRowClassName &&
    (prev.tableRowClassName !== undefined ||
      next.tableRowClassName !== undefined)
  ) {
    return false;
  }
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
    const prevCell = prevCells[i];
    const nextCell = nextCells[i];

    if (prevCell.id !== nextCell.id) {
      return false;
    }

    // For accessor columns, compare values
    if (prevCell.getValue() !== nextCell.getValue()) {
      return false;
    }
  }

  // If no differences are found, skip re-render
  return true;
}

const MemoizedRow = memo(InnerRow, shouldReRender);
MemoizedRow.displayName = 'MemoizedRow';
export default MemoizedRow;
