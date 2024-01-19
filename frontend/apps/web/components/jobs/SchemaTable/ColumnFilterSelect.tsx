import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/libs/utils';
import { CheckIcon, MagnifyingGlassIcon } from '@radix-ui/react-icons';
import Fuse from 'fuse.js';
import memoizeOne from 'memoize-one';
import { CSSProperties, ReactElement, useCallback, useState } from 'react';
import { AiOutlineFilter } from 'react-icons/ai';
import AutoSizer from 'react-virtualized-auto-sizer';
import { FixedSizeList as List } from 'react-window';

interface Props {
  allColumnFilters: Record<string, string[]>;
  setColumnFilters: (columnId: string, newValues: string[]) => void;
  columnId: string;
  possibleFilters: string[];
}

const createRowData = memoizeOne(
  (
    columnFilters: string[],
    uniqueColFilters: Set<string>,
    setColumnFilters: (columnId: string, newValues: string[]) => void,
    setOpen: (isOpen: boolean) => void,
    columnId: string,
    possibleFilters: string[]
  ) => ({
    columnFilters,
    uniqueColFilters,
    setColumnFilters,
    setOpen,
    columnId,
    possibleFilters,
  })
);

function getFuzzyPossibleFilters(
  possibleFilters: string[],
  fuzzyText: string | undefined
): string[] {
  if (!fuzzyText) {
    return possibleFilters;
  }
  // this seems to be performant, but may need or want to memoize this at some point
  const fuse = new Fuse(possibleFilters, { threshold: 0.3 });
  const fuzziedPossibleFilters = fuse.search(fuzzyText ?? '');
  return fuzziedPossibleFilters.map((pf) => pf.item);
}

export default function ColumnFilterSelect(props: Props) {
  const { allColumnFilters, setColumnFilters, columnId, possibleFilters } =
    props;
  const [open, setOpen] = useState(false);
  const [fuzzyText, setFuzzyText] = useState<string>();
  const filteredPossibleFilters = getFuzzyPossibleFilters(
    possibleFilters,
    fuzzyText
  );
  const columnFilters = allColumnFilters[columnId] ?? [];
  const uniqueColFilters = new Set(columnFilters);
  const itemData = createRowData(
    columnFilters,
    uniqueColFilters,
    setColumnFilters,
    setOpen,
    columnId,
    filteredPossibleFilters
  );
  const itemKey = useCallback(
    (index: number) => {
      return filteredPossibleFilters[index];
    },
    [filteredPossibleFilters]
  );

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          role="combobox"
          aria-expanded={open}
          className="hover:bg-gray-200 dark:hover:bg-gray-600 p-2"
        >
          <AiOutlineFilter />
          {columnFilters && columnFilters.length ? (
            <div
              id="notifbadge"
              className="bg-blue-500 w-[6px] h-[6px] text-white rounded-full text-[8px] relative top-[-8px] right-0"
            />
          ) : null}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="min-w-[175px] p-0">
        <div className="flex items-center border-b border-b-card-border px-3">
          <MagnifyingGlassIcon className="mr-2 h-4 w-4 shrink-0 opacity-50" />
          <Input
            className="h-10 bg-transparent px-0 py-3 outline-none border-none focus-visible:ring-0"
            placeholder="Search filters..."
            value={fuzzyText}
            onChange={(e) => setFuzzyText(e.target.value)}
          />
        </div>
        <div
          className="flex pt-1"
          style={{ height: Math.min(40 * filteredPossibleFilters.length, 300) }}
        >
          <AutoSizer>
            {({ height, width }) => (
              <List
                height={height}
                width={width}
                itemCount={filteredPossibleFilters.length}
                itemSize={35}
                itemData={itemData}
                itemKey={itemKey}
              >
                {VirtualCommandItem}
              </List>
            )}
          </AutoSizer>
        </div>
      </PopoverContent>
    </Popover>
  );
}

interface VirtualCommandItemProps {
  index: number;
  style: CSSProperties;
  data: VirtualCommandItemData;
}
interface VirtualCommandItemData {
  columnFilters: string[];
  uniqueColFilters: Set<string>;
  setOpen(isOpen: boolean): void;
  setColumnFilters(columnId: string, newValues: string[]): void;
  columnId: string;
  possibleFilters: string[];
}

function VirtualCommandItem(props: VirtualCommandItemProps): ReactElement {
  const { index, style, data } = props;
  const {
    columnFilters,
    uniqueColFilters,
    setColumnFilters,
    setOpen,
    columnId,
    possibleFilters,
  } = data;
  const possibleFilter = possibleFilters[index];
  return (
    <div
      className="flex px-4 cursor-default select-none items-center rounded-sm hover:bg-accent"
      style={style}
      key={`${possibleFilter}`}
      onClick={() => {
        // use i here instead of value because it lowercases the value
        const newValues = computeFilters(
          possibleFilter,
          columnFilters,
          uniqueColFilters
        );
        setColumnFilters(columnId, newValues);
        setOpen(false);
      }}
    >
      <CheckIcon
        className={cn(
          'mr-2 h-4 w-4',
          uniqueColFilters.has(possibleFilter) ? 'opacity-100' : 'opacity-0'
        )}
      />
      <span className="truncate tracking-tight">{possibleFilter}</span>
    </div>
  );
}

function computeFilters(
  newValue: string,
  currentValues: string[],
  uniqueCurrentValues: Set<string>
): string[] {
  if (uniqueCurrentValues.has(newValue)) {
    return currentValues.filter((v) => v != newValue);
  }
  return currentValues.concat(newValue);
}
