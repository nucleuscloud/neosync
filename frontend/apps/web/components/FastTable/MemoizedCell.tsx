import { Cell, flexRender } from '@tanstack/react-table';
import { memo } from 'react';

const MemoizedCell = memo(
  <TData,>({ cell }: { cell: Cell<TData, unknown> }) =>
    flexRender(cell.column.columnDef.cell, cell.getContext()),
  (prev, next) => {
    const prevValue = prev.cell.getValue();
    const nextValue = next.cell.getValue();

    if (
      prev.cell.column.id === 'isSelected' ||
      prev.cell.column.id === 'actions'
    ) {
      // Always re-render checkbox cells as getIsSelected() is always the same for both
      return false;
    }

    // For other columns, just compare the values
    return prevValue === nextValue;
  }
);
MemoizedCell.displayName = 'MemoizedCell';

export default MemoizedCell;
