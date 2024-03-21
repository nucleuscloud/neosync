import { Button } from '@/components/ui/button';
import { cn } from '@/libs/utils';
import {
  ArrowDownIcon,
  ArrowLeftIcon,
  ArrowRightIcon,
  ArrowUpIcon,
  DoubleArrowDownIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
  DoubleArrowUpIcon,
} from '@radix-ui/react-icons';
import { RowSelectionState } from '@tanstack/react-table';
import { ReactElement, useMemo, useState } from 'react';
import ListBox from '../ListBox/ListBox';
import { Mode, Row, getListBoxColumns } from './columns';

export interface Option {
  value: string;
  // label: string;
}
export type Action = 'add' | 'add-all' | 'remove' | 'remove-all';
interface Props {
  options: Option[];
  selected: Set<string>;
  onChange(value: Set<string>, action: Action): void;
  mode?: Mode;
}

export default function DualListBox(props: Props): ReactElement {
  const { options, selected, onChange, mode = 'many' } = props;

  const [leftSelected, setLeftSelected] = useState<RowSelectionState>({});
  const [rightSelected, setRightSelected] = useState<RowSelectionState>({});

  const leftCols = useMemo(
    () => getListBoxColumns({ title: 'Source', mode }),
    [mode]
  );
  const rightCols = useMemo(
    () => getListBoxColumns({ title: 'Destination', mode }),
    [mode]
  );

  const leftData = options
    .filter((value) => !selected.has(value.value))
    .map((value): Row => ({ value: value.value }));
  const rightData = options
    .filter((value) => selected.has(value.value))
    .map((value): Row => ({ value: value.value }));

  return (
    <div className="flex gap-3 flex-col md:flex-row">
      <div className="flex flex-1 border border-gray-300 overflow-hidden rounded-lg">
        <ListBox
          columns={leftCols}
          data={leftData}
          onRowSelectionChange={setLeftSelected}
          rowSelection={leftSelected}
          mode={mode}
        />
      </div>
      <div className="flex flex-row md:flex-col justify-center gap-2">
        <div className={cn(mode === 'single' ? 'hidden' : null)}>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(
                new Set(options.map((option) => option.value)),
                'add-all'
              );
              setLeftSelected({});
            }}
          >
            <DoubleArrowRightIcon className="hidden md:block" />
            <DoubleArrowDownIcon className="block md:hidden" />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              if (mode === 'single' && selected.size > 0) {
                return;
              }
              const newSet = new Set(selected);
              Object.entries(leftSelected).forEach(([key, isSelected]) => {
                if (isSelected) {
                  newSet.add(leftData[parseInt(key, 10)].value);
                }
              });
              onChange(newSet, 'add');
              setLeftSelected({});
            }}
          >
            <ArrowRightIcon className="hidden md:block" />
            <ArrowDownIcon className="block md:hidden" />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              const newSet = new Set(selected);
              Object.entries(rightSelected).forEach(([key, isSelected]) => {
                if (isSelected) {
                  newSet.delete(rightData[parseInt(key, 10)].value);
                }
              });
              onChange(newSet, 'remove');
              setRightSelected({});
            }}
          >
            <ArrowLeftIcon className="hidden md:block" />
            <ArrowUpIcon className="block md:hidden" />
          </Button>
        </div>
        <div className={cn(mode === 'single' ? 'hidden' : null)}>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(new Set(), 'remove-all');
              setRightSelected({});
            }}
          >
            <DoubleArrowLeftIcon className="hidden md:block" />
            <DoubleArrowUpIcon className="block md:hidden" />
          </Button>
        </div>
      </div>
      <div className="flex flex-1 border border-gray-300 rounded-lg overflow-hidden">
        <ListBox
          columns={rightCols}
          data={rightData}
          onRowSelectionChange={setRightSelected}
          rowSelection={rightSelected}
          mode={mode}
        />
      </div>
    </div>
  );
}
