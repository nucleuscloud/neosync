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
import {
  CustomTransformerConfig,
  Transformer,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { toTitleCase } from '@/util/util';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { TransformerWithType } from './schema-table';

interface Props {
  transformers: TransformerWithType[];
  value: Transformer;
  onSelect: (value: Transformer) => void;
  placeholder: string;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn(
            placeholder.startsWith('Bulk')
              ? 'justify-between w-[275px]'
              : 'justify-between w-[175px]'
          )}
        >
          <div className="whitespace-nowrap truncate w-[175px]">
            {toTitleCase(value.name)
              ? toTitleCase(value.name)
              : placeholder
                ? placeholder
                : 'Select a transformer'}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[300px] p-0"
        side="left"
        avoidCollisions={false}
      >
        <Command>
          <CommandInput placeholder={placeholder} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <div className="max-h-[400px] overflow-y-scroll">
            <CommandGroup heading="Custom">
              {transformers
                .filter((item) => item.transformerType == 'custom')
                .map((t, index) => (
                  <CommandItem
                    key={`${t?.name}-${index}`}
                    onSelect={(currentValue) => {
                      const selectedTransformer = FindTransformerByName(
                        currentValue,
                        transformers
                      );
                      const customTransformer = new Transformer({
                        ...selectedTransformer,
                        config: new TransformerConfig({
                          config: {
                            case: 'customTransformerConfig',
                            value: new CustomTransformerConfig({
                              id: selectedTransformer?.id,
                            }),
                          },
                        }),
                      });
                      onSelect(customTransformer);
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center">
                      <CheckIcon
                        className={cn(
                          'mr-2 h-4 w-4',
                          value.name == t?.name ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      {t?.name}
                      <div className="ml-2 text-gray-400 text-xs">
                        {t.dataType}
                      </div>
                    </div>
                  </CommandItem>
                ))}
            </CommandGroup>
            <CommandGroup heading="System">
              {transformers
                .filter((item) => item.transformerType == 'system')
                .map((t, index) => (
                  <CommandItem
                    key={`${t?.name}-${index}`}
                    onSelect={(currentValue) => {
                      const selectedTransformer = FindTransformerByName(
                        currentValue,
                        transformers
                      );
                      const systemTransformer = new Transformer({
                        name: selectedTransformer.name,
                        source: selectedTransformer.source,
                        config: selectedTransformer.config,
                      });
                      onSelect(systemTransformer);
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center">
                      <CheckIcon
                        className={cn(
                          'mr-2 h-4 w-4',
                          value.name == t?.name ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      {t?.name}
                      <div className="ml-2 text-gray-400 text-xs">
                        {t.dataType}
                      </div>
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

function FindTransformerByName(
  name: string,
  transformers: Transformer[]
): Transformer {
  return (
    transformers?.find((item) => item.name.toLowerCase() == name) ??
    new Transformer()
  );
}
