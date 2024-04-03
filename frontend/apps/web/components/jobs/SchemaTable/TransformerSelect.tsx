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
  TransformerSource,
  UserDefinedTransformer,
  UserDefinedTransformerConfig,
} from '@neosync/sdk';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';

type Side = (typeof SIDE_OPTIONS)[number];

var SIDE_OPTIONS: readonly ['top', 'right', 'bottom', 'left'];

interface Props {
  getTransformers(): {
    system: SystemTransformer[];
    userDefined: UserDefinedTransformer[];
  };
  value: JobMappingTransformerForm;
  buttonText: string;
  buttonClassName?: string;
  onSelect(value: JobMappingTransformerForm): void;
  side: Side;
  disabled: boolean;
}

export default function TransformerSelect(props: Props): ReactElement {
  const {
    getTransformers,
    value,
    onSelect,
    buttonText,
    side,
    disabled,
    buttonClassName,
  } = props;
  const [open, setOpen] = useState(false);

  const { system, userDefined } = open
    ? getTransformers()
    : { system: [], userDefined: [] };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          type="button"
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn('justify-between', buttonClassName)}
        >
          <div className="whitespace-nowrap truncate lg:w-[200px] text-left">
            {buttonText}
          </div>
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className="w-[350px] p-0"
        avoidCollisions={true} // this prevents the popover from overflowing out of the viewport (meaning there is content the user isn't able to access)
        side={side}
      >
        <Command>
          <CommandInput placeholder={buttonText} />
          <CommandEmpty>No transformers found.</CommandEmpty>
          <div className="max-h-[450px] overflow-y-scroll">
            {userDefined.length > 0 && (
              <CommandGroup heading="Custom">
                {userDefined.map((t) => {
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
                      </div>
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            )}
            <CommandGroup heading="System">
              {system.map((t) => {
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
