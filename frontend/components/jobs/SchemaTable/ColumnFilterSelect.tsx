import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/libs/utils';
import { CheckIcon } from '@radix-ui/react-icons';
import { useState } from 'react';
import { AiOutlineFilter } from 'react-icons/ai';

interface Props {
  allColumnFilters: Record<string, string[]>;
  setColumnFilters: (columnId: string, newValues: string[]) => void;
  columnId: string;
  possibleFilters: string[];
}

export default function ColumnFilterSelect(props: Props) {
  const { allColumnFilters, setColumnFilters, columnId, possibleFilters } =
    props;
  const [open, setOpen] = useState(false);

  const columnFilters = allColumnFilters[columnId];

  function computeFilters(newValue: string, currentValues: string[]): string[] {
    if (currentValues.includes(newValue)) {
      return currentValues.filter((v) => v != newValue);
    }
    return [...currentValues, newValue];
  }
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
        <Command>
          <CommandInput placeholder="Search filters..." />
          <CommandEmpty>No filters found.</CommandEmpty>
          <CommandGroup>
            {possibleFilters.map((i, index) => (
              <CommandItem
                key={`${i}-${index}`}
                onSelect={() => {
                  // use i here instead of value because it lowercases the value
                  const newValues = computeFilters(i, columnFilters || []);
                  setColumnFilters(columnId, newValues);
                  setOpen(false);
                }}
                value={i}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    columnFilters && columnFilters.includes(i)
                      ? 'opacity-100'
                      : 'opacity-0'
                  )}
                />
                <span className="truncate">{i}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
