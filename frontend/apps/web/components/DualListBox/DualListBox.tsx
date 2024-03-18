import { Button } from '@/components/ui/button';
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
import { Row, getListBoxColumns } from './columns';

export interface Option {
  value: string;
  // label: string;
}
export type Action = 'add' | 'add-all' | 'remove' | 'remove-all';
interface Props {
  options: Option[];
  selected: Set<string>;
  onChange(value: Set<string>, action: Action): void;
  title: string;
}

export default function DualListBox(props: Props): ReactElement {
  const { options, selected, onChange, title } = props;

  const [leftSelected, setLeftSelected] = useState<RowSelectionState>({});
  const [rightSelected, setRightSelected] = useState<RowSelectionState>({});

  const cols = useMemo(() => getListBoxColumns({ title }), []);
  const leftData = options
    .filter((value) => !selected.has(value.value))
    .map((value): Row => ({ value: value.value }));
  const rightData = options
    .filter((value) => selected.has(value.value))
    .map((value): Row => ({ value: value.value }));

  return (
    <div className="flex gap-3 flex-col md:flex-row">
      <div className="flex">
        <ListBox
          columns={cols}
          data={leftData}
          onRowSelectionChange={setLeftSelected}
          rowSelection={leftSelected}
        />
      </div>
      <div className="flex flex-row md:flex-col justify-center gap-2">
        <div>
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
        <div>
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
      <div className="flex">
        <ListBox
          columns={cols}
          data={rightData}
          onRowSelectionChange={setRightSelected}
          rowSelection={rightSelected}
        />
      </div>
    </div>
  );
}
