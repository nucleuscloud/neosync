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
import { CustomTransformer } from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  transformers: CustomTransformer[];
  value: string;
  onSelect: (value: string) => void;
  placeholder: string;
  defaultValue?: string;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder, defaultValue } = props;
  const [open, setOpen] = useState(false);

  console.log('transformer', transformers);

  let custom: CustomTransformer[] = [];
  let system: CustomTransformer[] = [];

  transformers.forEach((t) => {
    if (t.id) {
      custom.push(t);
    } else {
      system.push(t);
    }
  });

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between w-[160px]"
        >
          <div className="whitespace-nowrap truncate">{value}</div>

          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[250px] p-0"
        side="left"
        avoidCollisions={false}
      >
        <Command>
          <CommandInput placeholder={placeholder} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <div className="max-h-[400px] overflow-y-scroll">
            <CommandGroup heading="Custom">
              {custom.map((t, index) => (
                <CommandItem
                  key={`${t?.name}-${index}`}
                  onSelect={(currentValue) => {
                    onSelect(currentValue);
                    setOpen(false);
                  }}
                  value={t?.name}
                  defaultValue={defaultValue}
                >
                  <div className="flex flex-row items-center">
                    <CheckIcon
                      className={cn(
                        'mr-2 h-4 w-4',
                        value == t?.name ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                    {t?.name}
                    <div className="ml-2 text-gray-400 text-xs">{t.type}</div>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
            <CommandGroup heading="System">
              {system.map((t, index) => (
                <CommandItem
                  key={`${t?.name}-${index}`}
                  onSelect={(currentValue) => {
                    onSelect(currentValue);
                    setOpen(false);
                  }}
                  value={t?.name}
                  defaultValue={defaultValue}
                >
                  <div className="flex flex-row items-center">
                    <CheckIcon
                      className={cn(
                        'mr-2 h-4 w-4',
                        value == t?.name ? 'opacity-100' : 'opacity-0'
                      )}
                    />
                    {t?.name}
                    <div className="ml-2 text-gray-400 text-xs">{t.type}</div>
                  </div>
                </CommandItem>
              ))}
            </CommandGroup>
          </div>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
