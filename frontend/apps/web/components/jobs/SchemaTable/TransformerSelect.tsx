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
  Transformer,
  isSystemTransformer,
  isUserDefinedTransformer,
} from '@/shared/transformers';
import { JobMappingTransformerForm } from '@/yup-validations/jobs';
import {
  SystemTransformer,
  UserDefinedTransformer,
  UserDefinedTransformerConfig,
} from '@neosync/sdk';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

type Side = (typeof SIDE_OPTIONS)[number];

var SIDE_OPTIONS: readonly ['top', 'right', 'bottom', 'left'];

interface Props {
  transformers: Transformer[];
  value: JobMappingTransformerForm;
  onSelect(value: JobMappingTransformerForm): void;
  placeholder: string;
  side: Side;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder, side } = props;
  const [open, setOpen] = useState(false);

  const udfTransformers = transformers
    .filter(isUserDefinedTransformer)
    .sort((a, b) => a.name.localeCompare(b.name));
  const sysTransformers = transformers
    .filter(isSystemTransformer)
    .sort((a, b) => a.name.localeCompare(b.name));

  const udfTransformerMap = new Map(udfTransformers.map((t) => [t.id, t]));
  const sysTransformerMap = new Map(sysTransformers.map((t) => [t.source, t]));

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
          <div className="whitespace-nowrap truncate lg:w-[200px]">
            {getPopoverTriggerButtonText(
              value,
              udfTransformerMap,
              sysTransformerMap,
              placeholder
            )}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[350px] p-0"
        avoidCollisions={false}
        side={side}
      >
        <Command>
          <CommandInput placeholder={placeholder} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <div className="max-h-[450px] overflow-y-scroll">
            <CommandGroup heading="Custom">
              {udfTransformers.map((t) => {
                return (
                  <CommandItem
                    key={t.id}
                    onSelect={() => {
                      onSelect({
                        source: 'custom',
                        config: {
                          case: 'userDefinedTransformerConfig',
                          value: new UserDefinedTransformerConfig({
                            id: t.id,
                          }),
                        },
                      });
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center justify-between w-full">
                      <div className="flex flex-row items-center">
                        <CheckIcon
                          className={cn(
                            'mr-2 h-4 w-4',
                            value?.config?.case ===
                              'userDefinedTransformerConfig' &&
                              value?.source === 'custom' &&
                              value.config.value.id === t.id
                              ? 'opacity-100'
                              : 'opacity-0'
                          )}
                        />
                        <div className="items-center">{t?.name}</div>
                      </div>
                      <div className="ml-2 text-gray-400 text-xs">
                        {t.dataType}
                      </div>
                    </div>
                  </CommandItem>
                );
              })}
            </CommandGroup>
            <CommandGroup heading="System">
              {sysTransformers.map((t) => {
                return (
                  <CommandItem
                    key={t.source}
                    onSelect={() => {
                      let config = t.config?.config ?? {
                        case: '',
                        value: {},
                      };
                      if (!config.case) {
                        config = { case: '', value: {} };
                      }
                      onSelect({
                        source: t.source,
                        config: config,
                      });
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center justify-between w-full">
                      <div className=" flex flex-row items-center">
                        <CheckIcon
                          className={cn(
                            'mr-2 h-4 w-4',
                            value?.source === t?.source
                              ? 'opacity-100'
                              : 'opacity-0'
                          )}
                        />
                        <div className="items-center">{t?.name}</div>
                      </div>
                      <div className="ml-2 text-gray-400 text-xs">
                        {t.dataType}
                      </div>
                    </div>
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </div>
        </Command>
      </PopoverContent>
    </Popover>
  );
}

function getPopoverTriggerButtonText(
  value: JobMappingTransformerForm,
  udfTransformerMap: Map<string, UserDefinedTransformer>,
  systemTransformerMap: Map<string, SystemTransformer>,
  placeholder: string
): string {
  if (!value?.config) {
    return placeholder;
  }

  switch (value?.config?.case) {
    case 'userDefinedTransformerConfig':
      const id = value.config.value.id;
      return udfTransformerMap.get(id)?.name ?? placeholder;
    default:
      return systemTransformerMap.get(value.source)?.name ?? placeholder;
  }
}
