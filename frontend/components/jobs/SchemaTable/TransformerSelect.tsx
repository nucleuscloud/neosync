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
import { TransformerFormValues } from '@/yup-validations/jobs';
import {
  JobMappingTransformer,
  SystemTransformer,
  TransformerConfig,
  UserDefinedTransformer,
  UserDefinedTransformerConfig,
} from '@neosync/sdk';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

interface Props {
  transformers: Transformer[];
  value: JobMappingTransformer | TransformerFormValues;
  onSelect(value: JobMappingTransformer): void;
  placeholder: string;
}

export default function TransformerSelect(props: Props): ReactElement {
  const { transformers, value, onSelect, placeholder } = props;
  const [open, setOpen] = useState(false);

  const udfTransformers = transformers.filter(isUserDefinedTransformer);
  const sysTransformers = transformers.filter(isSystemTransformer);

  const udfTransformerMap = new Map(udfTransformers.map((t) => [t.id, t]));
  const sysTransformerMap = new Map(sysTransformers.map((t) => [t.source, t]));
  // Because of how the react-hook-form data is persisted, sometimes the initial data comes in untransformed
  // and is temporarily the uninstanced value on initial render. But subsequent renders it's correct.
  // This just combats that and ensures it's the right value
  const jmValue =
    value instanceof JobMappingTransformer
      ? value
      : JobMappingTransformer.fromJson(value);

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
            {getPopoverTriggerButtonText(
              jmValue,
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
        side="left"
        avoidCollisions={false}
      >
        <Command>
          <CommandInput placeholder={placeholder} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <div className="max-h-[400px] overflow-y-scroll">
            <CommandGroup heading="Custom">
              {udfTransformers.map((t) => {
                return (
                  <CommandItem
                    key={t.id}
                    onSelect={() => {
                      onSelect(
                        new JobMappingTransformer({
                          source: 'custom',
                          config: new TransformerConfig({
                            config: {
                              case: 'userDefinedTransformerConfig',
                              value: new UserDefinedTransformerConfig({
                                id: t.id,
                              }),
                            },
                          }),
                        })
                      );
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center justify-between w-full">
                      <div className="flex flex-row items-center">
                        <CheckIcon
                          className={cn(
                            'mr-2 h-4 w-4',
                            jmValue.config?.config.case ===
                              'userDefinedTransformerConfig' &&
                              jmValue?.source === 'custom' &&
                              jmValue.config.config.value.id === t.id
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
                      onSelect(
                        new JobMappingTransformer({
                          source: t.source,
                          config: t.config,
                        })
                      );
                      setOpen(false);
                    }}
                    value={t.name}
                  >
                    <div className="flex flex-row items-center justify-between w-full">
                      <div className=" flex flex-row items-center">
                        <CheckIcon
                          className={cn(
                            'mr-2 h-4 w-4',
                            value?.source == t?.source
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
  value: JobMappingTransformer,
  udfTransformerMap: Map<string, UserDefinedTransformer>,
  systemTransformerMap: Map<string, SystemTransformer>,
  placeholder: string
): string {
  if (!value.config) {
    return placeholder;
  }

  switch (value.config?.config?.case) {
    case 'userDefinedTransformerConfig':
      const id = value.config.config.value.id;
      return udfTransformerMap.get(id)?.name ?? placeholder;
    default:
      return systemTransformerMap.get(value.source)?.name ?? placeholder;
  }
}
