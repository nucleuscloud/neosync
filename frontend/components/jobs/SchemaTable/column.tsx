'use client';

import EditTransformerOptions from '@/app/transformers/EditTransformerOptions';
import { Checkbox } from '@/components/ui/checkbox';

import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
import {
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from '@/components/ui/form';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';

import { Switch } from '@/components/ui/switch';
import { cn } from '@/libs/utils';
import { DatabaseColumn } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Transformer } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';
import { PlainMessage } from '@bufbuild/protobuf';
import { CaretSortIcon, CheckIcon } from '@radix-ui/react-icons';
import { ColumnDef } from '@tanstack/react-table';
import { useState } from 'react';
import { DataTableColumnHeader } from './data-table-column-header';

interface GetColumnsProps {
  transformers?: Transformer[];
}

export function getColumns(
  props: GetColumnsProps
): ColumnDef<PlainMessage<DatabaseColumn>>[] {
  const { transformers } = props;

  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
          className="translate-y-[2px]"
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label="Select row"
          className="translate-y-[2px] "
        />
      ),
      enableSorting: false,
      enableHiding: false,
      enableColumnFilter: false,
    },
    {
      accessorKey: 'schema',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Schema" />
      ),
      cell: ({ row }) => <div>{row.getValue('schema')}</div>,
      enableSorting: true,
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'table',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Table" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('table')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'column',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Column" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <span className="truncate font-medium">
              {row.getValue('column')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'dataType',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Data Type" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2 ">
            <span className="truncate font-medium">
              {row.getValue('dataType')}
            </span>
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'transformer',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Transformer" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex space-x-2">
            <FormField
              name={`mappings.${row.index}.transformer.value`}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <div className="flex flex-row space-x-2">
                      <TansformerSelect
                        transformers={transformers || []}
                        value={field.value}
                        onSelect={field.onChange}
                      />
                      <EditTransformerOptions
                        transformer={transformers?.find(
                          (item) => item.value == field.value
                        )}
                        index={row.index}
                      />
                    </div>
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: 'arrIncludesSome',
    },
    {
      accessorKey: 'exclude',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title="Include" />
      ),
      cell: ({ row }) => {
        return (
          <div className="flex flex-row space-x-2">
            <FormField
              name={`mappings.${row.index}.exclude`}
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <Switch
                      checked={!field.value}
                      onCheckedChange={(checked: boolean) => {
                        field.onChange(!checked);
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        );
      },
      enableColumnFilter: true,
      filterFn: (row, id, value) => {
        const filter = row.getValue(id) ? 'exclude' : 'include';
        return filter.includes(value);
      },
    },
  ];
}

interface TransformersSelectProps {
  transformers: Transformer[];
  value: string;
  onSelect: (value: string) => void;
}

function TansformerSelect(props: TransformersSelectProps) {
  const { transformers, value, onSelect } = props;
  const [open, setOpen] = useState(false);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className="justify-between " //whitespace-nowrap
        >
          {value
            ? transformers.find((t) => t.value === value)?.value
            : 'Transformer'}
          <CaretSortIcon className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className=" p-0">
        <Command>
          <CommandInput placeholder="Search transformers..." />
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
                defaultValue={'passthrough'}
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
