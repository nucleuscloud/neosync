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
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  transformers: Transformer[];
  value: string;
  onSelect: (value: string) => void;
  placeholder: string;
  defaultValue?: string;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder, defaultValue } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="w-full justify-between"
        >
          <div className="truncate overflow-hidden text-ellipsis whitespace-nowrap">
            {value
              ? transformers.find((t) => t.value === value)?.value
              : placeholder}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[250px] p-0">
        <Command>
          <CommandInput placeholder={placeholder} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <CommandGroup>
            {transformers.map((t, index) => (
              <CommandItem
                key={`${t.value}-${index}`}
                onSelect={(currentValue) => {
                  onSelect(currentValue);
                  setOpen(false);
                }}
                value={t.value}
                defaultValue={defaultValue}
              >
                <CheckIcon
                  className={cn(
                    'mr-2 h-4 w-4',
                    value == t.value ? 'opacity-100' : 'opacity-0'
                  )}
                />
                {t.value}
              </CommandItem>
            ))}
          </CommandGroup>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
