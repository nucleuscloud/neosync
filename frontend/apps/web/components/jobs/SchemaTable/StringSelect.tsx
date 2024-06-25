import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { cn } from '@/libs/utils';
import { CheckIcon, ChevronDownIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  value?: string;
  values: string[];
  setValue(v: string): void;
  text: string;
}

export function StringSelect(props: Props): ReactElement {
  const { value, values, setValue, text } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-[200px] justify-between"
        >
          {value ? values.find((v) => v === value) : `Select ${text}...`}
          <ChevronDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[200px] p-0">
        <Command>
          <CommandInput placeholder={`Search ${text}...`} />
          <CommandList>
            <CommandEmpty>No tables found.</CommandEmpty>
            <CommandGroup>
              {values.map((v) => (
                <CommandItem
                  key={v}
                  value={v}
                  onSelect={(currentValue) => {
                    setValue(currentValue === value ? '' : currentValue);
                    setOpen(false);
                  }}
                >
                  <CheckIcon
                    className={cn(
                      'mr-2 h-4 w-4',
                      value === v ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  {v}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
