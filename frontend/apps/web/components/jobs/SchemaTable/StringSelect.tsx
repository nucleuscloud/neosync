import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Command,
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
import { CheckIcon, ChevronDownIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  value?: string;
  values: string[];
  badgeValueMap?: Record<string, string>;
  setValue(v: string): void;
  text: string;
}

export function StringSelect(props: Props): ReactElement {
  const { value, values, setValue, text, badgeValueMap } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className=" justify-between "
        >
          {value ? values.find((v) => v === value) : `Select ${text}...`}
          <ChevronDownIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="p-0 ml-20">
        <Command>
          <CommandInput placeholder={`Search ${text}...`} />
          <CommandGroup>
            {values.length > 0 ? (
              <>
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
                    <p className="mr-2">{v}</p>
                    {badgeValueMap && (
                      <Badge variant="secondary">{badgeValueMap[v]}</Badge>
                    )}
                  </CommandItem>
                ))}{' '}
              </>
            ) : (
              <div className="text-sm text-gray-500 p-4">No {text} found.</div>
            )}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
