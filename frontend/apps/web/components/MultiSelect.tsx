'use client';

import * as React from 'react';

import { Badge } from '@/components/ui/badge';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { cn } from '@/libs/utils';
import { Cross2Icon } from '@radix-ui/react-icons';
import { Command as CommandPrimitive } from 'cmdk';
import { useEffect } from 'react';

export interface Option {
  value: string;
  label: string;
  disable?: boolean;
  fixed?: boolean;
  [key: string]: string | boolean | undefined;
}
interface GroupOption {
  [key: string]: Option[];
}

interface MultipleSelectorProps {
  value?: Option[];
  defaultOptions?: Option[];
  placeholder?: string;
  emptyIndicator?: React.ReactNode;
  onChange?: (options: Option[]) => void;
  hidePlaceholderWhenSelected?: boolean;
  commandProps?: React.ComponentPropsWithoutRef<typeof Command>;
  /** Props of `CommandInput` */
  inputProps?: Omit<
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input>,
    'value' | 'placeholder' | 'disabled'
  >;
}

interface MultipleSelectorRef {
  selectedValue: Option[];
  input: HTMLInputElement;
}

function transToGroupOption(options: Option[], groupBy?: string) {
  if (options.length === 0) {
    return {};
  }
  if (!groupBy) {
    return {
      '': options,
    };
  }

  const groupOption: GroupOption = {};
  options.forEach((option) => {
    const key = (option[groupBy] as string) || '';
    if (!groupOption[key]) {
      groupOption[key] = [];
    }
    groupOption[key].push(option);
  });
  return groupOption;
}

function removePickedOption(groupOption: GroupOption, picked: Option[]) {
  const cloneOption = JSON.parse(JSON.stringify(groupOption)) as GroupOption;

  for (const [key, value] of Object.entries(cloneOption)) {
    cloneOption[key] = value.filter(
      (val) => !picked.find((p) => p.value === val.value)
    );
  }
  return cloneOption;
}

// using the multi-select component from here: https://shadcnui-expansions.typeart.cc/docs/multiple-selector
export const MultiSelect = React.forwardRef<
  MultipleSelectorRef,
  MultipleSelectorProps
>(
  (
    {
      value,
      onChange,
      placeholder,
      defaultOptions: arrayDefaultOptions = [],
      emptyIndicator,
      hidePlaceholderWhenSelected,
      commandProps,
      inputProps,
    }: MultipleSelectorProps,
    ref: React.Ref<MultipleSelectorRef>
  ) => {
    const inputRef = React.useRef<HTMLInputElement>(null);
    const [open, setOpen] = React.useState(false);

    const [selected, setSelected] = React.useState<Option[]>(value || []);
    const [options, _] = React.useState<GroupOption>(
      transToGroupOption(arrayDefaultOptions)
    );
    const [inputValue, setInputValue] = React.useState('');
    React.useImperativeHandle(
      ref,
      () => ({
        selectedValue: [...selected],
        input: inputRef.current as HTMLInputElement,
      }),
      [selected]
    );

    const handleUnselect = React.useCallback(
      (option: Option) => {
        const newOptions = selected.filter((s) => s.value !== option.value);
        setSelected(newOptions);
        onChange?.(newOptions);
      },
      [selected]
    );

    const handleKeyDown = React.useCallback(
      (e: React.KeyboardEvent<HTMLDivElement>) => {
        const input = inputRef.current;
        if (input) {
          if (e.key === 'Delete' || e.key === 'Backspace') {
            if (input.value === '' && selected.length > 0) {
              handleUnselect(selected[selected.length - 1]);
            }
          }
          // This is not a default behaviour of the <input /> field
          if (e.key === 'Escape') {
            input.blur();
          }
        }
      },
      [selected]
    );

    useEffect(() => {
      if (value) {
        setSelected(value);
      }
    }, [value]);

    const selectables = React.useMemo<GroupOption>(
      () => removePickedOption(options, selected),
      [options, selected]
    );

    const EmptyItem = React.useCallback(() => {
      if (!emptyIndicator) return undefined;

      // For async search that showing emptyIndicator
      if (Object.keys(options).length === 0) {
        return (
          <CommandItem value="-" disabled>
            {emptyIndicator}
          </CommandItem>
        );
      }

      return <CommandEmpty>{emptyIndicator}</CommandEmpty>;
    }, [emptyIndicator, options]);

    return (
      <Command
        {...commandProps}
        onKeyDown={(e) => {
          handleKeyDown(e);
          commandProps?.onKeyDown?.(e);
        }}
        className={cn(
          'overflow-visible bg-transparent',
          commandProps?.className
        )}
      >
        <div className="group rounded-md border border-input px-3 py-1  text-sm focus-within:ring-1 focus-within:ring-gray-400">
          <div className="flex flex-wrap gap-1">
            {selected.map((option) => {
              return (
                <Badge
                  key={option.value}
                  className={cn(
                    'data-[disabled]:bg-muted-foreground data-[disabled]:text-muted data-[disabled]:hover:bg-muted-foreground',
                    'data-[fixed]:bg-muted-foreground data-[fixed]:text-muted data-[fixed]:hover:bg-muted-foreground'
                  )}
                  variant="outline"
                  data-fixed={option.fixed}
                >
                  {option.label}
                  <button
                    className={cn(
                      'ml-1 rounded-full outline-none ring-offset-background ',
                      option.fixed && 'hidden'
                    )}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        handleUnselect(option);
                      }
                    }}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      e.stopPropagation();
                    }}
                    onClick={() => handleUnselect(option)}
                  >
                    <Cross2Icon className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                  </button>
                </Badge>
              );
            })}
            {/* Avoid having the "Search" Icon */}
            <CommandPrimitive.Input
              {...inputProps}
              ref={inputRef}
              value={inputValue}
              onValueChange={(value) => {
                setInputValue(value);
              }}
              onBlur={() => {
                setOpen(false);
              }}
              onFocus={(event) => {
                setOpen(true);
                inputProps?.onFocus?.(event);
              }}
              placeholder={
                hidePlaceholderWhenSelected && selected.length !== 0
                  ? ''
                  : placeholder
              }
              className={cn(
                'ml-2 flex-1 bg-transparent outline-none placeholder:text-muted-foreground h-7',
                inputProps?.className
              )}
            />
          </div>
        </div>
        <div className="relative mt-2">
          {open && (
            <CommandList className="absolute top-0 z-50 w-full rounded-md border bg-popover text-popover-foreground shadow-md outline-none animate-in">
              <>
                {EmptyItem()}
                {Object.entries(selectables).map(([key, dropdowns]) => (
                  <CommandGroup
                    key={key}
                    heading={key}
                    className="h-full overflow-auto"
                  >
                    <>
                      {dropdowns.map((option) => {
                        return (
                          <CommandItem
                            key={option.value}
                            value={option.value}
                            disabled={option.disable}
                            onMouseDown={(e) => {
                              e.preventDefault();
                              e.stopPropagation();
                            }}
                            onSelect={() => {
                              setInputValue('');
                              const newOptions = [...selected, option];
                              setSelected(newOptions);
                              onChange?.(newOptions);
                            }}
                            className={cn(
                              'cursor-pointer',
                              option.disable &&
                                'cursor-default text-muted-foreground'
                            )}
                          >
                            {option.label}
                          </CommandItem>
                        );
                      })}
                    </>
                  </CommandGroup>
                ))}
              </>
            </CommandList>
          )}
        </div>
      </Command>
    );
  }
);

MultiSelect.displayName = 'MultipleSelector';
