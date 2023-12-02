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
import { JobMappingTransformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import {
  TransformerConfig,
  UserDefinedTransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { TransformerWithType } from './schema-table';

interface Props {
  transformers: TransformerWithType[];
  value: TransformerWithType;
  onSelect: (value: JobMappingTransformer) => void;
  placeholder: string;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder } = props;
  const [open, setOpen] = useState(false);
  useState<TransformerWithType>();
  const [transformerName, setTransformerName] = useState<string>('');

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
            {transformerName
              ? transformerName
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
                      setTransformerName(selectedTransformer?.name ?? '');
                      const jobMappingTransformer = new JobMappingTransformer({
                        source:
                          FindTransformerByName(currentValue, transformers)
                            ?.source ?? '',
                        config: new TransformerConfig({
                          config: {
                            case: 'userDefinedTransformerConfig',
                            value: new UserDefinedTransformerConfig({
                              id: selectedTransformer?.id,
                            }),
                          },
                        }),
                      });
                      onSelect(jobMappingTransformer);
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
                      setTransformerName(selectedTransformer?.name ?? '');
                      const jobMappingTransformer = new JobMappingTransformer({
                        source:
                          FindTransformerByName(currentValue, transformers)
                            ?.source ?? '',
                        config: new TransformerConfig({
                          config: selectedTransformer?.config?.config,
                        }),
                      });
                      onSelect(jobMappingTransformer);
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
  transformers: TransformerWithType[]
): TransformerWithType | undefined {
  const t = transformers?.find((item) => item.name.toLowerCase() == name)!;

  return t;
}
