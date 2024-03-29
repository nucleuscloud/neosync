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
  JobMappingTransformerForm,
  convertJobMappingTransformerToForm,
} from '@/yup-validations/jobs';
import {
  JobMappingTransformer,
  SystemTransformer,
  TransformerConfig,
  TransformerDataType,
  TransformerSource,
  UserDefinedTransformer,
  UserDefinedTransformerConfig,
} from '@neosync/sdk';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

type Side = (typeof SIDE_OPTIONS)[number];

var SIDE_OPTIONS: readonly ['top', 'right', 'bottom', 'left'];

interface Props {
  userDefinedTransformers: UserDefinedTransformer[];
  systemTransformers: SystemTransformer[];

  userDefinedTransformerMap: Map<string, UserDefinedTransformer>;
  systemTransformerMap: Map<TransformerSource, SystemTransformer>;
  value: JobMappingTransformerForm;
  onSelect(value: JobMappingTransformerForm): void;
  placeholder: string;
  side: Side;
  disabled: boolean;
}

export default function TransformerSelect(props: Props): ReactElement {
  const {
    userDefinedTransformers,
    systemTransformers,
    userDefinedTransformerMap,
    systemTransformerMap,
    value,
    onSelect,
    placeholder,
    side,
    disabled,
  } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            placeholder.startsWith('Bulk')
              ? 'justify-between w-[275px]'
              : 'justify-between w-[175px]'
          )}
        >
          <div className="whitespace-nowrap truncate lg:w-[200px] text-left">
            {getPopoverTriggerButtonText(
              value,
              userDefinedTransformerMap,
              systemTransformerMap,
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
            {userDefinedTransformers.length > 0 && (
              <CommandGroup heading="Custom">
                {userDefinedTransformers.map((t) => {
                  const dataType = TransformerDataType[t.dataType];
                  return (
                    <CommandItem
                      key={t.id}
                      onSelect={() => {
                        onSelect(
                          convertJobMappingTransformerToForm(
                            new JobMappingTransformer({
                              source: TransformerSource.USER_DEFINED,
                              config: new TransformerConfig({
                                config: {
                                  case: 'userDefinedTransformerConfig',
                                  value: new UserDefinedTransformerConfig({
                                    id: t.id,
                                  }),
                                },
                              }),
                            })
                          )
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
                              value?.config?.case ===
                                'userDefinedTransformerConfig' &&
                                value?.source ===
                                  TransformerSource.USER_DEFINED &&
                                value.config.value.id === t.id
                                ? 'opacity-100'
                                : 'opacity-0'
                            )}
                          />
                          <div className="items-center">{t?.name}</div>
                        </div>
                        <div className="ml-2 text-gray-400 text-xs">
                          {dataType ? dataType.toLowerCase() : 'unknown'}
                        </div>
                      </div>
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            )}
            <CommandGroup heading="System">
              {systemTransformers.map((t) => {
                const dataType = TransformerDataType[t.dataType];
                return (
                  <CommandItem
                    key={t.source}
                    onSelect={() => {
                      onSelect(
                        convertJobMappingTransformerToForm(
                          new JobMappingTransformer({
                            source: t.source,
                            config: t.config,
                          })
                        )
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
                            value?.source === t?.source
                              ? 'opacity-100'
                              : 'opacity-0'
                          )}
                        />
                        <div className="items-center">{t?.name}</div>
                      </div>
                      <div className="ml-2 text-gray-400 text-xs">
                        {dataType ? dataType.toLowerCase() : 'unknown'}
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
  systemTransformerMap: Map<TransformerSource, SystemTransformer>,
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
