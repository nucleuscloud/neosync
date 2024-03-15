import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  DoubleArrowLeftIcon,
  DoubleArrowRightIcon,
} from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Option {
  value: string;
  label: string;
}
export type Action = 'add' | 'add-all' | 'remove' | 'remove-all';
interface Props {
  options: Option[];
  selected: Set<string>;
  onChange(value: Set<string>, action: Action): void;
}

export default function DualListBox(props: Props): ReactElement {
  const { options, selected, onChange } = props;
  const [leftStaged, setLeftStaged] = useState<Set<string>>(new Set());
  const [rightStaged, setRightStaged] = useState<Set<string>>(new Set());
  return (
    <div className="flex w-full gap-3">
      <div className="flex flex-1">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Select</TableHead>
              <TableHead>Table</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {options
              .filter((value) => !selected.has(value.value))
              .map((value) => {
                return (
                  <TableRow key={value.value}>
                    <TableCell>
                      <Checkbox
                        id={value.value}
                        checked={leftStaged.has(value.value)}
                        onCheckedChange={(state) => {
                          const newSet = new Set(leftStaged);
                          if (newSet.has(value.value)) {
                            newSet.delete(value.value);
                          } else {
                            newSet.add(value.value);
                          }
                          setLeftStaged(newSet);
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      <label htmlFor={value.value}>{value.value}</label>
                    </TableCell>
                  </TableRow>
                );
              })}
          </TableBody>
        </Table>
      </div>
      {/* <select
        multiple
        className="flex flex-1 w-full border-black-30 focus:ring-0 focus:border-none select:focus-none"
      >

      </select> */}
      <div className="flex flex-col gap-2">
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(
                new Set(options.map((option) => option.value)),
                'add-all'
              );
              setLeftStaged(new Set());
            }}
          >
            <DoubleArrowRightIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              const newSet = new Set(selected);
              leftStaged.forEach((value) => newSet.add(value));
              onChange(newSet, 'add');
              setLeftStaged(new Set());
            }}
          >
            <ArrowRightIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              const newSet = new Set(selected);
              rightStaged.forEach((value) => newSet.delete(value));
              onChange(newSet, 'remove');
              setRightStaged(new Set());
            }}
          >
            <ArrowLeftIcon />
          </Button>
        </div>
        <div>
          <Button
            type="button"
            variant="ghost"
            onClick={() => {
              onChange(new Set(), 'remove-all');
              setRightStaged(new Set());
            }}
          >
            <DoubleArrowLeftIcon />
          </Button>
        </div>
      </div>
      <div className="flex flex-1">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Select</TableHead>
              <TableHead>Table</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {options
              .filter((value) => selected.has(value.value))
              .map((value) => {
                return (
                  <TableRow key={value.value}>
                    <TableCell>
                      <Checkbox
                        id={value.value}
                        checked={rightStaged.has(value.value)}
                        onCheckedChange={(state) => {
                          const newSet = new Set(rightStaged);
                          if (newSet.has(value.value)) {
                            newSet.delete(value.value);
                          } else {
                            newSet.add(value.value);
                          }
                          setRightStaged(newSet);
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      <label htmlFor={value.value}>{value.value}</label>
                    </TableCell>
                  </TableRow>
                );
              })}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
