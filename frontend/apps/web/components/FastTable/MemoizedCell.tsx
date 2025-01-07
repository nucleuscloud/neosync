import { Cell, flexRender } from '@tanstack/react-table';
import { memo, ReactNode } from 'react';

interface Props<TData> {
  cell: Cell<TData, unknown>;
}

function InnerCell<TData>(props: Props<TData>): ReactNode {
  const { cell } = props;
  return flexRender(cell.column.columnDef.cell, cell.getContext());
}

function shouldReRender<TData>(
  prev: Props<TData>,
  next: Props<TData>
): boolean {
  const prevValue = prev.cell.getValue();
  const nextValue = next.cell.getValue();

  if (
    // todo: maybe these should be passed in as props so they are configurable by the caller
    prev.cell.column.id === 'isSelected' ||
    prev.cell.column.id === 'actions'
  ) {
    // Always re-render checkbox cells as getIsSelected() is always the same for both
    return false;
  }

  // For other columns, just compare the values
  return prevValue === nextValue;
}

const MemoizedCell = memo(InnerCell, shouldReRender);
MemoizedCell.displayName = 'MemoizedCell';
export default MemoizedCell;
